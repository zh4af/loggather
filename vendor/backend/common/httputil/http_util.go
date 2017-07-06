package httputil

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"

	"backend/common/clog"
	"backend/common/errcode"
	"backend/common/structutil"

	"third/gin"
	"third/http_client_cluster"
)

type CodoonApiResponse struct {
	Status string      `json:"status"`
	Data   interface{} `json:"data"`
	Desc   string      `json:"desc"`
}

// 移步至compare.go
//从header中取出版本号a_version, 与待比较版本b_version比较
func GetVersionCompare(c *gin.Context, b_version string) (int, error) {
	/*
		add by wuql 2016-7-23
		直接使用字符串比较
		从header中取出版本号a_version, 与待比较版本b_version比较
		0 : a_version 	== 	b_version
		1 : a_version 	> 	b_version
		-1: a_version 	< 	b_version
	*/

	header := c.Request.Header
	// 获取子版本号中最长的一个如 [1, 239, 45]
	// 返回3
	getMaxLen := func(al, bl []string) int {
		maxlen := 0
		for _, i := range al {
			if len(i) > maxlen {
				maxlen = len(i)
			}
		}
		for _, i := range bl {
			if len(i) > maxlen {
				maxlen = len(i)
			}
		}
		return maxlen
	}
	//字符串填充
	zFill := func(a string, length int) string {
		if len(a) < length {
			r := ""
			for i := 0; i < length-len(a); i++ {
				r += "0"
			}
			return r + a
		}
		return a
	}
	// why User-Agent
	if user_agent := header.Get("User-Agent"); user_agent != "" {
		if version, ok := FormatUserAgent(user_agent)["version"]; ok {
			if a_version, ok := version.(string); ok {
				al := strings.Split(a_version, ".")
				bl := strings.Split(b_version, ".")
				maxlen := getMaxLen(al, bl)
				ac, bc := "", ""
				for _, i := range al {
					ac = ac + zFill(i, maxlen) + "."
				}
				for _, i := range bl {
					bc = bc + zFill(i, maxlen) + "."
				}
				return strings.Compare(ac, bc), nil
			} else {
				return 0, errors.New("获取版本失败1")
			}
		} else {
			return 0, errors.New("获取版本失败2")
		}
	} else {
		return 0, errors.New("获取版本失败3")
	}
	return 0, nil
}

var sliceOfInts = reflect.TypeOf([]int(nil))
var sliceOfStrings = reflect.TypeOf([]string(nil))

//解析user_agent
func FormatUserAgent(user_agent string) map[string]interface{} {
	dealed_user_agent := strings.TrimSpace(strings.ToLower(user_agent))
	result := map[string]interface{}{
		"version":          "0.0.0",
		"iner_version":     0,
		"platform":         0,
		"platfrom_version": "",
		"device_type":      "",
	}

	if !strings.Contains(dealed_user_agent, "codoonsport(") {
		return result
	}

	var platform = 1
	if strings.Contains(dealed_user_agent, "ios") {
		platform = 0
	}

	dealed_user_agent = strings.Replace(dealed_user_agent, "codoonsport(", "", -1)
	array_user_agent := strings.Split(dealed_user_agent, ")")
	if array_user_agent == nil {
		return result
	}

	ver_list := strings.Split(array_user_agent[0], ";")
	if len(ver_list) != 3 {
		// 长度有误则为非法agent
		return result
	}
	app_version := strings.Split(ver_list[0], " ")
	platfrom_version := strings.Split(ver_list[1], " ")
	result["version"] = app_version[0]
	result["iner_version"] = app_version[1]
	result["platform"] = platform
	result["platform_version"] = platfrom_version[1]
	result["device_type"] = ver_list[2]
	return result
}

type CodoonUserAgent struct {
	Version         string      `json:"version"`
	InerVersion     interface{} `json:"iner_version"`
	Platform        int         `json:"platform"`
	PlatformVersion string      `json:"platform_version"`
	DeviceType      string      `json:"device_type"`
}

func (this *CodoonUserAgent) Format(user_agent string) {
	var (
		result map[string]interface{}
		b      []byte
	)
	result = FormatUserAgent(user_agent)
	b, _ = json.Marshal(&result)
	json.Unmarshal(b, this)
}

//strcmp comparable
//"1.1.1"->0001.0001.000001
func (this *CodoonUserAgent) MakeComparableVersion() (string, error) {
	var (
		vers     []string
		major    int64
		minor    int64
		revision int64
		err      error
		result   string
	)
	vers = strings.Split(this.Version, ".")
	if len(vers) < 3 {
		return "", fmt.Errorf("invalid user agent")
	}
	major, err = strconv.ParseInt(vers[0], 10, 32)
	if err != nil {
		return "", fmt.Errorf("invalid user agent")
	}
	minor, err = strconv.ParseInt(vers[1], 10, 32)
	if err != nil {
		return "", fmt.Errorf("invalid user agent")
	}
	revision, err = strconv.ParseInt(vers[2], 10, 32)
	if err != nil {
		return "", fmt.Errorf("invalid user agent")
	}
	result = fmt.Sprintf("%04d.%04d.%06d", major, minor, revision)
	if len(result) != 16 {
		return "", fmt.Errorf("invalid user agent")
	}
	return result, nil
}

func MakeComparableVersion(raw_version string) (string, error) {
	return (&CodoonUserAgent{Version: raw_version}).MakeComparableVersion()
}

//0001.0001.000001->"1.1.1"
func ParseVersionFromComparable(comparable_version string) (string, error) {
	var (
		vers     []string
		major    int64
		minor    int64
		revision int64
		err      error
	)
	vers = strings.Split(comparable_version, ".")
	if len(vers) < 3 {
		return "", fmt.Errorf("invaid input")
	}
	major, err = strconv.ParseInt(vers[0], 10, 32)
	if err != nil {
		return "", fmt.Errorf("invalid input")
	}
	minor, err = strconv.ParseInt(vers[1], 10, 32)
	if err != nil {
		return "", fmt.Errorf("invalid input")
	}
	revision, err = strconv.ParseInt(vers[2], 10, 32)
	if err != nil {
		return "", fmt.Errorf("invalid input")
	}
	return fmt.Sprintf("%d.%d.%d", major, minor, revision), nil
}

const (
	GT  = 0
	GTE = 1
	EQ  = 2
	ST  = 3
	STE = 4
)

// Compare codoon app version
func CompareVersion(version_a string, version_b string, oper int) (bool, error) {
	// oper:(0, u'>'), (1, u'>='), (2, u'=='), (3, u'<'), (4, u'<='), (other, u'不限制')
	if oper == -1 {
		return true, nil
	}
	int_list_a := []int{}
	int_list_b := []int{}

	if version_a == "" {
		version_a = "0.0.0"
	}

	if version_b == "" {
		version_b = "0.0.0"
	}

	err := StringToIntList(version_a, &int_list_a)
	if err != nil {
		fmt.Errorf("Version format error[version:%v]", version_a)
		return false, err
	}
	for len(int_list_a) < 3 {
		int_list_a = append(int_list_a, 0)
	}
	err = StringToIntList(version_b, &int_list_b)
	if err != nil {
		fmt.Errorf("Version format error[version:%v]", version_b)
		return false, err
	}
	for len(int_list_b) < 3 {
		int_list_b = append(int_list_b, 0)
	}
	if oper == 0 {
		for i := 0; i < len(int_list_a); i++ {
			if int_list_a[i] > int_list_b[i] {
				return true, nil
			} else if int_list_a[i] < int_list_b[i] {
				return false, nil
			} else {
				continue
			}
		}
		return false, nil
	} else if oper == 1 {
		for i := 0; i < 3; i++ {
			if int_list_a[i] > int_list_b[i] {
				return true, nil
			} else if int_list_a[i] < int_list_b[i] {
				return false, nil
			} else {
				continue
			}
		}
		return true, nil
	} else if oper == 2 {
		for i := 0; i < 3; i++ {
			if int_list_a[i] > int_list_b[i] {
				return false, nil
			} else if int_list_a[i] < int_list_b[i] {
				return false, nil
			} else {
				continue
			}
		}
		return true, nil
	} else if oper == 3 {
		for i := 0; i < 3; i++ {
			if int_list_a[i] > int_list_b[i] {
				return false, nil
			} else if int_list_a[i] < int_list_b[i] {
				return true, nil
			} else {
				continue
			}
		}
		return false, nil
	} else if oper == 4 {
		for i := 0; i < 3; i++ {
			if int_list_a[i] > int_list_b[i] {
				return false, nil
			} else if int_list_a[i] < int_list_b[i] {
				return true, nil
			} else {
				continue
			}
		}
		return true, nil
	} else {
		return true, nil
	}
}

func StringToIntList(s string, int_list *[]int) error {
	string_list := strings.Split(s, ".")
	for _, value := range string_list {
		i, err := strconv.Atoi(value)
		if err != nil {
			fmt.Println("Convert error[%d]", value)
			return err
		}
		*int_list = append(*int_list, i)
	}
	return nil
}

//解析http form或者body中的json为go结构体
func ParseHttpReqToArgs(r *http.Request, args interface{}) error {
	var err error
	ct := r.Header.Get("Content-Type")
	if ct == "application/json" {
		var body []byte
		body, err = ioutil.ReadAll(r.Body)
		if err != nil {
			clog.Logger.Error("UpdateUserInfo read body err : %s,%v", r.FormValue("user_id"), err)
			return err
		}
		clog.Logger.Debug("body %s", string(body))
		defer r.Body.Close()
		if err := json.Unmarshal(body, args); err != nil {
			clog.Logger.Error("Unmarshal body : %s,%s,%v", r.FormValue("user_id"), string(body), err)
			return err
		}
	} else {
		err = r.ParseForm()
		if nil != err {
			clog.Logger.Error("r.ParseForm err : %v", err)
			err = errcode.NewInternalError(errcode.DecodeErrCode, err)
		}
		err = ParseForm(r.Form, args)
		if nil != err {
			clog.Logger.Error("ParseForm err : %v", err)
			err = errcode.NewInternalError(errcode.DecodeErrCode, err)
		}

	}

	return err
}

// 解析form,body,param为go结构
func ParseHttpParamsToArgs(r *http.Request, args_strcut interface{}) error {
	var err error
	var args map[string]interface{} = make(map[string]interface{}, 1)

	if nil != r.Body {
		decoder := json.NewDecoder(r.Body)
		decoder.UseNumber()
		decoder.Decode(&args)

	}
	r.Body.Close()
	for k, param := range r.URL.Query() {
		value_int, err := strconv.ParseInt(param[0], 10, 0)
		if err != nil {
			args[k] = param[0]
		} else {
			args[k] = value_int
		}
	}
	err = r.ParseForm()
	if nil != err {
		for key, value := range r.Form {
			value_int, err := strconv.ParseInt(value[0], 10, 0)
			if err != nil {
				args[key] = value[0]
			} else {
				args[key] = value_int
			}
		}
	}

	args["user_agent"] = r.Header.Get("User-Agent")
	// b, err := json.Marshal(&args)
	// if len(b) <= 1024 {
	// 	Logger.Info("api request args: %s", string(b))
	// }
	err = structutil.MapToStruct(args_strcut, args, "json")
	return err
}

func WriteRespToBody(w http.ResponseWriter, resp interface{}) error {

	b, err := json.Marshal(resp)
	if err != nil {
		clog.Logger.Error("Marshal json to bytes error :%v", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return err
	}
	w.Write(b)
	return err
}

// func SendResponse(c *gin.Context, http_code int, data interface{}, err error) error {

// 	if err != nil {
// 		c.String(http_code, http.StatusText(http_code))
// 		return nil
// 	}

// 	b, err := json.Marshal(&data)
// 	if err != nil {
// 		clog.Logger.Error("Marshal json to bytes error :%v", err)
// 	}

// 	c.Writer.Write(b)

// 	return err
// }

func SendResponse(c *gin.Context, http_code int, data interface{}, err error) error {
	var resp CodoonApiResponse = CodoonApiResponse{"OK", nil, "success"}
	if err != nil {
		//CheckError(err)
		is_user_err, code, info := errcode.IsUserErr(err)
		if is_user_err {
			resp.Status = "Error"
			resp.Data = code
			resp.Desc = info
			fmt.Println("code :", code)
		} else {
			c.String(http_code, http.StatusText(http_code))

			// if 500 == http_code {
			// 	host_name, _ := os.Hostname()
			// 	//go SendSmsChina("8613518103463,8615108445455,8618380340755", fmt.Sprintf("500 code : %s %s %s %v", c.Request.Method, c.Request.URL.String(), host_name, err))
			// 	go SendAlertMail([]string{"duhb@codoon.com", "yanglin@codoon.com", "zhouhang@codoon.com"}, fmt.Sprintf("500 code : %s %s %s %s %v", c.Request.Method, c.Request.URL.String(), host_name, local.TraceId(), err))
			// }

			return nil
		}
	} else {
		resp.Data = data
	}
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.Header().Set("ServerTime", strconv.FormatInt(time.Now().Unix(), 10))
	//CheckCrossdomain(c)

	b, err := json.Marshal(&resp)
	if err != nil {
		clog.Logger.Error("Marshal json to bytes error :%v", err)
	}

	c.Writer.Header().Del("Content-length")
	if 0 != len(b) {
		//c.Writer.Header().Set("Content-length",strconv.Itoa(len(b)))
	}

	if len(b) > 30000 {
		clog.Logger.Info(string(b[:30000]))
	} else {
		clog.Logger.Info(string(b))
	}

	clog.Logger.Info("+++++++++++response header: %v ", c.Writer.Header())
	c.Writer.Write(b)

	return err
}

//复用
func SendRequest(http_method, urls string, req_body interface{}, req_form map[string]string, req_raw interface{}) (int, string, error) {
	/*
		tr := &http.Transport{
			DisableKeepAlives: true,
		}
		client := &http.Client{Transport: tr, Timeout: 10 * time.Second}
	*/
	form := url.Values{}
	var err error = nil
	var request *http.Request
	var body []byte

	if nil != req_body {
		request, _ = http.NewRequest(http_method, urls, nil)
		request.Header.Set("Content-Type", "application/json")
		b, _ := json.Marshal(req_body)
		request.Body = ioutil.NopCloser(strings.NewReader(string(b)))
	} else if nil != req_form {
		for key, value := range req_form {
			form.Set(key, value)
		}
		if "GET" == http_method {
			request, _ = http.NewRequest(http_method, urls+"?"+form.Encode(), nil)
		} else {
			request, _ = http.NewRequest(http_method, urls, strings.NewReader(form.Encode()))
		}
		request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	} else if nil != req_raw {
		var query = []byte(req_raw.(string))
		request, _ = http.NewRequest(http_method, urls, bytes.NewBuffer(query))
		request.Header.Set("Content-Type", "text/plain")
	}

	if nil == req_body && nil == req_form && nil == req_raw {
		request, _ = http.NewRequest(http_method, urls, nil)
	}

	clog.Logger.Debug("request %v", request)

	response, err := http_client_cluster.HttpClientClusterDo(request)
	if nil != err {
		err = errcode.NewInternalError(errcode.HttpErrCode, err)
		clog.Logger.Error("send request err :%v", err)
		return http.StatusNotFound, "", err
	}
	// avoid goroutine leak without closing body
	defer response.Body.Close()

	body, err = ioutil.ReadAll(response.Body)
	if nil != err {
		body = make([]byte, 0)
	}
	return response.StatusCode, string(body), err

}

//以form参数 发送http request
func SendFormRequest(http_method, urls string, req_form map[string]string) (int, string, error) {
	return SendRequest(http_method, urls, nil, req_form, nil)
}

//以结构体为参数 发送http request,结构体会被序列化为json 写入body
func SendJsonRequest(http_method, urls string, req_body interface{}) (int, string, error) {
	return SendRequest(http_method, urls, req_body, nil, nil)
}

//直接发送原始request
func SendRawRequest(http_method, urls string, req_raw interface{}) (int, string, error) {
	return SendRequest(http_method, urls, nil, nil, req_raw)
}

func SendRequestSecure(http_method, urls string, req_form map[string]string, secret string) (int, string, error) {
	/*
		tr := &http.Transport{
			DisableKeepAlives: true,
		}
		client := &http.Client{Transport: tr}
	*/
	form := url.Values{}
	var err error = nil
	var request *http.Request
	var body []byte

	if nil != req_form {
		for key, value := range req_form {
			form.Set(key, value)
		}
		if "GET" == http_method {
			request, _ = http.NewRequest(http_method, urls+"?"+form.Encode(), nil)
		} else {
			request, _ = http.NewRequest(http_method, urls, strings.NewReader(form.Encode()))
		}
		request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		request.Header.Set("Authorization", secret)
	}

	clog.Logger.Debug("request %v", request)

	response, err := http_client_cluster.HttpClientClusterDo(request)
	if nil != err {
		err = errcode.NewInternalError(errcode.HttpErrCode, err)
		clog.Logger.Error("send request err :%v", err)
		return http.StatusNotFound, "", err
	}

	if response.StatusCode >= http.StatusOK && response.StatusCode <= http.StatusPartialContent {
		defer response.Body.Close()
		body, err = ioutil.ReadAll(response.Body)
		if nil == err {
			clog.Logger.Debug("body:%v", string(body))
		}
		return response.StatusCode, string(body), err
	} else {
		err = errcode.NewInternalError(errcode.HttpErrCode, fmt.Errorf("http code :%d", response.StatusCode))
		clog.Logger.Error("send request err :%v", err)
		return response.StatusCode, "", err
	}

}

func HttpRequest(method, addr string, params map[string]string) ([]byte, error) {
	form := url.Values{}
	for k, v := range params {
		form.Set(k, v)
	}

	var request *http.Request
	var err error = nil
	if method == "GET" || method == "DELETE" {
		request, err = http.NewRequest(method, addr+"?"+form.Encode(), nil)
		if err != nil {
			return nil, err
		}
	} else {
		request, err = http.NewRequest(method, addr, strings.NewReader(form.Encode()))
		if err != nil {
			return nil, err
		}
		request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}

	//client := &http.Client{}
	response, err := http_client_cluster.HttpClientClusterDo(request)
	if nil != err {
		log.Printf("httpRequest: Do request (%+v) error:%v", request, err)
		return nil, err
	}
	defer response.Body.Close()
	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Printf("httpRequest: read response error:%v", err)
		return nil, err
	}
	return data, nil
}

//解析http form为go结构体
func ParseForm(form url.Values, obj interface{}) error {
	objT := reflect.TypeOf(obj)
	objV := reflect.ValueOf(obj)
	if !structutil.IsStructPtr(objT) {
		return fmt.Errorf("%v must be  a struct pointer", obj)
	}
	objT = objT.Elem()
	objV = objV.Elem()

	for i := 0; i < objT.NumField(); i++ {
		fieldV := objV.Field(i)
		if !fieldV.CanSet() {
			continue
		}

		fieldT := objT.Field(i)
		tags := strings.Split(fieldT.Tag.Get("form"), ",")
		var tag string
		if len(tags) == 0 || len(tags[0]) == 0 {
			tag = fieldT.Name
		} else if tags[0] == "-" {
			continue
		} else {
			tag = tags[0]
		}

		value := form.Get(tag)
		if len(value) == 0 {
			continue
		}

		switch fieldT.Type.Kind() {
		case reflect.Bool:
			if strings.ToLower(value) == "on" || strings.ToLower(value) == "1" || strings.ToLower(value) == "yes" {
				fieldV.SetBool(true)
				continue
			}
			if strings.ToLower(value) == "off" || strings.ToLower(value) == "0" || strings.ToLower(value) == "no" {
				fieldV.SetBool(false)
				continue
			}
			b, err := strconv.ParseBool(value)
			if err != nil {
				return err
			}
			fieldV.SetBool(b)
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			x, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				return err
			}
			fieldV.SetInt(x)
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			x, err := strconv.ParseUint(value, 10, 64)
			if err != nil {
				return err
			}
			fieldV.SetUint(x)
		case reflect.Float32, reflect.Float64:
			x, err := strconv.ParseFloat(value, 64)
			if err != nil {
				return err
			}
			fieldV.SetFloat(x)
		case reflect.Interface:
			fieldV.Set(reflect.ValueOf(value))
		case reflect.String:
			fieldV.SetString(value)
		case reflect.Struct:
			switch fieldT.Type.String() {
			case "time.Time":
				format := time.RFC3339
				if len(tags) > 1 {
					format = tags[1]
				}
				t, err := time.Parse(format, value)
				if err != nil {
					return err
				}
				fieldV.Set(reflect.ValueOf(t))
			}
		case reflect.Slice:
			if fieldT.Type == sliceOfInts {
				formVals := form[tag]
				fieldV.Set(reflect.MakeSlice(reflect.SliceOf(reflect.TypeOf(int(1))), len(formVals), len(formVals)))
				for i := 0; i < len(formVals); i++ {
					val, err := strconv.Atoi(formVals[i])
					if err != nil {
						return err
					}
					fieldV.Index(i).SetInt(int64(val))
				}
			} else if fieldT.Type == sliceOfStrings {
				formVals := form[tag]
				fieldV.Set(reflect.MakeSlice(reflect.SliceOf(reflect.TypeOf("")), len(formVals), len(formVals)))
				for i := 0; i < len(formVals); i++ {
					fieldV.Index(i).SetString(formVals[i])
				}
			}
		}
	}
	return nil
}
