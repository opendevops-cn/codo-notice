package biz

import (
	"context"
	"fmt"
	"regexp"
	"time"
)

type IRouterUseCase interface {
	List(ctx context.Context, query RouterQuery) ([]*Router, uint32, error)
	Get(ctx context.Context, id uint32) (*Router, error)
	Create(ctx context.Context, router *Router) (*Router, error)
	Update(ctx context.Context, router *Router, opts RouterUpdateOptions) error
	Delete(ctx context.Context, ids []uint32) error
}

type IRouterRepo interface {
	List(ctx context.Context, query RouterQuery) ([]*Router, error)
	Count(ctx context.Context, query RouterQuery) (uint32, error)
	Get(ctx context.Context, id uint32) (*Router, error)
	Create(ctx context.Context, router *Router) (*Router, error)
	Update(ctx context.Context, router *Router, opts RouterUpdateOptions) error
	Delete(ctx context.Context, ids []uint32) error
}

type RouterList []*Router

// Router 模板定义
type Router struct {
	// ID 主键
	ID uint32
	// 创建时间
	CreatedAt time.Time
	// 更新时间
	UpdatedAt time.Time
	// 创建人
	CreatedBy string
	// 更新人
	UpdatedBy string
	// 路由名称，唯一，最大长度256
	Name string
	// 路由描述，最大长度1024
	Description string
	// 是否激活，可选值："yes"/"no"，默认值为"yes"，最大长度16
	Status string
	// 通道ID，默认值为0
	ChannelID uint32
	// 触发条件，二维数组形式
	ConditionList QueryGroup
}

type CombineType int32

const (
	CombineTypeAND = CombineType(1)
	CombineTypeOR  = CombineType(2)
)

type QueryGroup struct {
	// 条件组
	Queries []*Query
	Groups  []*QueryGroup
	// CombineType 目前只用 and
	CombineType CombineType
}

func (x *QueryGroup) MatchLabels(labels map[string]string) (bool, error) {
	if x.CombineType == CombineTypeAND {
		for _, query := range x.Queries {
			matched, err := query.MatchLabels(labels)
			if err != nil {
				return false, err
			}
			if !matched {
				return false, nil
			}
		}
		for _, group := range x.Groups {
			matched, err := group.MatchLabels(labels)
			if err != nil {
				return false, err
			}
			if !matched {
				return false, nil
			}
		}
		return true, nil
	}
	// 如果是 OR 组合，只要有一个匹配即可
	if x.CombineType == CombineTypeOR {
		for _, query := range x.Queries {
			matched, err := query.MatchLabels(labels)
			if err != nil {
				return false, err
			}
			if matched {
				return true, nil
			}
		}
		for _, group := range x.Groups {
			matched, err := group.MatchLabels(labels)
			if err != nil {
				return false, err
			}
			if matched {
				return true, nil
			}
		}
		return false, nil
	}
	return false, fmt.Errorf("[QueryGroup][MatchLabels] unknown combine type: %d", x.CombineType)
}

func (x *Query) MatchLabels(labels map[string]string) (bool, error) {
	labelValue := labels[x.Label]

	switch x.Operator {
	case OperatorEQ:
		return labelValue == x.Value, nil
	case OperatorNEQ:
		return labelValue != x.Value, nil
	case OperatorREGEX:
		m, err := regexp.MatchString(x.Value, labelValue)
		if err != nil {
			return false, fmt.Errorf("[Query][MatchLabels] regex compile failed: %w, label=%s, pattern=%s", err, x.Label, x.Value)
		}
		return m, nil
	case OperatorNREGEX:
		m, err := regexp.MatchString(x.Value, labelValue)
		if err != nil {
			return false, fmt.Errorf("[Query][MatchLabels] regex compile failed: %w, label=%s, pattern=%s", err, x.Label, x.Value)
		}
		return !m, nil
	}
	return false, fmt.Errorf("[Query][MatchLabels] unknown operator: %s", x.Operator)
}

const (
	OperatorEQ     = "==" // equals
	OperatorNEQ    = "!=" // does not equal
	OperatorREGEX  = "=~" // matches regex
	OperatorNREGEX = "!~" // does not match regex
)

// Query 查询条件定义
type Query struct {
	// 标签名称
	Label string
	// 操作符
	Operator string
	// 条件值
	Value string
	// 索引值
	Index int32
	// 状态值
	Status int32
}

type RouterQuery struct {
	ListAll bool
	// 每页条数
	PageSize int32
	// 第几页
	PageNum int32
	// 正序或倒序 ascend  descend
	Order string
	// 全局搜索关键字
	SearchText string
	// 搜索字段
	SearchField string
	// 排序关键字
	Field string
	// yes:缓存，no:不缓存
	Cache string
	// 多字段搜索,精准匹配
	FilterMap map[string]interface{}
}

type RouterUpdateOptions struct {
	// 路由名称，唯一，最大长度256
	Name bool
	// 路由描述，最大长度1024
	Description bool
	// 是否激活，可选值："yes"/"no"，默认值为"yes"，最大长度16
	Status bool
	// 通道ID，默认值为0
	ChannelId bool
	// 触发条件，二维数组形式
	ConditionList bool
}

type RouterUseCase struct {
	repo IRouterRepo
}

func NewIRouterUseCase(x *RouterUseCase) IRouterUseCase {
	return x
}

func NewRouterUseCase(repo IRouterRepo) *RouterUseCase {
	return &RouterUseCase{repo: repo}
}

func (x *RouterUseCase) List(ctx context.Context, query RouterQuery) ([]*Router, uint32, error) {
	list, err := x.repo.List(ctx, query)
	if err != nil {
		return nil, 0, err
	}
	query.ListAll = true
	count, err := x.repo.Count(ctx, query)
	if err != nil {
		return nil, 0, err
	}

	return list, count, nil
}

func (x *RouterUseCase) Get(ctx context.Context, id uint32) (*Router, error) {
	result, err := x.repo.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (x *RouterUseCase) Create(ctx context.Context, router *Router) (*Router, error) {
	result, err := x.repo.Create(ctx, router)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (x *RouterUseCase) Update(ctx context.Context, router *Router, opts RouterUpdateOptions) error {
	err := x.repo.Update(ctx, router, opts)
	if err != nil {
		return err
	}

	return nil
}

func (x *RouterUseCase) Delete(ctx context.Context, ids []uint32) error {
	err := x.repo.Delete(ctx, ids)
	if err != nil {
		return err
	}

	return nil
}
