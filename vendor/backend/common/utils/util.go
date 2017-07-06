package utils

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"
)

type bufferPool struct {
	*sync.Pool
}

func NewBufferPool() *bufferPool {
	return &bufferPool{
		&sync.Pool{New: func() interface{} {
			return bytes.NewBuffer(make([]byte, 0, 1024))
		}},
	}
}
func (bp *bufferPool) Get() *bytes.Buffer {
	return (bp.Pool.Get()).(*bytes.Buffer)
}
func (bp *bufferPool) Put(b *bytes.Buffer) {
	b.Truncate(0)
	bp.Pool.Put(b)
}

func ParseToLocalTime(str string) time.Time {
	tm, _ := time.ParseInLocation("2006-01-02 15:04:05", str, time.Now().Location())
	return tm
}

func ParseDateToLocalTime(str string) time.Time {
	tm, _ := time.ParseInLocation("2006-01-02", str, time.Now().Location())
	return tm
}

func TimeToLocalFormat(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}

func WritePid() error {
	pid_fp, err := os.OpenFile("./server.pid", os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "open pid file failed[%s]\n", err)
		return err
	}
	defer pid_fp.Close()

	pid := os.Getpid()

	pid_fp.WriteString(strconv.Itoa(pid))
	return nil
}

func num64(i interface{}) interface{} {
	switch i.(type) {
	case nil:
		return int64(0)
	case bool:
		b := i.(bool)
		if b {
			return int64(1)
		}
		return int64(0)
	case int8:
		return int64(i.(int8))
	case int16:
		return int64(i.(int16))
	case int32 /*rune*/ :
		return int64(i.(int32))
	case int:
		return int64(i.(int))
	case int64:
		return i.(int64)
	case uint8 /*byte*/ :
		return uint64(i.(uint8))
	case uint16:
		return uint64(i.(uint16))
	case uint32:
		return uint64(i.(uint32))
	case uint:
		return uint64(i.(uint))
	case uint64:
		return uint64(i.(uint64))
	case float32: //precision
		return uint64(i.(float32))
	case float64: //precision
		return uint64(i.(float64))
	case *int8:
		ptr := i.(*int8)
		if ptr == nil {
			return int64(0)
		}
		return int64(*ptr)
	case *int16:
		ptr := i.(*int16)
		if ptr == nil {
			return int64(0)
		}
		return int64(*ptr)
	case *int32 /* *rune */ :
		ptr := i.(*int32)
		if ptr == nil {
			return int64(0)
		}
		return int64(*ptr)
	case *int:
		ptr := i.(*int)
		if ptr == nil {
			return int64(0)
		}
		return int64(*ptr)
	case *int64:
		ptr := i.(*int64)
		if ptr == nil {
			return int64(0)
		}
		return int64(*ptr)
	case *uint8 /* *byte */ :
		ptr := i.(*uint8)
		if ptr == nil {
			return uint64(0)
		}
		return uint64(*ptr)
	case *uint16:
		ptr := i.(*uint16)
		if ptr == nil {
			return uint64(0)
		}
		return uint64(*ptr)
	case *uint:
		ptr := i.(*uint)
		if ptr == nil {
			return uint64(0)
		}
		return uint64(*ptr)
	case *uint32:
		ptr := i.(*uint32)
		if ptr == nil {
			return uint64(0)
		}
		return uint64(*ptr)
	case *uint64:
		ptr := i.(*uint64)
		if ptr == nil {
			return uint64(0)
		}
		return uint64(*ptr)
	default:
		return i
	}
}

func MustNumI64(i interface{}) int64 {
	v := num64(i)
	switch v.(type) {
	case int64:
		return int64(v.(int64)) //float may be lost precision
	case uint64:
		return int64(v.(uint64)) //float and unsigned may be lost precision
	default:
		panic("unknown-type")
	}
}

func NumI64(i interface{}) (int64, error) {
	v := num64(i)
	switch v.(type) {
	case int64:
		return int64(v.(int64)), nil //float may be lost precision
	case uint64:
		return int64(v.(uint64)), nil //float and unsigned may be lost precision
	default:
		return 0, errors.New("unknown-type")
	}
}

func MustNumU64(i interface{}) uint64 {
	v := num64(i)
	switch v.(type) {
	case int64:
		return uint64(v.(int64)) //float may be lost precision
	case uint64:
		return uint64(v.(uint64)) //float may be lost precision
	default:
		panic("unknown-type")
	}
}

func NumU64(i interface{}) (uint64, error) {
	v := num64(i)
	switch v.(type) {
	case int64:
		return uint64(v.(int64)), nil //float may be lost precision
	case uint64:
		return uint64(v.(uint64)), nil //float may be lost precision
	default:
		return 0, errors.New("unknown-type")
	}
}

func GetLastReturnPos(src []byte) int {
	if src == nil {
		return -1
	}

	for i := len(src) - 1; i >= 0; i-- {
		if src[i] == 10 { // \n
			return i
		}
	}
	return -1
}
