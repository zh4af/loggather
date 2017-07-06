package client

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	// "blast/common/util"
	"backend/common/clog"
	"backend/common/config"
	"backend/common/errcode"
	"backend/common/utils"
	"github.com/zh4af/loggather/protocol"
)

const (
	SINGLE_GATHER_NUM = 1024 * 100 // 100K
)

type LogFileRecordInfo struct {
	Data map[string]int `json:"data"` // {"data":{"file1":10,"file2":2,"file3":3}} key:文件名 value:读取的位置
	sync.Mutex
}

var gRecordInfo LogFileRecordInfo
var gRecordFP *os.File
var gBufPool = utils.NewBufferPool()

func RunLogClient() {
	var err error

	tick := time.NewTicker(time.Second * 10)

	gRecordFP, err = os.OpenFile("./log_record_info.json", os.O_RDWR|os.O_CREATE, 0644)
	defer gRecordFP.Close()
	if nil != err {
		clog.Logger.Error("open file err: %v", err)
		return
	}
	buf, err := ioutil.ReadAll(gRecordFP)
	if len(buf) <= 0 {
		gRecordInfo.Data = make(map[string]int, 1)
	} else {
		// decoder := json.NewDecoder(gRecordFP)
		// err = decoder.Decode(&gRecordInfo)
		fmt.Println("buf: ", string(buf))
		json.Unmarshal(buf, &gRecordInfo)
		if nil != err {
			clog.Logger.Error("decode json err: %v", err)
			return
		}
		fmt.Println("gRecordInfo: ", gRecordInfo)
	}
	for {
		select {
		case <-tick.C:
			gatherDirLog()
		}
	}
}

func gatherDirLog() {
	var err error
	var out []byte
	var file_name string // file base name
	var file_list []string
	var wg sync.WaitGroup

	cmd := fmt.Sprintf("lsof +d %s |grep REG |awk '{print $9}'", config.Config.External["LogGatherDir"])
	out, err = exec.Command("/bin/sh", "-c", cmd).Output()
	if err != nil {
		clog.Logger.Error("lsof err: %v", err)
		return
	}
	file_list = strings.Split(string(out), "\n")
	fmt.Println("file_list: ", file_list)
	for i := range file_list {
		if file_list[i] <= "" {
			continue
		}
		file_names := strings.Split(file_list[i], "/")
		if len(file_names) >= 2 {
			file_name = file_names[len(file_names)-1]
		} else {
			file_name = file_list[i]
		}
		fmt.Println("file_name: ", file_name)
		wg.Add(1)
		if stpos, ok := gRecordInfo.Data[file_name]; ok {
			go gatherSingleLog(file_name, stpos, &wg)
		} else {
			go gatherSingleLog(file_name, 0, &wg)
		}
	}
	wg.Wait()

	gRecordFP.Truncate(0)
	buf, err := json.Marshal(&gRecordInfo)
	gRecordFP.Write(buf)
}

// file_name: 文件名
// stpos: 起始的读取位置
func gatherSingleLog(file_name string, stpos int, wg *sync.WaitGroup) {
	defer wg.Done()
	var rbuf []byte = make([]byte, SINGLE_GATHER_NUM)

	fp, err := os.OpenFile(config.Config.External["LogGatherDir"]+file_name, os.O_RDONLY, 0644)
	defer fp.Close()
	if nil != err {
		clog.Logger.Error("open file err: %v", err)
		return
	}
	rn, err := fp.ReadAt(rbuf, int64(stpos))
	if nil != err {
		clog.Logger.Error("read from file: %s err: %v", config.Config.External["LogGatherDir"]+file_name, err)
		return
	}
	// 丢弃最后被截断的一行，放到下次读取
	if rbuf[len(rbuf)-1] != 10 {
		lastRetPos := utils.GetLastReturnPos(rbuf)
		if lastRetPos > 0 {
			rn = lastRetPos + 1
			rbuf = rbuf[:rn]
		}
	}

	var wbuf *bytes.Buffer = gBufPool.Get()
	defer gBufPool.Put(wbuf)
	if wbuf == nil {
		err = errcode.NewInternalError(errcode.InternalErrorCode, err)
		clog.Logger.Error("get memory from buf pool err: %v", err)
		return
	}
	gWriter, err := gzip.NewWriterLevel(wbuf, gzip.BestCompression)
	defer gWriter.Close()
	_, err = gWriter.Write(rbuf)
	err = gWriter.Flush()
	if nil != err {
		clog.Logger.Error("gzip compress data err: %v", err)
		return
	}

	body := protocol.LogGatherReport{
		FileName:    file_name,
		LogInfoGzip: wbuf.Bytes(),
	}
	b, err := json.Marshal(&body)
	req, err := http.NewRequest("POST", config.Config.External["LogReportUrl"], ioutil.NopCloser(strings.NewReader(string(b))))
	rsp, err := http.DefaultClient.Do(req)
	if nil != err || rsp.StatusCode != http.StatusOK {
		clog.Logger.Error("post http to report log err: %v, rsp: %v", err, rsp)
		return
	}

	gRecordInfo.Lock()
	if gRecordInfo.Data == nil {
		gRecordInfo.Data = make(map[string]int, 1)
	}
	gRecordInfo.Data[file_name] = stpos + rn
	gRecordInfo.Unlock()
}
