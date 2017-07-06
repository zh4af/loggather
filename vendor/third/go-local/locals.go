package local

import (
	"third/context"
	"sync"
	// "time"
)

var locals = &struct {
	sync.RWMutex
	ctx map[uint64]context.Context
}{
	ctx: make(map[uint64]context.Context),
}

func get(goid uint64) context.Context {
	locals.RLock()
	ctx := locals.ctx[goid]
	locals.RUnlock()

	if ctx == nil {
		ctx = context.Background()
	}
	return ctx
}

func set(goid uint64, ctx context.Context) {
	locals.Lock()
	locals.ctx[goid] = ctx
	locals.Unlock()
}

func temp(goid uint64, key, val interface{}) context.Context {
	ctx := context.WithValue(get(goid), key, val)
	set(goid, ctx)
	return ctx
}

func clear(goid uint64) context.Context {
	locals.Lock()
	ctx := locals.ctx[goid]
	delete(locals.ctx, goid)
	locals.Unlock()
	return ctx
}

/*
var DelayDuration = 5 * time.Second

func delayClear(goid uint64, duration ...time.Duration) {
	d := DelayDuration
	if len(duration) > 0 {
		d = duration[0]
	}
	go func(goid uint64, duration time.Duration) {
		time.Sleep(duration)
		clear(goid)
	}(Goid(), d)
}
*/

func Get() context.Context {
	return get(Goid())
}

func Set(ctx context.Context) {
	set(Goid(), ctx)
}

/*
func DelaySet(ctx context.Context, duration ...time.Duration) {
	goid := Goid()
	set(goid, ctx)
	delayClear(goid, duration...)
}
*/

func Clear() context.Context {
	return clear(Goid())
}

func Temp(key, val interface{}) context.Context {
	return temp(Goid(), key, val)
}

func Value(key interface{}) interface{} {
	ctx := Get()
	if ctx == nil {
		return nil
	}

	return ctx.Value(key)
}

func Go(fn func()) {
	GoContext(Get(), fn)
}

func GoContext(ctx context.Context, fn func()) {
	go func() {
		Set(ctx)
		defer Clear()

		fn()
	}()
}
