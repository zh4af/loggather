package sentryclient

import (
	"backend/common/clog"
	"backend/common/config"
	"fmt"
	"third/raven-go"
)

var SentryClient *raven.Client

//add tag to param; modify by yuanxiang
func InitSentryClientWithUrl(url string, tag map[string]string) error {
	var err error
	fmt.Println(url)
	SentryClient, err = raven.NewWithTags(url, tag)
	if nil != err {
		clog.Logger.Error("init sentry client err")
		return err
	}
	return nil
}

func trace() *raven.Stacktrace {
	return raven.NewStacktrace(2, 2, nil)
}

//发送sentry
//不能放入协程执行，否则无法追踪到调用链，要注意性能问题
func SendErrorToSentry(err error) {
	if nil == SentryClient {
		return
	}
	var in_err error
	packet := raven.NewPacket(err.Error(), raven.NewException(err, trace()))

	eventID, ch := SentryClient.Capture(packet, nil)
	in_err = <-ch
	message := fmt.Sprintf("Error event with id %s,%v", eventID, in_err)
	clog.Logger.Error(message)
}

//记录错误日志 并且发送sentry
func ErrorAndSentry(err error, format string, v ...interface{}) {
	if nil != clog.Logger {
		clog.Logger.Error(format, v...)
		SendErrorToSentry(err)
	}
}

//废弃
// 废弃原因 命名不当 性能不佳 －>SendErrorToSentry
func CheckError(err error) {
	if nil == SentryClient {
		return
	}
	var in_err error
	packet := raven.NewPacket(err.Error(), raven.NewException(err, trace()))

	eventID, ch := SentryClient.Capture(packet, nil)
	in_err = <-ch
	message := fmt.Sprintf("Error event with id %s,%v", eventID, in_err)
	clog.Logger.Error(message)
}

//废弃
// 废弃原因 耦合过紧 －>InitSentryClientWithUrl
func InitSentryClient() error {
	var err error
	fmt.Println(config.Config.SentryUrl)
	SentryClient, err = raven.NewWithTags(config.Config.SentryUrl, map[string]string{"codoon": "test"})
	if nil != err {
		clog.Logger.Error("init sentry client err")
		return err
	}
	return nil
}
