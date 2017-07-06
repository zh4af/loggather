package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"
)

type MysqlConfig struct {
	MysqlConn            string
	MysqlConnectPoolSize int
}

type LogConfig struct {
	LogDir      string
	LogFile     string
	LogLevel    string
	LogFormat   string
	ProcessName string
}

type RedisConfig struct {
	RedisConn      string
	RedisPasswd    string
	ReadTimeout    int
	ConnectTimeout int
	WriteTimeout   int
	IdleTimeout    int
	MaxIdle        int
	MaxActive      int
	RedisDb        string
}

type RPCSetting struct {
	Addr string
	Net  string
}

type CeleryQueue struct {
	Url   string
	Queue string
}

type OssConfig struct {
	AccessKeyId     string
	AccessKeySecret string
	Region          string
	Bucket          string
}

type Configure struct {
	MysqlSetting  map[string]MysqlConfig
	RedisSetting  map[string]RedisConfig
	RpcSetting    map[string]RPCSetting
	CelerySetting map[string]CeleryQueue
	OssSetting    map[string]OssConfig
	//httpclient    map[string]OssConfig
	SentryUrl     string
	LogDir        string    //不推荐
	LogFile       string    //不推荐
	LogLevel      string    //不推荐
	LogSetting    LogConfig //推荐
	Listen        string
	RpcListen     string
	External      map[string]string
	ExternalInt64 map[string]int64
	GormDebug     bool   //sql 输出开关
	StaticDir     string //静态文件目录设置
	Environment   string //环境变量区分
}

var Config *Configure
var g_config_file_last_modify_time time.Time
var g_local_conf_file string

// 从file加载配置,并初始化到config结构体
// example: InitConfigFile("/var/www/codoon/feedserver/config/config.json", cfg)
func InitConfigFile(filename string, config *Configure) error {

	fmt.Println("filename", filename)
	fi, err := os.Stat(filename)
	if err != nil {
		fmt.Println("ReadFile: ", err.Error())
		return err
	}

	if g_config_file_last_modify_time.Equal(fi.ModTime()) {
		return nil //fmt.Errorf("Config File Have No Change")
	} else {
		g_config_file_last_modify_time = fi.ModTime()
	}

	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Println("ReadFile: ", err.Error())
		return err
	}

	if err := json.Unmarshal(bytes, config); err != nil {
		err = fmt.Errorf("unmarshal error :%s", string(bytes))
		log.Println(err.Error())
		return err
	}
	fmt.Println("conifg :", *config)
	g_local_conf_file = filename
	Config = config
	return nil
}

// by liudan 2010-10-21
// 从etcd或file加载配置,如果filename为""则从etcd加载,并把json反序列化到config结构体,etcd路径固定   /config/%s/%s", service, env
// example: GetCfgFromEtcdOrFile([]string{"http://etcd.in.codoon.com"},"", "webmiddleware",cfg)
func GetCfgFromEtcdOrFile(etcdAddr, filename, service string, cfg interface{}) ([]byte, error) {
	if etcdAddr != "" {
		if s, err := LoadContentFromEtcd(strings.Split(etcdAddr, ","), service, "online"); err != nil {
			return nil, err
		} else {
			data := []byte(s)
			return data, json.Unmarshal(data, cfg)
		}
	}
	if filename != "" {
		if data, err := ioutil.ReadFile(filename); err != nil {
			return nil, err
		} else {
			return data, json.Unmarshal(data, cfg)
		}
	}

	return nil, errors.New("both etcdAddr and filename are empty")
}
