package local

import "reflect"

func loadUint64(val reflect.Value) uint64 {
	switch val.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return uint64(val.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return val.Uint()
	}

	return 0
}

func dumpUint64(val reflect.Value, u uint64) {
	switch val.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		val.SetInt(int64(u))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		val.SetUint(u)
	}
}

func Uint64(key interface{}) uint64 {
	return loadUint64(reflect.ValueOf(Value(key)))
}

func Int64(key interface{}) int64 {
	return int64(Uint64(key))
}

func String(key interface{}) string {
	switch val := Value(key).(type) {
	case string:
		return val
	}

	return ""
}
