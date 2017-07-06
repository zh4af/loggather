// Copyright 2014 Manu Martinez-Almeida.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package gin

import (
	"fmt"
	"log"
	"runtime"
	"third/go-colorable"
	"time"
)

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

var levelNames = []string{
	"CRITICAL",
	"ERROR",
	"WARNING",
	"NOTICE",
	"INFO",
	"DEBUG",
}

type LeveledLogger interface {
	GetLevelExt() map[string]int
	SetLevelExt(int, string)
}

type LoggerInfo struct {
	Name    string
	LLogger LeveledLogger
}

func getLevelName(l int) string {
	if l < 0 || l >= len(levelNames) {
		return "unknown"
	}
	return levelNames[l]
}

func getNameLevel(name string) int {
	for k, v := range levelNames {
		if v == name {
			return k
		}
	}

	return 0
}

func ErrorLogger() HandlerFunc {
	return ErrorLoggerT(ErrorTypeAll)
}

func ErrorLoggerT(typ uint32) HandlerFunc {
	return func(c *Context) {
		c.Next()

		errs := c.Errors.ByType(typ)
		if len(errs) > 0 {
			// -1 status code = do not change current one
			c.JSON(-1, c.Errors)
		}
	}
}

func Logger() HandlerFunc {
	stdlogger := log.New(colorable.NewColorableStdout(), "", 0)
	//errlogger := log.New(os.Stderr, "", 0)

	return func(c *Context) {
		// Start timer
		start := time.Now()

		// Process request
		c.Next()

		// Stop timer
		end := time.Now()
		latency := end.Sub(start)

		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()
		statusColor := colorForStatus(statusCode)
		methodColor := colorForMethod(method)

		stdlogger.Printf("[GIN] %v |%s %3d %s| %12v | %s |%s  %s %-7s %s %s %s\n%s",
			end.Format("2006/01/02 - 15:04:05"),
			statusColor, statusCode, reset,
			latency,
			clientIP,
			methodColor, method, reset,
			c.Request.URL.Path,
			c.Request.URL.String(),
			c.Request.URL.Opaque,
			c.Errors.String(),
		)
	}
}

func colorForStatus(code int) string {
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

func colorForMethod(method string) string {
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

func StdoutLogger(c *Context) {
	start := time.Now()

	c.Next()

	cost := time.Now().Sub(start)
	clientIP := c.ClientIP()
	method := c.Request.Method
	statusCode := c.Writer.Status()
	reqId := c.Request.Header.Get("codoon_request_id")
	if reqId == "" {
		reqId = "-"
	}
	srvCode := c.Request.Header.Get("codoon_service_code")
	if srvCode == "" {
		srvCode = "-"
	}

	// expected format:
	//100.97.90.49 2016-11-11.15:15:56 "POST /api/msgsys_acks HTTP/1.0" 11101515564527644053639076611600 200 67 "-" "CodoonSport(7.2.0 1320;iOS 10.0.2;iPhone)" "58.63.1.102" 0.011 0.009
	fmt.Printf("[GINHTTP] %s %d %s %s %d %s %s %s \"%s\" %.03f\n",
		clientIP,
		runtime.NumGoroutine(),
		time.Now().Format("2006-01-02.15:04:05"),
		method,
		statusCode,
		c.Request.URL.Path,
		reqId,
		srvCode,
		c.Request.UserAgent(),
		cost.Seconds(),
	)

}
