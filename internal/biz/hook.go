package biz

import (
	"context"
	nethttp "net/http"

	"codo-notice/internal/conf"
	"codo-notice/internal/ievents"

	"github.com/ccheers/xpkg/xmsgbus"
	larkcard "github.com/larksuite/oapi-sdk-go/v3/card"
	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	"github.com/larksuite/oapi-sdk-go/v3/core/httpserverext"
	larkevent "github.com/larksuite/oapi-sdk-go/v3/event"
	loggersdk "github.com/opendevops-cn/codo-golang-sdk/logger"
)

type IHookUseCase interface {
	HandleLarkCard(ctx context.Context, r *nethttp.Request, w nethttp.ResponseWriter) error
}

type HookUseCase struct {
	logCfg  *conf.Otel_Log
	hookCfg *conf.Hook

	pub xmsgbus.IPublisher[*ievents.LarkCardEvent]
}

func NewIHookUseCase(x *HookUseCase) IHookUseCase {
	return x
}

func NewHookUseCase(bc *conf.Bootstrap, iMsgBus xmsgbus.IMsgBus, iTopicManager xmsgbus.ITopicManager, otelOptions *xmsgbus.OTELOptions) *HookUseCase {
	return &HookUseCase{
		logCfg:  bc.Otel.Log,
		hookCfg: bc.Hook,
		pub:     xmsgbus.NewPublisher[*ievents.LarkCardEvent](iMsgBus, iTopicManager, otelOptions),
	}
}

func (x *HookUseCase) HandleLarkCard(ctx context.Context, r *nethttp.Request, w nethttp.ResponseWriter) error {
	// todo 这里考虑转换成使用 多个 APPID 配置的方式
	larkCardCfg := x.hookCfg.GetLarkCard()
	httpserverext.NewCardActionHandlerFunc(
		larkcard.NewCardActionHandler(
			larkCardCfg.VerificationToken, larkCardCfg.EncryptKey,
			func(ctx context.Context, action *larkcard.CardAction) (interface{}, error) {
				// 发送领域事件
				err := x.pub.Publish(ctx, &ievents.LarkCardEvent{
					OpenID:        action.OpenID,
					Token:         action.Token,
					OpenMessageID: action.OpenMessageID,
					OpenChatId:    action.OpenChatId,
					Action:        action.Action,
				})
				if err != nil {
					return nil, err
				}
				return nil, nil
			}),
		larkevent.WithLogLevel(x.convertLarkLogLevel()),
	)(w, r)
	return nil
}

func (x *HookUseCase) convertLarkLogLevel() larkcore.LogLevel {
	switch loggersdk.ParseLevel(x.logCfg.Level) {
	case loggersdk.LevelDebug:
		return larkcore.LogLevelDebug
	case loggersdk.LevelInfo:
		return larkcore.LogLevelInfo
	case loggersdk.LevelWarn:
		return larkcore.LogLevelWarn
	case loggersdk.LevelError:
		return larkcore.LogLevelError
	}

	return larkcore.LogLevelInfo
}
