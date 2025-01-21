package alerts

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"codo-notice/internal/biz"
	"codo-notice/internal/conf"
	"codo-notice/internal/ievents"

	"github.com/ccheers/xpkg/generic/arrayx"
	"github.com/ccheers/xpkg/sync/errgroup"
	"github.com/ccheers/xpkg/xmsgbus"
	"github.com/google/uuid"
	larkimsdk "github.com/larksuite/oapi-sdk-go/v3"
	larkcard "github.com/larksuite/oapi-sdk-go/v3/card"
	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	larkcontact "github.com/larksuite/oapi-sdk-go/v3/service/contact/v3"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	"github.com/opendevops-cn/codo-golang-sdk/client/xhttp"
	loggersdk "github.com/opendevops-cn/codo-golang-sdk/logger"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

var ErrFsAppInvalidToken = fmt.Errorf("invalid token")

type LarkCardCallbackInfo struct {
	WebhookUUID string
	Webhook     biz.AlertWebhook
	CardJSON    string
}

type ILarkCardCallbackRepo interface {
	Create(ctx context.Context, info *LarkCardCallbackInfo) error
	Get(ctx context.Context, webhookUUID string) (*LarkCardCallbackInfo, error)
}

type FeishuAppAlerter struct {
	logger *loggersdk.Helper

	// todo 这里后面重构去掉
	fsAppCfg *conf.NotifyConfig_FsApp

	repo   ILarkCardCallbackRepo
	render *TemplateRender

	client xhttp.IClient

	mu            sync.Mutex
	larkClientMap map[string]*larkimsdk.Client

	iMsgBus       xmsgbus.IMsgBus
	iTopicManager xmsgbus.ITopicManager
	otelOptions   *xmsgbus.OTELOptions
}

func NewFeishuAppAlerter(ctx context.Context, logger loggersdk.Logger, repo ILarkCardCallbackRepo,
	render *TemplateRender, client xhttp.IClient, bc *conf.Bootstrap,
	iMsgBus xmsgbus.IMsgBus, iTopicManager xmsgbus.ITopicManager, otelOptions *xmsgbus.OTELOptions,
) (*FeishuAppAlerter, func()) {
	x := &FeishuAppAlerter{
		logger:        loggersdk.NewHelper(logger),
		fsAppCfg:      bc.NotifyConfig.Fsapp,
		repo:          repo,
		render:        render,
		client:        client,
		mu:            sync.Mutex{},
		larkClientMap: make(map[string]*larkimsdk.Client),
		iMsgBus:       iMsgBus,
		iTopicManager: iTopicManager,
		otelOptions:   otelOptions,
	}

	// 同步监听任务
	ctx, cancel := context.WithCancel(ctx)
	eg := errgroup.WithCancel(ctx)
	eg.Go(func(ctx context.Context) error {
		return x.listenLarkCard(ctx)
	})
	eg.Go(func(ctx context.Context) error {
		return x.listenLarkApp(ctx)
	})

	return x, func() {
		cancel()
		_ = eg.Wait()
	}
}

func (x *FeishuAppAlerter) Alert(ctx context.Context, ac biz.AlertContext) error {
	if !ac.AlertNotifyData.FeiShuApp.IsValid() {
		return nil
	}

	if err := x.send2FeiShuApp(ctx, ac); err != nil {
		return err
	}
	return nil
}

func (x *FeishuAppAlerter) Type() biz.NotifyType {
	return biz.NotifyFeiShuApp
}

func (x *FeishuAppAlerter) send2FeiShuApp(ctx context.Context, ac biz.AlertContext) error {
	cfg := ac.AlertNotifyData.FeiShuApp

	larkClient, err := x.getLarkClient(ctx, cfg.AppID)
	if err != nil {
		return fmt.Errorf("[FeishuAppAlerter][send2FeiShuApp][getLarkClient] err=%w", err)
	}
	// get open id
	/*
		if err := xh.GetUserInfo(param.PhoneSet); err != nil { return err }
		var openIds []string
		for _, phone := range param.PhoneSet {
			if id, ok := xh.widget.OpenId[phone]; ok {
				openIds = append(openIds, id)
			}
		}
	*/

	jsonDoc := `{"msg_type":"interactive"}`
	// 设置通知用户
	jsonDoc, err = sjson.Set(jsonDoc, "user_ids", arrayx.Map(ac.CC, func(t *biz.User) string {
		return t.FsId
	}))
	if err != nil {
		return fmt.Errorf("[FeishuAppAlerter][send2FeiShuApp][sjson.Set] err=%w, path=open_ids", err)
	}

	// 构建卡片内容
	card, err := x.buildCardJSON(ctx, ac)
	if err != nil {
		return err
	}

	cardJsonStr, err := card.JSON()
	if err != nil {
		return fmt.Errorf("[FeishuAppAlerter][send2FeiShuApp][json.Marshal] err=%w", err)
	}
	jsonDoc, err = sjson.SetRaw(jsonDoc, "card", cardJsonStr)
	if err != nil {
		return fmt.Errorf("[FeishuAppAlerter][send2FeiShuApp][sjson.Set] err=%w, path=card", err)
	}

	// send

	var jsonBody map[string]interface{}
	_ = json.Unmarshal([]byte(jsonDoc), &jsonBody)
	resp, err := larkClient.Post(ctx, "/open-apis/message/v4/batch_send/", jsonBody, larkcore.AccessTokenTypeApp)
	if err != nil {
		return fmt.Errorf("[FeishuAppAlerter][send2FeiShuApp][NewRequest] err=%w, api=/message/v4/batch_send/, json=%s", err, jsonDoc)
	}
	type Resp struct {
		Code    int    `json:"code"`
		Message string `json:"msg"`
	}
	var dst Resp
	err = json.Unmarshal(resp.RawBody, &dst)
	if err != nil {
		return fmt.Errorf("[FeishuAppAlerter][send2FeiShuApp][Decode] err=%w", err)
	}

	if dst.Code != 0 {
		if dst.Code == 99991663 || dst.Code == 99991668 {
			// https://open.feishu.cn/document/uAjLw4CM/ugTN1YjL4UTN24CO1UjN/trouble-shooting/how-to-fix-99991663-error
			return fmt.Errorf("%w, msg=%s, code=%d", ErrFsAppInvalidToken, dst.Message, dst.Code)
		}
		return fmt.Errorf("[FeishuAppAlerter][send2FeiShuApp][SendFailed] code: %d, message: %s", dst.Code, dst.Message)
	}

	return nil
}

func (x *FeishuAppAlerter) buildCardJSON(ctx context.Context, ac biz.AlertContext) (*larkcard.MessageCard, error) {
	cfg := ac.AlertNotifyData.FeiShuApp
	// 配置内容
	text, err := x.render.Render(ctx, ac.Template.Content, ac.Args)
	if err != nil {
		return nil, fmt.Errorf("[FeishuAppAlerter][buildCardJSON][Render] err=%w, template=%s, args=%+v", err, ac.Template.Content, ac.Args)
	}
	// 头部信息
	cardHeader := larkcard.NewMessageCardHeader().
		// 标题主题颜色。支持 "blue"|"wathet"|"tuiquoise"|"green"|"yellow"|"orange"|"red"|"carmine"|"violet"|"purple"|"indigo"|"grey"|"default"。默认值 default。
		Template(getTitleColor(text, ac)).
		Title(
			larkcard.NewMessageCardPlainText().Content(ac.Title),
		).Build()

	// 正文
	cardContent := larkcard.NewMessageCardMarkdown().Content(text).Build()

	// 添加通知人信息
	noticer, manager := ac.CC, ac.Manager
	var noticerFields []*larkcard.MessageCardField
	if len(noticer) > 0 {
		noticerFields = append(noticerFields,
			larkcard.NewMessageCardField().
				IsShort(false).
				Text(
					larkcard.NewMessageCardLarkMd().
						Content(
							fmt.Sprintf("**通知人：**%s", strings.Join(arrayx.Map(noticer, func(t *biz.User) string {
								return t.Nickname
							}), ", ")),
						).
						Build(),
				).
				Build(),
		)
	}
	if len(manager) > 0 {
		noticerFields = append(noticerFields,
			larkcard.NewMessageCardField().
				IsShort(false).
				Text(
					larkcard.NewMessageCardLarkMd().
						Content(
							fmt.Sprintf("**负责人：**%s", strings.Join(arrayx.Map(manager, func(t *biz.User) string {
								return t.Nickname
							}), ", ")),
						).
						Build(),
				).
				Build(),
		)
	}

	// 回调信息
	var cardCallbackActions []larkcard.MessageCardActionElement
	uuidWebhookMapping := make(map[string]biz.AlertWebhook)
	if len(ac.AlertWebhooks) > 0 {
		// https://open.feishu.cn/document/server-docs/im-v1/message-card/overview
		for _, webhook := range ac.AlertWebhooks {
			webhookUUID := uuid.NewString()
			uuidWebhookMapping[webhookUUID] = webhook

			//  default：默认样式
			//	primary：强调样式
			//	danger：警示样式
			btnType := larkcard.MessageCardButtonTypeDefault
			if webhook.AlertCallbackMode.IsApprove() {
				btnType = larkcard.MessageCardButtonTypePrimary
			}
			if webhook.AlertCallbackMode.IsReject() {
				btnType = larkcard.MessageCardButtonTypeDanger
			}
			cardCallbackActions = append(cardCallbackActions,
				larkcard.NewMessageCardEmbedButton().
					Type(btnType).
					Text(larkcard.NewMessageCardLarkMd().Content(webhook.Alias).Build()).
					Value(map[string]interface{}{
						"uuid":          webhookUUID,
						"callback_args": ac.WebhookCallbackArgs,
						"app_id":        cfg.AppID,
					}).
					Build(),
			)
		}
	}

	// 整理 elements
	var cardElements []larkcard.MessageCardElement
	cardElements = append(cardElements, cardContent)
	if len(noticerFields) > 0 {
		cardElements = append(cardElements, larkcard.NewMessageCardHr().Build())
		cardElements = append(cardElements,
			larkcard.NewMessageCardDiv().
				Fields(noticerFields).
				Build(),
		)
	}
	if len(cardCallbackActions) > 0 {
		cardElements = append(cardElements, larkcard.NewMessageCardHr().Build())
		cardElements = append(cardElements,
			larkcard.NewMessageCardAction().
				Layout(larkcard.MessageCardActionLayoutFlow.Ptr()).
				Actions(cardCallbackActions).
				Build(),
		)
	}

	card := larkcard.NewMessageCard().
		Config(
			larkcard.NewMessageCardConfig().
				// 是否为共享卡片。为 true 时即更新卡片的内容对所有收到这张卡片的人员可见。默认值 false。
				UpdateMulti(true).
				WideScreenMode(true).
				Build(),
		).
		Header(cardHeader).
		Elements(cardElements).
		Build()

	// todo 将 URL 和 webhookUUID 存储到数据库中
	cardJSON, err := card.JSON()
	if err != nil {
		return nil, fmt.Errorf("[FeishuAppAlerter][buildCardJSON][JSON] err=%w", err)
	}
	for hookUUID, webhook := range uuidWebhookMapping {
		err := x.repo.Create(ctx, &LarkCardCallbackInfo{
			WebhookUUID: hookUUID,
			Webhook:     webhook,
			CardJSON:    cardJSON,
		})
		if err != nil {
			return nil, fmt.Errorf("[FeishuAppAlerter][buildCardJSON][Create] err=%w", err)
		}
	}

	return card, nil
}

func (x *FeishuAppAlerter) listenLarkCard(ctx context.Context) error {
	sub := xmsgbus.NewSubscriber[*ievents.LarkCardEvent](
		(*ievents.LarkCardEvent)(nil).Topic(),
		"FeishuWebhookAlerter",
		x.iMsgBus,
		x.otelOptions,
		x.iTopicManager,
		xmsgbus.WithCheckFunc(func(ctx context.Context, dst *ievents.LarkCardEvent) bool {
			if dst.Action == nil {
				return false
			}
			_, ok := dst.Action.Value["uuid"]
			return ok
		}),
		xmsgbus.WithHandleFunc(func(ctx context.Context, dst *ievents.LarkCardEvent) error {
			// 业务处理
			// 根据UUID 获取对应的告警信息
			webhookUUID, ok := dst.Action.Value["uuid"].(string)
			if !ok {
				x.logger.Errorf(ctx, "[FeishuWebhookAlerter][listenLarkCard][Handle] uuid not found")
				return nil
			}
			appID, ok := dst.Action.Value["app_id"].(string)
			if !ok {
				x.logger.Errorf(ctx, "[FeishuWebhookAlerter][listenLarkCard][Handle] app_id not found")
				return nil
			}

			callbackArgs, _ := dst.Action.Value["callback_args"].(string)

			// todo 这里拿 client 有问题 后续逻辑应该是:
			// 1. 用户在后台配置飞书账号
			// 2. 持续同步飞书账号, 加载账号配置成 client
			// 3. 根据 appid 即可获取 client
			// 获取飞书用户ID
			x.mu.Lock()
			larkClient := x.larkClientMap[appID]
			x.mu.Unlock()
			if larkClient == nil {
				x.logger.Errorf(ctx, "[FeishuWebhookAlerter][listenLarkCard][Handle] lark client not found")
				return nil
			}
			openID := dst.OpenID
			usrInfo, err := larkClient.Contact.User.Get(ctx, larkcontact.NewGetUserReqBuilder().UserIdType("open_id").UserId(openID).Build())
			if err != nil {
				return fmt.Errorf("[FeishuWebhookAlerter][listenLarkCard][Handle][GetUser] err=%w, open_id=%s", err, openID)
			}
			if !usrInfo.Success() {
				return fmt.Errorf("[FeishuWebhookAlerter][listenLarkCard][Handle][GetUser] code=%d, msg=%s, log_id=%s", usrInfo.Code, usrInfo.Msg, usrInfo.LogId())
			}

			// 通过告警信息对应的 webhook, 进行回调
			webHookInfo, err := x.repo.Get(ctx, webhookUUID)
			if err != nil {
				return fmt.Errorf("[FeishuWebhookAlerter][listenLarkCard][Handle][Get] err=%w, uuid=%s", err, webhookUUID)
			}

			userEmail := ""
			if usrInfo.Data.User.Email != nil {
				userEmail = *usrInfo.Data.User.Email
			}
			if usrInfo.Data.User.EnterpriseEmail != nil {
				userEmail = *usrInfo.Data.User.EnterpriseEmail
			}
			bs, _ := json.Marshal(biz.WebhookRequest{
				UserEmail:    userEmail,
				CallbackArgs: callbackArgs,
			})

			req, err := http.NewRequest(http.MethodPost, webHookInfo.Webhook.URL, bytes.NewReader(bs))
			if err != nil {
				return fmt.Errorf("[FeishuWebhookAlerter][listenLarkCard][Handle][NewRequest] err=%w, url=%s", err, webHookInfo.Webhook.URL)
			}
			resp, err := x.client.Do(ctx, req)
			if err != nil {
				return fmt.Errorf("[FeishuWebhookAlerter][listenLarkCard][Handle][Do] err=%w", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				buf := make([]byte, 4096)
				n, _ := resp.Body.Read(buf)
				return fmt.Errorf("[FeishuWebhookAlerter][listenLarkCard][Handle][StatusCode] code=%d, body=%s", resp.StatusCode, string(buf[:n]))
			}

			// 解析 proc result
			var webhookResp biz.WebhookResponse
			err = json.NewDecoder(resp.Body).Decode(&webhookResp)
			if err != nil {
				return fmt.Errorf("[FeishuWebhookAlerter][listenLarkCard][Handle][Decode] err=%w, url=%s", err, webHookInfo.Webhook.URL)
			}

			procResult := fmt.Sprintf("%s[已处理]", webHookInfo.Webhook.Alias)
			if webhookResp.Result.HookResult != "" {
				procResult = webhookResp.Result.HookResult
			}

			// 重新渲染 card
			cardJSON := webHookInfo.CardJSON
			renderJSON := cardJSON
			gjson.Parse(cardJSON).Get("elements").ForEach(func(elementIndex, element gjson.Result) bool {
				if element.Get("tag").String() == "action" {
					element.Get("actions").ForEach(func(actionIndex, action gjson.Result) bool {
						// 构建正确的路径
						basePath := fmt.Sprintf("elements.%d.actions.%d", elementIndex.Int(), actionIndex.Int())

						if action.Get("value.uuid").String() == webhookUUID {
							renderJSON, _ = sjson.Set(renderJSON, basePath+".text.content", procResult)
							renderJSON, _ = sjson.Set(renderJSON, basePath+".disabled", true)
							renderJSON, _ = sjson.SetRaw(renderJSON, basePath+".value", "{}")
						} else {
							renderJSON, _ = sjson.Set(renderJSON, basePath+".disabled", true)
							renderJSON, _ = sjson.SetRaw(renderJSON, basePath+".value", "{}")
						}
						return true
					})
				}
				return true
			})

			// 修改告警按钮内容
			patchResp, err := larkClient.Im.Message.Patch(ctx, larkim.NewPatchMessageReqBuilder().
				MessageId(dst.OpenMessageID).
				Body(larkim.NewPatchMessageReqBodyBuilder().Content(renderJSON).Build()).
				Build())
			if err != nil {
				return fmt.Errorf("[FeishuWebhookAlerter][listenLarkCard][Handle][Patch] err=%w, open_message_id=%s", err, dst.OpenMessageID)
			}

			if !patchResp.Success() {
				return fmt.Errorf("[FeishuWebhookAlerter][listenLarkCard][Handle][Patch] code=%d, msg=%s, log_id=%s", patchResp.Code, patchResp.Msg, patchResp.LogId())
			}
			return nil
		}),
	)

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}
		err := sub.Handle(ctx)
		if err != nil && !errors.Is(err, context.DeadlineExceeded) {
			x.logger.Errorf(ctx, "[FeishuWebhookAlerter][listenLarkCard][Handle] err=%v", err)
		}
	}
}

func (x *FeishuAppAlerter) listenLarkApp(ctx context.Context) error {
	// todo 后续从存储中获得 appid 和 appsecret
	// 现在暂时从配置获取
	mm := make(map[string]*larkimsdk.Client)

	appID := x.fsAppCfg.AppId
	appSecret := x.fsAppCfg.AppSecret
	larkClient := x.buildLarkClient(ctx, appID, appSecret)

	// 因为从配置获取, 所以只有一个
	mm[appID] = larkClient
	x.mu.Lock()
	x.larkClientMap = mm
	x.mu.Unlock()
	return nil
}

type larkHttpWrapper struct {
	client xhttp.IClient
}

func (x *larkHttpWrapper) Do(request *http.Request) (*http.Response, error) {
	return x.client.Do(request.Context(), request)
}

func (x *FeishuAppAlerter) getLarkClient(_ context.Context, appID string) (*larkimsdk.Client, error) {
	x.mu.Lock()
	defer x.mu.Unlock()
	larkClient := x.larkClientMap[appID]
	if larkClient == nil {
		return nil, fmt.Errorf("[FeishuAppAlerter][getLarkClient] appID=%s not found", appID)
	}
	return larkClient, nil
}

func (x *FeishuAppAlerter) buildLarkClient(_ context.Context, appID, appSecret string) *larkimsdk.Client {
	return larkimsdk.NewClient(
		appID, appSecret,
		larkimsdk.WithEnableTokenCache(true),
		larkimsdk.WithHttpClient(&larkHttpWrapper{x.client}),
	)
}
