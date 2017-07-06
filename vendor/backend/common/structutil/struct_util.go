package structutil

import (
	"fmt"
	"reflect"
	"strings"
)

func IsStructPtr(t reflect.Type) bool {
	return t.Kind() == reflect.Ptr && t.Elem().Kind() == reflect.Struct
}

//check the item whether in object
//object must be slice
func HasItem(obj interface{}, item interface{}) bool {
	for i := 0; i < reflect.ValueOf(obj).Len(); i++ {
		if reflect.DeepEqual(reflect.ValueOf(obj).Index(i).Interface(), item) {
			return true
		}
	}
	return false
}

// tag: form, json, 不填默认为json
func MapToStruct(to interface{}, from map[string]interface{}, tag string) error {
	var err error
	toV := reflect.ValueOf(to)
	if toV.Kind() != reflect.Ptr {
		err = fmt.Errorf("not a ptr")
		return err
	}
	if tag == "" {
		tag = "json"
	}

	toT := reflect.TypeOf(to).Elem()
	toV = toV.Elem()

	for i := 0; i < toT.NumField(); i++ {
		fieldV := toV.Field(i)
		if !fieldV.CanSet() {
			continue
		}

		fieldT := toT.Field(i)
		tags := strings.Split(fieldT.Tag.Get(tag), ",")
		var tag string
		if len(tags) == 0 || len(tags[0]) == 0 {
			tag = fieldT.Name
		} else if tags[0] == "-" {
			continue
		} else {
			tag = tags[0]
		}

		value, ok := from[tag]
		if !ok {
			continue
		}
		// fmt.Println("value type: ", reflect.ValueOf(value).Kind())

		// fieldV.Set(reflect.ValueOf(value))
		switch fieldT.Type.Kind() {
		case reflect.Interface:
			fieldV.Set(reflect.ValueOf(value))
		case reflect.String:
			fieldV.SetString(value.(string))
		case reflect.Bool:
			fieldV.SetBool(value.(bool))
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if reflect.ValueOf(value).Kind() == reflect.Int { //
				fieldV.SetInt(int64(value.(int)))
			}
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			if reflect.ValueOf(value).Kind() == reflect.Int {
				fieldV.SetUint(uint64(value.(int)))
			}
		case reflect.Float32, reflect.Float64:
			if reflect.ValueOf(value).Kind() == reflect.Float64 {
				fieldV.SetFloat(value.(float64))
			}
		}

	}
	return err
}

//拷贝结构体
//Be careful to use, from,to must be pointer
func DumpStruct(to interface{}, from interface{}) {
	fromv := reflect.ValueOf(from)
	tov := reflect.ValueOf(to)
	if fromv.Kind() != reflect.Ptr || tov.Kind() != reflect.Ptr {
		return
	}

	from_val := reflect.Indirect(fromv)
	to_val := reflect.Indirect(tov)

	for i := 0; i < from_val.Type().NumField(); i++ {
		fdi_from_val := from_val.Field(i)
		fd_name := from_val.Type().Field(i).Name
		fdi_to_val := to_val.FieldByName(fd_name)

		if fdi_to_val.IsValid() && fdi_from_val.Type() == fdi_to_val.Type() {
			fdi_to_val.Set(fdi_from_val)
		}
	}
}

//拷贝slice
func DumpList(to interface{}, from interface{}) {
	val_from := reflect.ValueOf(from)
	val_to := reflect.ValueOf(to)

	if val_from.Type().Kind() == reflect.Slice && val_to.Type().Kind() == reflect.Slice &&
		val_from.Len() == val_to.Len() {
		for i := 0; i < val_from.Len(); i++ {
			DumpStruct(val_to.Index(i).Addr().Interface(), val_from.Index(i).Addr().Interface())
		}
	}
}
