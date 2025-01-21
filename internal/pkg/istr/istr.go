package istr

import (
	"encoding/json"
	"fmt"
	"reflect"
)

func GetString(v interface{}) string {
	elem := reflect.Indirect(reflect.ValueOf(v))
	switch elem.Kind() {
	case reflect.Array, reflect.Slice, reflect.Map, reflect.Struct:
		bs, _ := json.Marshal(v)
		return fmt.Sprintf("%s", string(bs))
	case reflect.Bool:
		return fmt.Sprintf("%v", v)
	default:
		return fmt.Sprintf("%v", v)
	}
}
