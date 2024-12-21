package biz

import (
	"context"
	"time"
)

type IChannelUseCase interface {
	List(ctx context.Context, query ChannelQuery) ([]*Channel, uint32, error)
	Get(ctx context.Context, id uint32) (*Channel, error)
	Create(ctx context.Context, channel *Channel) (*Channel, error)
	Update(ctx context.Context, channel *Channel, opts ChannelUpdateOptions) error
	Delete(ctx context.Context, ids []uint32) error
}

type IChannelRepo interface {
	List(ctx context.Context, query ChannelQuery) ([]*Channel, error)
	Count(ctx context.Context, query ChannelQuery) (uint32, error)
	Get(ctx context.Context, id uint32) (*Channel, error)
	Create(ctx context.Context, channel *Channel) (*Channel, error)
	Update(ctx context.Context, channel *Channel, opts ChannelUpdateOptions) error
	Delete(ctx context.Context, ids []uint32) error
}

// Channel 渠道定义
type Channel struct {
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
	// 通道名称，唯一，最大长度256
	Name string
	// 用途，不为空，默认值为"default"，最大长度45
	Use string
	// 用户列表
	User []string
	// 用户组ID列表
	Group []uint32
	// 联系点列表
	ContactPoints []*ContactPoint
	// 自定义项列表
	CustomItems []*ContactPoint
	// 字段重写规则，用于短信电话消息
	DefaultRule map[string]string
}

type ContactPoint struct {
	// ID 主键
	Id uint32
	// 创建时间
	CreatedAt time.Time
	// 更新时间
	UpdatedAt time.Time
	// 类型，不为空，默认值为"default"，最大长度16
	Type string
	// 通道ID
	ChannelId uint32
	// 模板ID
	TemplateId uint32
	// 地址，最大长度1024
	Addr string
	// 密钥，最大长度1024
	Secret string
	// 等级列表
	Severity []string
	// 等级描述，最大长度1024
	SeverityDesc string
	// 是否显示，可选值："yes"/"no"，默认值为"no"，最大长度16
	Show string
	// 是否固定，可选值："yes"/"no"，默认值为"yes"，最大长度16
	Fixed string
	// 通知回调参数
	AlertWebhooks []AlertWebhook
}

type ChannelQuery struct {
	ListAll bool
	// 每页条数
	PageSize int32 `protobuf:"varint,1,opt,name=page_size,json=pageSize,proto3" json:"page_size,omitempty"`
	// 第几页
	PageNum int32 `protobuf:"varint,2,opt,name=page_num,json=pageNum,proto3" json:"page_num,omitempty"`
	// 正序或倒序 ascend  descend
	Order string `protobuf:"bytes,3,opt,name=order,proto3" json:"order,omitempty"`
	// 全局搜索关键字
	SearchText string `protobuf:"bytes,4,opt,name=search_text,json=searchText,proto3" json:"search_text,omitempty"`
	// 搜索字段
	SearchField string `protobuf:"bytes,5,opt,name=search_field,json=searchField,proto3" json:"search_field,omitempty"`
	// 排序关键字
	Field string `protobuf:"bytes,6,opt,name=field,proto3" json:"field,omitempty"`
	// yes:缓存，no:不缓存
	Cache string `protobuf:"bytes,7,opt,name=cache,proto3" json:"cache,omitempty"`
	// 多字段搜索,精准匹配
	FilterMap map[string]any `protobuf:"bytes,8,rep,name=filter_map,json=filterMap,proto3" json:"filter_map,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
}

type ChannelUpdateOptions struct {
	Name          bool
	Use           bool
	User          bool
	ContactPoints bool
	CustomItems   bool
}

type ChannelUseCase struct {
	repo IChannelRepo
}

func NewIChannelUseCase(x *ChannelUseCase) IChannelUseCase {
	return x
}

func NewChannelUseCase(repo IChannelRepo) *ChannelUseCase {
	return &ChannelUseCase{repo: repo}
}

func (x *ChannelUseCase) List(ctx context.Context, query ChannelQuery) ([]*Channel, uint32, error) {
	list, err := x.repo.List(ctx, query)
	if err != nil {
		return nil, 0, err
	}
	query.ListAll = true
	cnt, err := x.repo.Count(ctx, query)
	if err != nil {
		return nil, 0, err
	}

	return list, cnt, nil
}

func (x *ChannelUseCase) Get(ctx context.Context, id uint32) (*Channel, error) {
	return x.repo.Get(ctx, id)
}

func (x *ChannelUseCase) Create(ctx context.Context, channel *Channel) (*Channel, error) {
	return x.repo.Create(ctx, channel)
}

func (x *ChannelUseCase) Update(ctx context.Context, channel *Channel, opts ChannelUpdateOptions) error {
	return x.repo.Update(ctx, channel, opts)
}

func (x *ChannelUseCase) Delete(ctx context.Context, ids []uint32) error {
	return x.repo.Delete(ctx, ids)
}
