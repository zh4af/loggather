## How to use customized `gin` in codoon?

### Features
* Graceful exit
* Change log level at runtime
* Support custom handlers

### Sample Code


```go
package main

import (
	"net/http"
	"os"
	"third/gin"
	"third/go-logging"
)

var log *logging.Logger = logging.MustGetLogger("testlog")

func initLog() []gin.LoggerInfo {
	backend := logging.NewLogBackend(os.Stdout, "", 0)
	leveledBackend := logging.AddModuleLevel(backend)
	leveledBackend.SetLevel(logging.INFO, "")
	log.SetBackend(leveledBackend)
	return []gin.LoggerInfo{
		gin.LoggerInfo{
			Name:    "defautllogger",
			LLogger: leveledBackend,
		},
	}
}

func hiHandler(c *gin.Context) {
	// SetReqID set reqID+1 into header. If reqID == 0, it will get reqID from header and set reqID+1 to header
	c.SetReqID(2016)
	// ClientIP return client real IP based on codoon habit
	log.Info("ip:%s", c.ClientIP())
	// GetReqID return request id based on codoon habit
	log.Info("req_id:%d", c.GetReqID())
	c.JSON(http.StatusOK, gin.H{"rsp": "hi"})
}

func main() {
	loggers := initLog()
	// prepare customized handlers for AdminServer
	handlers := []gin.HandlerInfo{
		gin.HandlerInfo{
			Method:  "GET",
			Path:    "/hi",
			Handler: hiHandler,
		},
	}
	// AdminServer is only utilized for OPS, your logic should launch another ginengine server.
	ginengine := gin.UseAdminServer(":8082", loggers, handlers)
	// By default, HandleSignal captures interrupt, kill signals.
	// Your can pass other signas to it.
	ginengine.HandleSignal()

}

```

### Usage

* Graceful exit: send intterupt(`Ctrl+C`)/kill(`kill -9 pid`) signal to process or `curl http://localhost:8082/admin/gracefulexit` 
* Show log level: `http://localhost:8082/admin/show_log_level`

```json
{
  "Data": [
    {
      "levels": [
        {
          "level": "INFO",
          "module": ""
        }
      ],
      "name": "defautllogger"
    }
  ],
  "Description": "",
  "Status": "OK"
}
```

* Change log level: `curl http://localhost:8082/admin/set_log_level -d 'name=defautllogger&level=DEBUG'`

```json
{
  "Data": [
    {
      "levels": [
        {
          "level": "DEBUG",
          "module": ""
        }
      ],
      "name": "defautllogger"
    }
  ],
  "Description": "",
  "Status": "OK"
}
```

* Call custom handler: `curl http://localhost:8082/hi`

### Limitations
* `multiLogger` of `go-logging` do NOT support change log level at runtime