package clog

import (
	"backend/common/config"
	"errors"
	"fmt"
	"log"
	"os"
	"third/gin"
	"third/go-logging"
)

var Logger *logging.Logger
var MysqlLogger *log.Logger
var backend_info_leveld logging.LeveledBackend

func init() {

	Logger = logging.MustGetLogger("log_not_configed")

	info_log_fp := os.Stdout
	err_log_fp := os.Stderr
	format_str := "%{color}%{level:.4s}:%{time:2006-01-02 15:04:05.000}[%{id:03x}][%{goroutineid}/%{goroutinecount}] %{shortfile}%{color:reset} %{message}"

	backend_info := logging.NewLogBackend(info_log_fp, "", 0)
	backend_err := logging.NewLogBackend(err_log_fp, "", 0)
	format := logging.MustStringFormatter(format_str)
	backend_info_formatter := logging.NewBackendFormatter(backend_info, format)
	backend_err_formatter := logging.NewBackendFormatter(backend_err, format)

	backend_info_leveld = logging.AddModuleLevel(backend_info_formatter)
	backend_info_leveld.SetLevel(logging.NOTICE, "")

	backend_err_leveld := logging.AddModuleLevel(backend_err_formatter)
	backend_err_leveld.SetLevel(logging.WARNING, "")

	logging.SetBackend(backend_info_leveld, backend_err_leveld)
}

func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// 初始化logger
// example: InitLoggerByConfig(&cfg.LogSetting)
func InitLoggerByConfig(config *config.LogConfig) (*logging.Logger, error) {

	if nil == config {
		fmt.Println("please set logger path and logger level with SetLoggerLevel SetLoggerDir")
		return nil, nil
	}

	ok, _ := PathExists(config.LogDir)
	if !ok {
		err := os.MkdirAll(config.LogDir, 0777)
		if nil != err {
			fmt.Println("can't make dir : %s, %v", config.LogDir, err)
			return nil, err
		}
	}

	LogFormat := "%{color}%{level:.4s}:%{time:2006-01-02 15:04:05.000}[%{id:03x}][%{goroutineid}/%{goroutinecount}] %{shortfile}%{color:reset} %{message}"
	if "" == config.LogFormat {
		config.LogFormat = LogFormat
	}

	Logger = logging.MustGetLogger(config.ProcessName)

	sql_log_fp, err := logging.NewFileLogWriter(config.LogDir+"/"+config.LogFile+".mysql", false, 1024*1024*1024)
	if err != nil {
		fmt.Println("open file[%s.mysql] failed[%s]", config.LogFile, err)
		return nil, err
	}

	MysqlLogger = log.New(sql_log_fp, "", log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile)

	info_log_fp, err := logging.NewFileLogWriter(config.LogDir+"/"+config.LogFile, false, 1024*1024*1024)
	if err != nil {
		fmt.Println("open file[%s] failed[%s]", config.LogFile, err)
		return nil, err
	}

	err_log_fp, err := logging.NewFileLogWriter(config.LogDir+"/"+config.LogFile+".wf", false, 1024*1024*1024)
	if err != nil {
		fmt.Println("open file[%s.wf] failed[%s]", config.LogFile, err)
		return nil, err
	}

	backend_info := logging.NewLogBackend(info_log_fp, "", 0)
	backend_err := logging.NewLogBackend(err_log_fp, "", 0)
	format := logging.MustStringFormatter(config.LogFormat)
	backend_info_formatter := logging.NewBackendFormatter(backend_info, format)
	backend_err_formatter := logging.NewBackendFormatter(backend_err, format)

	backend_info_leveld = logging.AddModuleLevel(backend_info_formatter)
	switch config.LogLevel {
	case "ERROR":
		backend_info_leveld.SetLevel(logging.ERROR, "")
	case "WARNING":
		backend_info_leveld.SetLevel(logging.WARNING, "")
	case "NOTICE":
		backend_info_leveld.SetLevel(logging.NOTICE, "")
	case "INFO":
		backend_info_leveld.SetLevel(logging.INFO, "")
	case "DEBUG":
		backend_info_leveld.SetLevel(logging.DEBUG, "")
	default:
		backend_info_leveld.SetLevel(logging.ERROR, "")
	}

	backend_err_leveld := logging.AddModuleLevel(backend_err_formatter)
	backend_err_leveld.SetLevel(logging.WARNING, "")

	logging.SetBackend(backend_info_leveld, backend_err_leveld)

	return Logger, err
}

func ChangeLogLevel(LogLevel string) {
	switch LogLevel {
	case "ERROR":
		backend_info_leveld.SetLevel(logging.ERROR, "")
	case "WARNING":
		backend_info_leveld.SetLevel(logging.WARNING, "")
	case "NOTICE":
		backend_info_leveld.SetLevel(logging.NOTICE, "")
	case "INFO":
		backend_info_leveld.SetLevel(logging.INFO, "")
	case "DEBUG":
		backend_info_leveld.SetLevel(logging.DEBUG, "")
	default:
		backend_info_leveld.SetLevel(logging.ERROR, "")
	}
}

// by liudan 2016.10.14
// RegisterChangeLogLevelToGin register backend_info_leveld into gin engine.
// In this way, you can get backend_info_leveld level through http://gin-service-addr/admin/gin/show_log_level,
// and set backend_info_leveld level through http://gin-service-addr/admin/gin/set_log_level.
// Have a look at http://git.in.codoon.com/third/gin/blob/master/USAGE.md#usage for more.
func RegisterChangeLogLevelToGin(engine *gin.Engine) error {
	if backend_info_leveld == nil {
		return errors.New("logger backend is nil")
	}
	loggerInfos := []gin.LoggerInfo{
		gin.LoggerInfo{
			Name:    "info_log",
			LLogger: backend_info_leveld,
		},
	}
	engine.RegisterLoggerInfo(loggerInfos)
	return nil
}

// 设置log打印深度，如果使用Debugf等函数，则需配置额外深度为1
// example: SetExtraCalldepth(1)
func SetExtraCalldepth(depth int) {
	if nil != Logger {
		Logger.ExtraCalldepth = depth
	}
}

func Debugf(format string, v ...interface{}) {
	if nil != Logger {
		Logger.Debug(format, v...)
	}
}

func Infof(format string, v ...interface{}) {
	if nil != Logger {
		Logger.Info(format, v...)
	}
}

func Noticef(format string, v ...interface{}) {
	if nil != Logger {
		Logger.Notice(format, v...)
	}
}

func Warnf(format string, v ...interface{}) {
	if nil != Logger {
		Logger.Warning(format, v...)
	}
}

func Errorf(format string, v ...interface{}) {
	if nil != Logger {
		Logger.Error(format, v...)
	}
}

// by liudan
var (
	green   = string([]byte{27, 91, 57, 55, 59, 52, 50, 109})
	white   = string([]byte{27, 91, 57, 48, 59, 52, 55, 109})
	yellow  = string([]byte{27, 91, 57, 55, 59, 52, 51, 109})
	red     = string([]byte{27, 91, 57, 55, 59, 52, 49, 109})
	blue    = string([]byte{27, 91, 57, 55, 59, 52, 52, 109})
	magenta = string([]byte{27, 91, 57, 55, 59, 52, 53, 109})
	cyan    = string([]byte{27, 91, 57, 55, 59, 52, 54, 109})
	reset   = string([]byte{27, 91, 48, 109})
)

func ColorForStatus(code int) string {
	switch {
	case code >= 200 && code <= 299:
		return green
	case code >= 300 && code <= 399:
		return white
	case code >= 400 && code <= 499:
		return yellow
	default:
		return red
	}
}

func ColorForMethod(method string) string {
	switch {
	case method == "GET":
		return blue
	case method == "POST":
		return cyan
	case method == "PUT":
		return yellow
	case method == "DELETE":
		return red
	case method == "PATCH":
		return green
	case method == "HEAD":
		return magenta
	case method == "OPTIONS":
		return white
	default:
		return reset
	}
}

func ColorForReset() string {
	return reset
}
