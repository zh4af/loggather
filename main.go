package main

import (
	"flag"
	"fmt"
	"runtime"

	// "blast/common/util"
	"backend/common/clog"
	"backend/common/config"
	"github.com/zh4af/loggather/client"
	"github.com/zh4af/loggather/server"
)

const DEFAULT_CONF_FILE = "./conf/loggather.conf"
const (
	SERVERNAME = "loggather"
)

var EtcdHost string

var g_conf_file string = DEFAULT_CONF_FILE
var g_actor_type string
var g_cpupro_file string = ""
var g_mempro_file string = ""
var g_config config.Configure

const (
	ACTOR_TYPE_CLIENT = "client"
	ACTOR_TYPE_SERVER = "server"
)

func init() {
	const usage = "loggather [-c config_file][-a actor_type][-p cpupro file][-m mempro file]"
	flag.StringVar(&g_conf_file, "c", "", usage)
	flag.StringVar(&g_actor_type, "a", "", usage)
	flag.StringVar(&g_cpupro_file, "p", "", usage)
	flag.StringVar(&g_mempro_file, "m", "", usage)
}

func main() {
	//set runtime variable
	runtime.GOMAXPROCS(runtime.NumCPU())
	fmt.Println("runtime.NumCPU() ", runtime.NumCPU())
	//get flag
	flag.Parse()

	fmt.Println(g_conf_file)

	//init config
	// err := common.InitConfigFile(g_conf_file)
	_, err := config.GetCfgFromEtcdOrFile(EtcdHost, g_conf_file, SERVERNAME, &g_config)
	if err != nil {
		fmt.Println(err)
		return
	}

	//init log
	_, err = clog.InitLogger(g_config.LogFile + g_actor_type)
	if err != nil {
		fmt.Println("init log error")
		return
	}

	switch g_actor_type {
	case ACTOR_TYPE_CLIENT:
		client.RunLogClient()
	case ACTOR_TYPE_SERVER:
		fallthrough
	default:
		// go util.RegisterSelfToConsul(g_config.Listen)
		server.StartHttpServer(g_config.Listen)
	}
}
