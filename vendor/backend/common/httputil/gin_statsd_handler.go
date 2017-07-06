// by liudan
package httputil

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"runtime"
	"runtime/debug"
	"strings"
	"third/g2s"
	"third/gin"
	"time"
)

const (
	STATS_RPT_CHAN_BUF = 1e5
)

var (
	_statter    g2s.Statter
	_statsRptCh chan *StatsRpt
	_reg        *regexp.Regexp = regexp.MustCompile(`[:/]`)
	_sampleRate float32        = 0.1
	_tableName                 = "backend_service_api"
)

type StatsRpt struct {
	Dur        time.Duration
	Service    string
	Api        string
	HttpMethod string
	HttpCode   int
}

func initStatsD(addr string) {
	if addr == "" {
		log.Printf("initStatsD, addr is empty")
		os.Exit(1)
	}

	statter, err := g2s.Dial("udp", addr)
	if err != nil {
		log.Println("initStatsD, %s error:%v", addr, err)
		os.Exit(1)
	}
	_statter = statter
	_statsRptCh = make(chan *StatsRpt, STATS_RPT_CHAN_BUF)
	go consumeStats()
}

func Counter(bucket string, n ...int) {
	_statter.Counter(_sampleRate, bucket, n...)
}

func Timing(bucket string, d ...time.Duration) {
	_statter.Timing(_sampleRate, bucket, d...)
}

func Gauge(bucket string, v ...string) {
	_statter.Gauge(_sampleRate, bucket, v...)
}

func bucketName(service, api, method string, httpCode, grn int, gcn int64) string {
	service = strings.Replace(service, ":", ".", -1)
	hn, _ := os.Hostname()
	hn = strings.Replace(hn, ":", ".", -1)
	// return fmt.Sprintf("%s,service=%s,api=%s,method=%s,http_code=%d,goroutine=%d,gc=%d",
	// 	_tableName, service, api, method, httpCode, grn, gcn)
	//  be aware the little hack: we prefixed a gauge value for go routine to reduce sysmtem call
	return fmt.Sprintf("%s,service=%s,instance=%s,api=%s,method=%s,http_code=%d:%d|g",
		_tableName, service, hn, api, method, httpCode, grn)

}

func consumeStats() {
	for {
		select {
		case s := <-_statsRptCh:
			grn := runtime.NumGoroutine()
			gcs := &debug.GCStats{}
			debug.ReadGCStats(gcs)
			bucket := bucketName(s.Service, s.Api, s.HttpMethod, s.HttpCode, grn, gcs.NumGC)
			Timing(bucket, s.Dur)
		}
	}
}

func GinSimpleStatter() gin.HandlerFunc {
	return GinStatter("api-stats.in.codoon.com:8125", "")
}

// gin发送statsd统计中间件
// example: GinStatter("statsd.in.codoon.com", "")
func GinStatter(statsdAddr, service string) gin.HandlerFunc {
	initStatsD(statsdAddr)

	return func(c *gin.Context) {
		start := time.Now()

		path := c.Request.URL.Path
		for i := range c.Params {
			strings.Replace(path, c.Params[i].Value, "."+c.Params[i].Key, -1)
		}

		c.Next()

		_statsRptCh <- &StatsRpt{
			Dur:        time.Since(start),
			Service:    c.Request.Host,
			Api:        path,
			HttpMethod: c.Request.Method,
			HttpCode:   c.Writer.Status(),
		}
	}
}
