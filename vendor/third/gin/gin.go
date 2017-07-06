// Copyright 2014 Manu Martinez-Almeida.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package gin

import (
	"html/template"
	"log"
	"math"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"third/gin/render"
	"third/httprouter"
	"time"
)

const (
	AbortIndex            = math.MaxInt8 / 2
	MIMEJSON              = "application/json"
	MIMEHTML              = "text/html"
	MIMEXML               = "application/xml"
	MIMEXML2              = "text/xml"
	MIMEPlain             = "text/plain"
	MIMEPOSTForm          = "application/x-www-form-urlencoded"
	MIMEPOSTForm2B        = "application/x-www-form-urlencode" // be compatible with codoon Android. WTF!
	MIMEMultipartPOSTForm = "multipart/form-data"
)

type (
	HandlerFunc func(*Context)

	// Represents the web framework, it wraps the blazing fast httprouter multiplexer and a list of global middlewares.
	Engine struct {
		*RouterGroup
		HTMLRender         render.Render
		Default404Body     []byte
		Default405Body     []byte
		pool               sync.Pool
		allNoRouteNoMethod []HandlerFunc
		noRoute            []HandlerFunc
		noMethod           []HandlerFunc
		router             *httprouter.Router
		logger             []LoggerInfo
		adminRouteOnce     *sync.Once
	}

	HandlerInfo struct {
		Method  string
		Path    string
		Handler HandlerFunc
	}
)

func Raw() *Engine {
	engine := &Engine{
		adminRouteOnce: &sync.Once{},
	}
	engine.RouterGroup = &RouterGroup{
		Handlers:     nil,
		absolutePath: "/",
		engine:       engine,
	}
	engine.router = httprouter.New()
	engine.Default404Body = []byte("404 page not found")
	engine.Default405Body = []byte("405 method not allowed")
	engine.router.NotFound = engine.handle404
	engine.router.MethodNotAllowed = engine.handle405
	engine.pool.New = func() interface{} {
		c := &Context{Engine: engine}
		c.Writer = &c.writermem
		return c
	}
	return engine
}

// Returns a new blank Engine instance without any middleware attached.
// The most basic configuration
func New() *Engine {
	engine := Raw()
	engine.Use(StdoutLogger)
	engine.useAdminRoute()
	return engine
}

// Returns a Engine instance with the Logger and Recovery already attached.
func Default() *Engine {
	engine := Raw()
	engine.Use(Recovery(), StdoutLogger)
	engine.useAdminRoute()
	return engine
}

func (engine *Engine) LoadHTMLGlob(pattern string) {
	if IsDebugging() {
		render.HTMLDebug.AddGlob(pattern)
		engine.HTMLRender = render.HTMLDebug
	} else {
		templ := template.Must(template.ParseGlob(pattern))
		engine.SetHTMLTemplate(templ)
	}
}

func (engine *Engine) LoadHTMLFiles(files ...string) {
	if IsDebugging() {
		render.HTMLDebug.AddFiles(files...)
		engine.HTMLRender = render.HTMLDebug
	} else {
		templ := template.Must(template.ParseFiles(files...))
		engine.SetHTMLTemplate(templ)
	}
}

func (engine *Engine) SetHTMLTemplate(templ *template.Template) {
	engine.HTMLRender = render.HTMLRender{
		Template: templ,
	}
}

// Adds handlers for NoRoute. It return a 404 code by default.
func (engine *Engine) NoRoute(handlers ...HandlerFunc) {
	engine.noRoute = handlers
	engine.rebuild404Handlers()
}

func (engine *Engine) NoMethod(handlers ...HandlerFunc) {
	engine.noMethod = handlers
	engine.rebuild405Handlers()
}

func (engine *Engine) Use(middlewares ...HandlerFunc) {
	engine.RouterGroup.Use(middlewares...)
	engine.rebuild404Handlers()
	engine.rebuild405Handlers()
}

func (engine *Engine) rebuild404Handlers() {
	engine.allNoRouteNoMethod = engine.combineHandlers(engine.noRoute)
}

func (engine *Engine) rebuild405Handlers() {
	engine.allNoRouteNoMethod = engine.combineHandlers(engine.noMethod)
}

func (engine *Engine) handle404(w http.ResponseWriter, req *http.Request) {
	c := engine.createContext(w, req, nil, engine.allNoRouteNoMethod)
	// set 404 by default, useful for logging
	c.Writer.WriteHeader(404)
	c.Next()
	if !c.Writer.Written() {
		if c.Writer.Status() == 404 {
			c.Data(-1, MIMEPlain, engine.Default404Body)
		} else {
			c.Writer.WriteHeaderNow()
		}
	}
	engine.reuseContext(c)
}

func (engine *Engine) handle405(w http.ResponseWriter, req *http.Request) {
	c := engine.createContext(w, req, nil, engine.allNoRouteNoMethod)
	// set 405 by default, useful for logging
	c.Writer.WriteHeader(405)
	c.Next()
	if !c.Writer.Written() {
		if c.Writer.Status() == 405 {
			c.Data(-1, MIMEPlain, engine.Default405Body)
		} else {
			c.Writer.WriteHeaderNow()
		}
	}
	engine.reuseContext(c)
}

// ServeHTTP makes the router implement the http.Handler interface.
func (engine *Engine) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	engine.router.ServeHTTP(writer, request)
}

func (engine *Engine) Run(addr string) error {
	debugPrint("Listening and serving HTTP on %s\n", addr)
	if err := http.ListenAndServe(addr, engine); err != nil {
		return err
	}
	return nil
}

func (engine *Engine) RunTLS(addr string, cert string, key string) error {
	debugPrint("Listening and serving HTTPS on %s\n", addr)
	if err := http.ListenAndServeTLS(addr, cert, key, engine); err != nil {
		return err
	}
	return nil
}

func (engine *Engine) RigsterHttpHandler(hi HandlerInfo) {
	switch hi.Method {
	case "GET":
		engine.GET(hi.Path, hi.Handler)
	case "DELETE":
		engine.DELETE(hi.Path, hi.Handler)
	case "POST":
		engine.POST(hi.Path, hi.Handler)
	case "PUT":
		engine.PUT(hi.Path, hi.Handler)
	default:
		engine.GET(hi.Path, hi.Handler)
	}
}

func (engine *Engine) HandleSignal(signals ...os.Signal) {
	defer log.Println("gin: ByeBye!")
	sig := make(chan os.Signal, 1)
	if len(signals) == 0 {
		signals = append(signals, os.Interrupt, syscall.SIGTERM)
	}
	signal.Notify(sig, signals...)

	s := <-sig
	log.Printf("gin: graceful exit action from signal [%s]", s.String())
	gracefulExit()
}

// graceful exit
var exitOnce sync.Once

func gracefulExit() {
	onceFunc := func() {
		log.Println("gin: graceful exiting...")
		setExit(true)
		wait := func() <-chan struct{} {
			c := make(chan struct{})
			go func() {
				wgReqs.Wait()
				c <- struct{}{}
			}()
			return c
		}
		select {
		case <-wait():
			log.Println("gin: graceful exit OK")
		case <-time.After(60 * time.Second):
			log.Println("gin: graceful exit timeout")
		}
	}
	exitOnce.Do(onceFunc)
}

// Deprecated, please use RegisterLoggerInfo and RegisterHandlerInfo instead.
// gin admin server, for dynamic set log level, graceful exit, pprof, etc.
func UseAdminServer(addr string, logger []LoggerInfo, handler []HandlerInfo) *Engine {
	engine := New()
	engine.logger = logger
	engine.useAdminRoute()

	for _, h := range handler {
		engine.RigsterHttpHandler(h)
	}

	go func() {
		if err := engine.Run(addr); err != nil {
			log.Printf("run UseAdminServer [addr:%s] error:%v", addr, err)
			os.Exit(1)
		}
	}()

	return engine
}

func (engine *Engine) useAdminRoute() {
	f := func() {
		g := engine.Group("/admin/gin")
		{
			// log level
			g.GET("/show_log_level", engine.showloglevelHandler)
			g.POST("/set_log_level", engine.setloglevelHandler)
			// graceful exit
			g.GET("/gracefulexit", engine.gracefulExitHandler)
			// pprof
			g.GET("/debug/pprof/", WrapF(pprof.Index))
			g.GET("/debug/pprof/:name", pprofHandler)
			g.POST("/debug/pprof/:name", pprofHandler)
			g.POST("/debug/set_block_rate", setBlockProfileRateHanlder)

		}
	}

	engine.adminRouteOnce.Do(f)

}

// by liudan
// RegisterLoggerInfo register loggerInfo into gin engine.
// In this way, you can get loggerInfo level through http://gin-service-addr/admin/gin/show_log_level,
// and set loggerInfo level through http://gin-service-addr/admin/gin/set_log_level.
func (engine *Engine) RegisterLoggerInfo(loggerInfo []LoggerInfo) {
	engine.logger = loggerInfo
}

func (engine *Engine) RegisterHandlerInfo(handler []HandlerInfo) {
	for _, h := range handler {
		engine.RigsterHttpHandler(h)
	}
}

func (engine *Engine) showloglevelHandler(c *Context) {
	ret := []interface{}{}
	for _, l := range engine.logger {
		_levels := l.LLogger.GetLevelExt()
		levels := []map[string]string{}
		for k, v := range _levels {
			levels = append(levels, map[string]string{"module": k, "level": getLevelName(v)})
		}
		ret = append(ret, map[string]interface{}{
			"name":   l.Name,
			"levels": levels,
		})
	}
	codoonRsp(c, "OK", ret, "")
}

type LogLevelReq struct {
	Name   string `form:"name" binding:"required"`
	Module string `form:"module"`
	Level  string `form:"level" binding:"required"`
}

func (engine *Engine) setloglevelHandler(c *Context) {
	req := &LogLevelReq{}
	if !c.Bind(req) {
		codoonRsp(c, "Error", "", "missing params")
		return
	}

	for _, l := range engine.logger {
		if l.Name == req.Name {
			module := req.Module
			if module == "" {
				module = "*"
			}
			levelName := strings.ToUpper(req.Level)
			l.LLogger.SetLevelExt(getNameLevel(levelName), module)
		}
	}

	engine.showloglevelHandler(c)
}

func (engine *Engine) gracefulExitHandler(c *Context) {
	log.Printf("gin: graceful exit action from http api [%s]", c.ClientIP())
	go func() {
		gracefulExit()
		os.Exit(0)

	}()
	codoonRsp(c, "OK", "", "graceful exiting")
	c.Writer.Flush()
	time.Sleep(1 * time.Second) // wait for 1 second ensure flush data to client
}

func codoonRsp(c *Context, status string, data interface{}, desc interface{}) {
	c.JSON(http.StatusOK, H{
		"Status":      status,
		"Data":        data,
		"Description": desc,
	})
}

func pprofHandler(c *Context) {
	// there is a hard-coding in net/http/pprof package, so we rewrite `index` route
	name := c.Param("name")
	log.Printf("[%s]", name)
	switch name {
	case "cmdline":
		pprof.Cmdline(c.Writer, c.Request)
	case "profile":
		pprof.Profile(c.Writer, c.Request)
	case "symbol":
		pprof.Symbol(c.Writer, c.Request)
	case "trace":
		pprof.Trace(c.Writer, c.Request)
	default:
		pprof.Handler(name).ServeHTTP(c.Writer, c.Request)

	}
}

func setBlockProfileRateHanlder(c *Context) {
	srate := c.PostForm("rate")
	rate, err := strconv.Atoi(srate)
	if err != nil {
		codoonRsp(c, "Error", "", "parse rate failed:"+err.Error())
		return
	}
	runtime.SetBlockProfileRate(rate)
	codoonRsp(c, "OK", "", "设置成功")
}

// 2017-01-05 by liudan
// AllowCORS make engine accept CORS request from any Origin.
// Make sure OPTIONS route do not conflict with other route.
func (engine *Engine) AllowCORS() {
	engine.Use(func(c *Context) {
		if origin := c.Request.Header.Get("Origin"); origin != "" {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
			c.Writer.Header().Set("Access-Control-Allow-Headers",
				"Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		}
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
		}
	})
}
