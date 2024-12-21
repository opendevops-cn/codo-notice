package data

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/ccheers/xpkg/generic/arrayx"
)

// ReqTable 表格请求参数
type ReqTable struct {
	PageSize    int                    // 每页条数
	PageNum     int                    // 第几页
	Order       string                 // 正序或倒序 ascend  descend
	SearchText  string                 // 全局搜索关键字
	SearchField string                 // 搜索字段
	Field       string                 // 排序关键字
	Cache       string                 // yes:缓存，no:不缓存
	FilterMap   map[string]interface{} // 多字段搜索,精准匹配
}

type commonCondition struct {
	listAll bool
	offset  int
	limit   int
	where   string
	order   string
	values  []interface{}
}

// convertQuery
// notSearch - 模糊搜索排除字段
// transFields - 前端搜索转移字段
func (rp *ReqTable) convertQuery(
	model interface{},
	notSearch []string,
	transFields map[string]string,
) (*commonCondition, error) {
	buildInFields := []string{"id", "created_at", "updated_at"}
	notSearch = append(notSearch, buildInFields...)
	notSearch = append(notSearch, []string{"", "-"}...)
	if transFields == nil {
		transFields = make(map[string]string)
	}

	// 先获取数据库所有字段 /通过反射方式映射数据库查询字段
	var fds []string
	s := reflect.ValueOf(model).Elem()
	for i := 0; i < s.NumField(); i++ {
		tag := s.Type().Field(i).Tag.Get("json")
		if !arrayx.ContainsAny(notSearch, []string{tag}) {
			fds = append(fds, tag)
		}
	}
	// 默认排序字段
	if !arrayx.ContainsAny(fds, []string{rp.Field}) && !arrayx.ContainsAny(buildInFields, []string{rp.Field}) {
		rp.Field = "updated_at"
	}
	sortField := rp.Field
	if v, ok := transFields[sortField]; ok {
		sortField = v
	}
	order := fmt.Sprintf("%s asc", sortField)
	if rp.Order == "descend" {
		order = fmt.Sprintf("%s desc", sortField)
	}

	// 确认精准搜索字段
	var wheres []string
	var values []interface{}
	for k, v := range rp.FilterMap {
		if arrayx.ContainsAny(fds, []string{k}) {
			if v, ok := transFields[k]; ok {
				k = v
			}
			wheres = append(wheres, fmt.Sprintf("(`%s` in (?))", k))
			values = append(values, v)
		}
	}

	// 确认模糊搜索字段
	var likeWhere []string
	if rp.SearchText != "" {
		search := strings.Split(rp.SearchField, ",")
		for _, v := range fds {
			if rp.SearchField == "" || arrayx.ContainsAny(search, []string{v}) {
				if val, ok := transFields[v]; ok {
					v = val
				}
				likeWhere = append(likeWhere, fmt.Sprintf("(`%s` REGEXP ?)", v))
				values = append(values, rp.SearchText)
			}
		}
	}

	if len(likeWhere) != 0 {
		wheres = append(wheres, fmt.Sprintf("(%s)", strings.Join(likeWhere, " OR ")))
	}

	offset := rp.PageSize * (rp.PageNum - 1)
	return &commonCondition{
		offset: offset,
		limit:  rp.PageSize,
		where:  strings.Join(wheres, " AND "),
		values: values,
		order:  order,
	}, nil
}
