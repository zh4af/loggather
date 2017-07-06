// by liudan
package httputil

import (
	"backend/common/sentryclient"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"
	"third/context"
	"third/gin"
	"third/http_client_cluster"
	"time"
)

const (
	GIN_CTX = "gin_ctx"
)

var _httpClient *http.Client

// The Client's Transport typically has internal state (cached TCP connections),
// so Clients should be reused instead of created as needed.
// Clients are safe for concurrent use by multiple goroutines.
func init() {
	tr := &http.Transport{
		Dial: func(network, addr string) (conn net.Conn, err error) {
			return net.DialTimeout(network, addr, 5*time.Second)
		},
	}
	_httpClient = &http.Client{
		Transport: tr,
		Timeout:   45 * time.Second,
	}
}

func HttpSimpleRequest(ctx context.Context, method, addr string, params map[string]string) ([]byte, error) {
	data, _, err := HttpRequestWithCode(ctx, method, addr, params)
	return data, err
}

func HttpRequestWithCode(ctx context.Context, method, addr string, params map[string]string) ([]byte, int, error) {
	// form := url.Values{}
	// for k, v := range params {
	// 	form.Set(k, v)
	// }

	// var request *http.Request
	// var err error = nil
	// if method == "GET" || method == "DELETE" {
	// 	request, err = http.NewRequest(method, addr+"?"+form.Encode(), nil)
	// 	if err != nil {
	// 		return nil, 0, err
	// 	}
	// } else {
	// 	request, err = http.NewRequest(method, addr, strings.NewReader(form.Encode()))
	// 	if err != nil {
	// 		return nil, 0, err
	// 	}
	// 	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	// }

	// // inject header to coloring call chain
	// if ctx != nil {
	// 	if c, ok := ctx.Value(GIN_CTX).(*gin.Context); ok {
	// 		request.Header.Set(CODOON_REQUEST_ID, c.Request.Header.Get(CODOON_REQUEST_ID))
	// 		request.Header.Set(CODOON_SERVICE_CODE, c.Request.Header.Get(CODOON_SERVICE_CODE))
	// 	}
	// }

	// response, err := _httpClient.Do(request)
	// if nil != err {
	// 	log.Printf("httpRequest: Do request (%+v) error:%v", request, err)
	// 	return nil, 0, err
	// }
	// defer response.Body.Close()
	// data, err := ioutil.ReadAll(response.Body)
	// if err != nil {
	// 	// send sentry
	// 	CheckError(errors.New(fmt.Sprintf("%s %s %s", method, addr, err.Error())))
	// 	// send stats
	// 	if _statter != nil {
	// 		bucket := fmt.Sprintf("%s_error", strings.Replace(request.URL.Path, "/", "_", -1))
	// 		_statter.Counter(1.0, bucket, 1)
	// 	}
	// 	log.Printf("httpRequest: read response error:%v", err)
	// 	return nil, 0, err
	// }
	// if response.StatusCode != http.StatusOK {
	// 	CheckError(errors.New(fmt.Sprintf("%s %s %d", method, addr, response.StatusCode)))
	// 	if _statter != nil {
	// 		bucket := fmt.Sprintf("%s_%d", strings.Replace(request.URL.Path, "/", "_", -1), response.StatusCode)
	// 		_statter.Counter(1.0, bucket, 1)
	// 	}
	// }
	// return data, response.StatusCode, nil
	return HttpRequestWithHeader(ctx, method, addr, params, nil)
}

//带额外header参数 发送httprequest
func HttpRequestWithHeader(ctx context.Context, method, addr string, params, headers map[string]string) ([]byte, int, error) {
	form := url.Values{}
	for k, v := range params {
		form.Set(k, v)
	}

	var request *http.Request
	var err error = nil
	if method == "GET" || method == "DELETE" {
		request, err = http.NewRequest(method, addr+"?"+form.Encode(), nil)
		if err != nil {
			return nil, 0, err
		}
	} else {
		request, err = http.NewRequest(method, addr, strings.NewReader(form.Encode()))
		if err != nil {
			return nil, 0, err
		}
		request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}

	// inject header to coloring call chain
	if ctx != nil {
		if c, ok := ctx.Value(GIN_CTX).(*gin.Context); ok {
			request.Header.Set(CODOON_REQUEST_ID, c.Request.Header.Get(CODOON_REQUEST_ID))
			request.Header.Set(CODOON_SERVICE_CODE, c.Request.Header.Get(CODOON_SERVICE_CODE))
		}
	}

	// write custom headers
	for k, v := range headers {
		request.Header.Set(k, v)
	}

	response, err := http_client_cluster.HttpClientClusterDo(request)
	if nil != err {
		log.Printf("httpRequest: Do request (%+v) error:%v", request, err)
		return nil, 0, err
	}
	if response != nil && response.Body != nil {
		defer response.Body.Close()
	}
	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		// send sentry
		sentryclient.SendErrorToSentry(errors.New(fmt.Sprintf("%s %s %s", method, addr, err.Error())))
		// send stats
		if _statter != nil {
			bucket := fmt.Sprintf("%s_error", strings.Replace(request.URL.Path, "/", "_", -1))
			_statter.Counter(1.0, bucket, 1)
		}
		log.Printf("httpRequest: read response error:%v", err)
		return nil, 0, err
	}

	if !(response.StatusCode >= http.StatusOK && response.StatusCode <= http.StatusPartialContent) {
		sentryclient.SendErrorToSentry(errors.New(fmt.Sprintf("%s %s %d", method, addr, response.StatusCode)))
		if _statter != nil {
			bucket := fmt.Sprintf("%s_%d", strings.Replace(request.URL.Path, "/", "_", -1), response.StatusCode)
			_statter.Counter(1.0, bucket, 1)
		}
	}
	return data, response.StatusCode, nil
}

//带额外header参数 发送httprequest
func HttpRawRequest(method, addr string, body io.Reader) ([]byte, error) {
	req, err := http.NewRequest(method, addr, body)
	if err != nil {
		return nil, err
	}

	if method == "POST" || method == "PUT" {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}

	client := &http.Client{}
	rsp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer rsp.Body.Close()
	return ioutil.ReadAll(rsp.Body)
}

func HttpRequestPostWithUrl(ctx context.Context, method, addr string, params, urls map[string]string) ([]byte, int, error) {
	form1 := url.Values{}
	for k, v := range params {
		form1.Set(k, v)
	}

	form2 := url.Values{}
	for k, v := range urls {
		form2.Set(k, v)
	}

	var request *http.Request
	var err error = nil
	if method != "POST" {
		return nil, 0, err
	} else {
		request, err = http.NewRequest(method, addr+"?"+form2.Encode(), strings.NewReader(form1.Encode()))
		if err != nil {
			return nil, 0, err
		}
		request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}

	// inject header to coloring call chain
	if ctx != nil {
		if c, ok := ctx.Value(GIN_CTX).(*gin.Context); ok {
			request.Header.Set(CODOON_REQUEST_ID, c.Request.Header.Get(CODOON_REQUEST_ID))
			request.Header.Set(CODOON_SERVICE_CODE, c.Request.Header.Get(CODOON_SERVICE_CODE))
		}
	}

	response, err := http_client_cluster.HttpClientClusterDo(request)
	if nil != err {
		log.Printf("httpRequest: Do request (%+v) error:%v", request, err)
		return nil, 0, err
	}
	if response != nil && response.Body != nil {
		defer response.Body.Close()
	}
	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		// send sentry
		sentryclient.SendErrorToSentry(errors.New(fmt.Sprintf("%s %s %s", method, addr, err.Error())))
		// send stats
		if _statter != nil {
			bucket := fmt.Sprintf("%s_error", strings.Replace(request.URL.Path, "/", "_", -1))
			_statter.Counter(1.0, bucket, 1)
		}
		log.Printf("httpRequest: read response error:%v", err)
		return nil, 0, err
	}

	if !(response.StatusCode >= http.StatusOK && response.StatusCode <= http.StatusPartialContent) {
		sentryclient.SendErrorToSentry(errors.New(fmt.Sprintf("%s %s %d", method, addr, response.StatusCode)))
		if _statter != nil {
			bucket := fmt.Sprintf("%s_%d", strings.Replace(request.URL.Path, "/", "_", -1), response.StatusCode)
			_statter.Counter(1.0, bucket, 1)
		}
	}
	return data, response.StatusCode, nil
}
