package server

import (
	"net/http"
	"third/gin"
	"time"

	// "blast/common/util"
	"backend/common/clog"
	"backend/common/httputil"
	"github.com/zh4af/loggather/protocol"
)

func ReportLogHandle(c *gin.Context) {
	defer httputil.MyRecovery()
	handle_start_time := time.Now()

	var err error
	var req protocol.LogGatherReport
	var reply protocol.LogGatherResp
	var http_code = http.StatusOK

	if err = httputil.ParseHttpParamsToArgs(c.Request, &req); nil != err {
		clog.Logger.Error("parse http req err: %v", err)
		http_code = http.StatusBadRequest
		goto Info
	}

	err = ReportLog(&req, &reply)

Info:
	httputil.SendResponse(c, http_code, reply, err)
	clog.Logger.Info("[cmd:ReportLog][FileName:%s][Cost:%dus][Err:%v]",
		req.FileName, time.Now().Sub(handle_start_time).Nanoseconds()/1000, err)
}
