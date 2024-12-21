package istr

import (
	"encoding/json"
	"fmt"
)

func GetString(v interface{}) string {
	switch t := v.(type) {
	default:
		return ""
	case nil:
		return ""
	case int, int16, int32, int64, uint, uint16, uint32, uint64:
		return fmt.Sprintf("%d", t)
	case float32, float64:
		// 不保留小数位 以免出现歧义 其他情况请使用字符串匹配
		return fmt.Sprintf("%.0f", t)
	case json.Number:
		return t.String()
	case string:
		return t
	}
}
