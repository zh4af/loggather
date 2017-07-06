package clog

import (
	"backend/common/config"
	"fmt"
	"log"
	"third/go-logging"
	"third/raven-go"
)

//废弃
// 废弃原因 耦合过紧 －>InitLoggerByConfig
func InitLogger(process_name string) (*logging.Logger, error) {

	if nil == config.Config {
		fmt.Println("please set logger path and logger level with SetLoggerLevel SetLoggerDir")
		return nil, nil
	}
	//format_str := "%{color}%{level}:[%{time:2006-01-02 15:04:05.000}][goroutine:%{goroutinecount}][%{shortfile}]%{color:reset}[%{message}]"
	format_str := "%{color}%{level:.4s}:%{time:2006-01-02 15:04:05.000}[%{id:03x}][%{goroutineid}/%{goroutinecount}] %{shortfile}%{color:reset} %{message}"
	Logger = logging.MustGetLogger(process_name)

	sql_log_fp, err := logging.NewFileLogWriter(config.Config.LogDir+"/"+process_name+".log.mysql", false, 1024*1024*1024)
	if err != nil {
		fmt.Println("open file[%s.mysql] failed[%s]", config.Config.LogFile, err)
		return nil, err
	}

	MysqlLogger = log.New(sql_log_fp, "", log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile)

	info_log_fp, err := logging.NewFileLogWriter(config.Config.LogDir+"/"+process_name+".log", false, 1024*1024*1024)
	if err != nil {
		fmt.Println("open file[%s] failed[%s]", config.Config.LogFile, err)
		return nil, err
	}

	err_log_fp, err := logging.NewFileLogWriter(config.Config.LogDir+"/"+process_name+".log.wf", false, 1024*1024*1024)
	if err != nil {
		fmt.Println("open file[%s.wf] failed[%s]", config.Config.LogFile, err)
		return nil, err
	}

	backend_info := logging.NewLogBackend(info_log_fp, "", 0)
	backend_err := logging.NewLogBackend(err_log_fp, "", 0)
	format := logging.MustStringFormatter(format_str)
	backend_info_formatter := logging.NewBackendFormatter(backend_info, format)
	backend_err_formatter := logging.NewBackendFormatter(backend_err, format)

	backend_info_leveld = logging.AddModuleLevel(backend_info_formatter)
	switch config.Config.LogLevel {
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

	//add sentry log author:yuanxiang
	sentry_client, err := raven.NewWithTags(config.Config.SentryUrl, map[string]string{"servicename": "servicename"})
	if nil != err {
		log.Fatalf("init sentry client err")
		return nil, err
	}
	sentry_err := logging.NewSentryBackend(sentry_client, logging.ERROR)
	sentry_formatter := logging.NewBackendFormatter(sentry_err, format)
	sentry_err_leveld := logging.AddModuleLevel(sentry_formatter)
	sentry_err_leveld.SetLevel(logging.ERROR, "")

	logging.SetBackend(backend_info_leveld, backend_err_leveld, sentry_err_leveld)

	return Logger, err
}

//废弃
// 废弃原因 耦合过紧 －>InitLoggerByConfig
func InitLogger1(process_name, format_str string) (*logging.Logger, error) {

	if nil == config.Config {
		fmt.Println("please set logger path and logger level with SetLoggerLevel SetLoggerDir")
		return nil, nil
	}

	Logger = logging.MustGetLogger(process_name)

	sql_log_fp, err := logging.NewFileLogWriter(config.Config.LogDir+"/"+process_name+".log.mysql", false, 1024*1024*1024)
	if err != nil {
		fmt.Println("open file[%s.mysql] failed[%s]", config.Config.LogFile, err)
		return nil, err
	}

	MysqlLogger = log.New(sql_log_fp, "", log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile)

	info_log_fp, err := logging.NewFileLogWriter(config.Config.LogDir+"/"+process_name+".log", false, 1024*1024*1024)
	if err != nil {
		fmt.Println("open file[%s] failed[%s]", config.Config.LogFile, err)
		return nil, err
	}

	err_log_fp, err := logging.NewFileLogWriter(config.Config.LogDir+"/"+process_name+".log.wf", false, 1024*1024*1024)
	if err != nil {
		fmt.Println("open file[%s.wf] failed[%s]", config.Config.LogFile, err)
		return nil, err
	}

	backend_info := logging.NewLogBackend(info_log_fp, "", 0)
	backend_err := logging.NewLogBackend(err_log_fp, "", 0)
	format := logging.MustStringFormatter(format_str)
	backend_info_formatter := logging.NewBackendFormatter(backend_info, format)
	backend_err_formatter := logging.NewBackendFormatter(backend_err, format)

	backend_info_leveld = logging.AddModuleLevel(backend_info_formatter)
	switch config.Config.LogLevel {
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

	//add sentry log author:yuanxiang
	sentry_client, err := raven.NewWithTags(config.Config.SentryUrl, map[string]string{"servicename": "servicename"})
	if nil != err {
		log.Fatalf("init sentry client err")
		return nil, err
	}
	sentry_err := logging.NewSentryBackend(sentry_client, logging.ERROR)
	sentry_formatter := logging.NewBackendFormatter(sentry_err, format)
	sentry_err_leveld := logging.AddModuleLevel(sentry_formatter)
	sentry_err_leveld.SetLevel(logging.ERROR, "")

	logging.SetBackend(backend_info_leveld, backend_err_leveld, sentry_err_leveld)

	return Logger, err
}

// 废弃原因 耦合过紧 －>InitLoggerByConfig
func SetLoggerDir(logDir string) {
	if config.Config == nil {
		config.Config = &config.Configure{}
	}

	config.Config.LogDir = logDir
}

// 废弃原因 耦合过紧 －>InitLoggerByConfig
func SetLoggerLevel(LogLevel string) {
	if config.Config == nil {
		config.Config = &config.Configure{}
	}

	config.Config.LogLevel = LogLevel
}
