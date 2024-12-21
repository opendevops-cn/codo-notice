package biz

import (
	"context"
	"net/url"
	"time"

	"codo-notice/internal/conf"
)

type ITemplateUseCase interface {
	List(ctx context.Context, query TemplateQuery) ([]*Template, uint32, error)
	Get(ctx context.Context, id uint32) (*Template, error)
	Create(ctx context.Context, template *Template) (*Template, error)
	Update(ctx context.Context, template *Template, opts TemplateUpdateOptions) error
	Delete(ctx context.Context, ids []uint32) error
}

type ITemplateRepo interface {
	List(ctx context.Context, query TemplateQuery) ([]*Template, error)
	Count(ctx context.Context, query TemplateQuery) (uint32, error)
	Get(ctx context.Context, id uint32) (*Template, error)
	Create(ctx context.Context, template *Template) (*Template, error)
	Update(ctx context.Context, template *Template, opts TemplateUpdateOptions) error
	Delete(ctx context.Context, ids []uint32) error
}

// Template 模板定义
type Template struct {
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
	// 模板名称，唯一且不为空，最大长度256
	Name string
	// 模板内容
	Content string
	// 模板类型，不为空，默认值为"default"，最大长度16
	Type NotifyType
	// 模板用途，不为空，默认值为"default"，最大长度45
	Use string
	// 是否为默认模板，可选值："yes"/"no"，默认值为"no"，最大长度16
	Default string
	// path 路径信息
	Path string
}

func (x *Template) NotifyTemplatePath(gatewayPrefix string) string {
	values := make(url.Values)
	values.Set("type", string(x.Type))
	values.Set("tpl", x.Name)
	switch x.Type {
	case NotifyEmail:
		values.Set("email", "邮箱地址")
	case NotifyFeiShu, NotifyDingDing, NotifyQiYeWX, NotifyWebhook:
		values.Set("url", "机器人地址")
		values.Set("secret", "签名密钥")
	case NotifyAliyunDX, NotifyAliyunDH, NotifyTengXunDX, NotifyTengXunDH:
		values.Set("phone", "手机号")
	case NotifyFeiShuApp, NotifyDingDingApp, NotifyQiYeWXApp:
		values.Set("at", "域账号")
	}
	uri, _ := url.Parse(gatewayPrefix)
	uri = uri.JoinPath("/api/noc/v1/alert")
	// 暂时不用 encode 的结果, 为了前端的可读性
	// uri.RawQuery = values.Encode()
	uri.RawQuery, _ = url.QueryUnescape(values.Encode())
	return uri.String()
}

type TemplateQuery struct {
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

type TemplateUpdateOptions struct {
	// 模板名称，唯一且不为空，最大长度256
	Name bool
	// 模板内容
	Content bool
	// 模板类型，不为空，默认值为"default"，最大长度16
	Type bool
	// 模板用途，不为空，默认值为"default"，最大长度45
	Use bool
}

type TemplateUseCase struct {
	repo ITemplateRepo
	bc   *conf.Bootstrap
}

func NewITemplateUseCase(x *TemplateUseCase) ITemplateUseCase {
	return x
}

func NewTemplateUseCase(repo ITemplateRepo, bc *conf.Bootstrap) *TemplateUseCase {
	return &TemplateUseCase{repo: repo, bc: bc}
}

func (x *TemplateUseCase) List(ctx context.Context, query TemplateQuery) ([]*Template, uint32, error) {
	list, err := x.repo.List(ctx, query)
	if err != nil {
		return nil, 0, err
	}
	query.ListAll = true
	count, err := x.repo.Count(ctx, query)
	if err != nil {
		return nil, 0, err
	}

	for i := range list {
		list[i].Path = list[i].NotifyTemplatePath(x.bc.Metadata.GatewayPrefix)
	}

	return list, count, nil
}

func (x *TemplateUseCase) Get(ctx context.Context, id uint32) (*Template, error) {
	template, err := x.repo.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	template.Path = template.NotifyTemplatePath(x.bc.Metadata.GatewayPrefix)
	return template, nil
}

func (x *TemplateUseCase) Create(ctx context.Context, template *Template) (*Template, error) {
	return x.repo.Create(ctx, template)
}

func (x *TemplateUseCase) Update(ctx context.Context, template *Template, opts TemplateUpdateOptions) error {
	return x.repo.Update(ctx, template, opts)
}

func (x *TemplateUseCase) Delete(ctx context.Context, ids []uint32) error {
	return x.repo.Delete(ctx, ids)
}
