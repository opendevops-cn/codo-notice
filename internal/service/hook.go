package service

import (
	"net/http"

	"codo-notice/internal/biz"

	khttp "github.com/go-kratos/kratos/v2/transport/http"
)

type RouteInfo struct {
	Method string
	Path   string
	Handle khttp.HandlerFunc
}

type HookService struct {
	uc biz.IHookUseCase
}

func NewHookService(uc biz.IHookUseCase) *HookService {
	return &HookService{uc: uc}
}

func (x *HookService) Routes() []*RouteInfo {
	return []*RouteInfo{
		x.LarkCardHook(),
	}
}

func (x *HookService) LarkCardHook() *RouteInfo {
	return &RouteInfo{
		Method: http.MethodPost,
		Path:   "/hook/v1/lark-card-hook",
		Handle: func(ctx khttp.Context) error {
			return x.uc.HandleLarkCard(ctx, ctx.Request(), ctx.Response())
		},
	}
}
