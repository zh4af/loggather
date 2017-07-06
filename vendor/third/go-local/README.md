# go-local

GoroutineLocal for golang (like ThreadLocal in Java)

## Usage

> 1. Use Go

```go
package main

import (
    "time"
    "third/go-local"
)

func main() {
    local.Temp("key", "value")
    defer local.Clear()

    if local.String("key") == "value" {
        println("main Set OK!")
    } else {
        println("main Set FAILED!")
    }

    local.Go(func() {
        if local.String("key") == "value" {
            println("go Set OK!")
        } else {
            println("go Set FAILED!")
        }
    })

    time.Sleep(time.Second)
}
// prints:
// main Set OK!
// go Set OK!
```

> 2. Use Trace

```go
package main

import "third/go-local"

type T struct {
    local.TraceParam
    S string
}

func main() {
    local.TempTraceInfoArgs(&T{
        TraceParam: local.TraceParam{
            TraceId:  1,
            SpanId:   2,
            ParentId: 0,
        },
    })
    defer local.Clear()

    println(local.TraceId())  // prints: 1
    println(local.SpanId())   // prints: 2
    println(local.ParentId()) // prints: 0

    var t T
    local.FillTraceArgs(&t)

    println(t.TraceId)  // prints: 1
    println(t.SpanId)   // prints: 201
    println(t.ParentId) // prints: 2

}
```

# Author

Lyn Young

