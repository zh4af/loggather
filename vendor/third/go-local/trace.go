package local

import (
	"log"
	"net/http"
	"reflect"
	"strconv"
	"sync"
	"third/context"
	// "time"
)

const (
	keyTraceId  = `__trace_id__`
	keySpanId   = `__span_id__`
	keyParentId = `__parent_id__`
	keySpanInc  = `__span_inc__`

	maxSpanPerScope = 100
)

type span struct {
	sync.RWMutex
	id uint64
}

func spanCtx(ctx context.Context, spanId uint64) context.Context {
	return context.WithValue(ctx, keySpanInc, &span{id: spanId * maxSpanPerScope})
}

func nextSpanId() uint64 {
	sp, ok := Value(keySpanInc).(*span)
	if !ok {
		log.Printf("span not in ctx with trace %v", TraceId())
		return 0
	}

	if sp == nil {
		log.Printf("span is nil in trace %v", TraceId())
		return 0
	}

	sp.Lock()
	sp.id++
	nextId := sp.id
	sp.Unlock()
	return nextId
}

func TraceId() string {
	return String(keyTraceId)
}

func SpanId() uint64 {
	return Uint64(keySpanId)
}

func ParentId() uint64 {
	return Uint64(keyParentId)
}

type TraceParam struct {
	TraceId  string
	SpanId   uint64
	ParentId uint64
}

var (
	TraceIdFieldName  = `TraceId`
	SpanIdFieldName   = `SpanId`
	ParentIdFieldName = `ParentId`
)

func parseIdField(args interface{}, fieldName string) uint64 {
	val := reflect.Indirect(reflect.ValueOf(args))
	if val.Kind() != reflect.Struct {
		return 0
	}

	return loadUint64(val.FieldByName(fieldName))
}

func parseField(args interface{}, fieldName string) string {
	val := reflect.Indirect(reflect.ValueOf(args))
	if val.Kind() != reflect.Struct {
		return ""
	}

	val = val.FieldByName(fieldName)

	switch val.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		strconv.FormatInt(val.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		strconv.FormatUint(val.Uint(), 10)
	case reflect.String:
		return val.String()
	}

	return ""
}

func parseTraceId(args interface{}) string {
	return parseField(args, TraceIdFieldName)
}

func parseSpanId(args interface{}) uint64 {
	return parseIdField(args, SpanIdFieldName)
}

func parseParentId(args interface{}) uint64 {
	return parseIdField(args, ParentIdFieldName)
}

func ParseTraceParam(args interface{}) *TraceParam {
	return &TraceParam{
		TraceId:  parseTraceId(args),
		SpanId:   parseSpanId(args),
		ParentId: parseParentId(args),
	}
}

func genTraceInfo(goid uint64, traceId string, spanId, parentId uint64) context.Context {
	ctx := spanCtx(get(goid), spanId)
	ctx = context.WithValue(ctx, keyTraceId, traceId)
	ctx = context.WithValue(ctx, keySpanId, spanId)
	ctx = context.WithValue(ctx, keyParentId, parentId)
	return ctx
}

func GoTrace(traceId string, spanId, parentId uint64, fn func()) context.Context {
	ctx := genTraceInfo(Goid(), traceId, spanId, parentId)
	GoContext(ctx, fn)
	return ctx
}

func GoTraceArgs(args interface{}, fn func()) context.Context {
	return GoTrace(parseTraceId(args), parseSpanId(args), parseParentId(args), fn)
}

func fillTraceMap(args map[string]interface{}) {
	args[TraceIdFieldName] = TraceId()
	args[SpanIdFieldName] = nextSpanId()
	args[ParentIdFieldName] = SpanId()
}

func FillTraceArgs(args interface{}) interface{} {
	switch m := args.(type) {
	case map[string]interface{}:
		fillTraceMap(m)
		return args
	case *map[string]interface{}:
		fillTraceMap(*m)
		return args
	default:
	}

	traceId := TraceId()
	spanId := SpanId()

	val := reflect.ValueOf(args)
	if val.Kind() != reflect.Ptr {
		log.Printf("cannot fill unaddressed value in trace[%v]: %T", traceId, args)
		return args
	}

	val = val.Elem()
	if val.Kind() != reflect.Struct {
		log.Printf("cannot fill unstructure variable in trace[%v]: %T", traceId, args)
		return args
	}

	trace := val.FieldByName(TraceIdFieldName)
	if !trace.CanAddr() || !trace.CanSet() {
		log.Printf("cannot set field `%v` in trace[%v]: %T", TraceIdFieldName, traceId, args)
		return args
	}
	if trace.Kind() == reflect.String {
		trace.SetString(traceId)
	}

	span := val.FieldByName(SpanIdFieldName)
	if !span.CanAddr() || !span.CanSet() {
		log.Printf("cannot set field `%v` in trace[%v]: %T", SpanIdFieldName, traceId, args)
		return args
	}
	dumpUint64(span, nextSpanId())

	parent := val.FieldByName(ParentIdFieldName)
	if !parent.CanAddr() || !parent.CanSet() {
		log.Printf("cannot set field `%v` in trace[%v]: %T", ParentIdFieldName, traceId, args)
		return args
	}
	dumpUint64(parent, spanId)
	return args
}

// Following 2 functions are used in go func() { used here }
// Do not forget `defer Clear()`
func TempTraceInfo(traceId string, spanId, parentId uint64) context.Context {
	goid := Goid()
	ctx := genTraceInfo(goid, traceId, spanId, parentId)
	set(goid, ctx)
	return ctx
}

func TempTraceInfoArgs(args interface{}) context.Context {
	return TempTraceInfo(parseTraceId(args), parseSpanId(args), parseParentId(args))
}

/*
func DelayTrace(traceId, spanId, parentId uint64, duration ...time.Duration) context.Context {
	goid := Goid()
	ctx := genTraceInfo(goid, traceId, spanId, parentId)
	set(goid, ctx)
	delayClear(goid, duration...)
	return ctx
}

func DelayTraceArgs(args interface{}, duration ...time.Duration) context.Context {
	return DelayTrace(parseTraceId(args), parseSpanId(args), parseParentId(args), duration...)
}
*/

func parseHttp(req *http.Request, headerName string) string {
	if req == nil {
		return "0"
	}

	return req.Header.Get(headerName)
}

func parseIdHttp(req *http.Request, headerName string) uint64 {
	if req == nil {
		return 0
	}

	id, err := strconv.ParseUint(parseHttp(req, headerName), 10, 64)
	if err != nil {
		// log.Printf("parse `%v` from http request header error: %v", headerName, err)
		return 0
	}

	return id
}

func parseTraceIdHttp(req *http.Request) string {
	return parseHttp(req, TraceIdFieldName)
}

func parseSpanIdHttp(req *http.Request) uint64 {
	return parseIdHttp(req, SpanIdFieldName)
}

func parseParentIdHttp(req *http.Request) uint64 {
	return parseIdHttp(req, ParentIdFieldName)
}

func ParseTraceParamHttp(req *http.Request) *TraceParam {
	return &TraceParam{
		TraceId:  parseTraceIdHttp(req),
		SpanId:   parseSpanIdHttp(req),
		ParentId: parseParentIdHttp(req),
	}
}

func GoTraceHttp(req *http.Request, fn func()) context.Context {
	return GoTrace(parseTraceIdHttp(req), parseSpanIdHttp(req), parseParentIdHttp(req), fn)
}

func FillTraceHttp(req *http.Request) *http.Request {
	if req == nil {
		return req
	}

	req.Header.Set(TraceIdFieldName, TraceId())
	req.Header.Set(SpanIdFieldName, strconv.FormatUint(nextSpanId(), 10))
	req.Header.Set(ParentIdFieldName, strconv.FormatUint(SpanId(), 10))
	return req
}

func TempTraceInfoHttp(req *http.Request) context.Context {
	return TempTraceInfo(parseTraceIdHttp(req), parseSpanIdHttp(req), parseParentIdHttp(req))
}

/*
func DelayTraceHttp(req *http.Request, duration ...time.Duration) context.Context {
	return DelayTrace(parseTraceIdHttp(req), parseSpanIdHttp(req), parseParentIdHttp(req), duration...)
}
*/
