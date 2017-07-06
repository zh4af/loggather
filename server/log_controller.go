package server

import (
	"bytes"
	"compress/gzip"
	"io/ioutil"
	"os"

	"backend/common/clog"
	"backend/common/config"
	"backend/common/errcode"
	// "backend/common/utils"
	"github.com/zh4af/loggather/protocol"
)

// var gBufPool = utils.NewBufferPool()

func ReportLog(req *protocol.LogGatherReport, reply *protocol.LogGatherResp) error {
	var err error
	var out []byte

	if _, err = os.Stat(config.Config.External["LogGatherDir"]); nil != err {
		os.Mkdir(config.Config.External["LogGatherDir"], 0644)
	}
	file_fp, err := os.OpenFile(config.Config.External["LogGatherDir"]+req.FileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	defer file_fp.Close()
	if nil != err {
		clog.Logger.Error("open log file err: %v", err)
		return err
	}

	// var buf_src *bytes.Buffer = gBufPool.Get()
	// var buf_dst *bytes.Buffer = gBufPool.Get()
	// defer gBufPool.Put(buf_src)
	// defer gBufPool.Put(buf_dst)
	// if buf_src == nil || buf_dst == nil {
	// 	err = errcode.NewInternalError(errcode.InternalErrorCode, err)
	// 	clog.Logger.Error("get memory from buf pool err: %v", err)
	// 	return err
	// }
	// if _, err = buf_src.Write(req.LogInfoGzip); nil != err {
	// 	clog.Logger.Error("buf write data err: %v", err)
	// 	return err
	// }
	buf_src := bytes.NewBuffer(req.LogInfoGzip)
	gReader, err := gzip.NewReader(buf_src)
	defer gReader.Close()
	// _, err = gReader.Read(buf_dst.Bytes())
	out, _ = ioutil.ReadAll(gReader)
	if nil != err {
		err = errcode.NewInternalError(errcode.DecodeErrCode, err)
		clog.Logger.Error("decode base64 str err: %v", err)
		return err
	}

	write_n, err := file_fp.Write(out)
	if nil != err {
		clog.Logger.Error("write log file err: %v", err)
		return err
	}

	clog.Logger.Debug("write to log file: %s bytes: %d", req.FileName, write_n)

	return err
}
