package biz

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"codo-notice/internal/conf"
	"codo-notice/internal/pkg/istr"

	"github.com/ccheers/xpkg/generic/arrayx"
	"github.com/ccheers/xpkg/sync/errgroup"
	loggersdk "github.com/opendevops-cn/codo-golang-sdk/logger"
	"go.uber.org/multierr"
)

var ErrNoRouterMatch = fmt.Errorf("no routers was matched")

// 相关保留字段
const (
	FieldsMessage     = "message"      // 通用消息字段 (从 jsonBody 取值)
	FieldsManager     = "codo_manager" // 业务负责人 (从 httpQuery + jsonBody 取值)
	FieldsNoticer     = "codo_noticer" // 内置通知人 (从 httpQuery + jsonBody 取值)
	FieldsSeverity    = "severity"     // 告警等级 [fatal | error | warn | info] (从 httpQuery + jsonBody 取值)
	FieldsTitle       = "title"        // 告警标题 (从 httpQuery + jsonBody 取值)
	FieldsNativeTitle = "codo_title"   // 原生标题 不做其他处理 (从 httpQuery + jsonBody 取值)
	FieldsStatus      = "status"       // 告警状态 [resolved | firing | padding] (从 httpQuery + jsonBody 取值)
	FieldsAppCn       = "cmdb_bizcn"   // 业务中文描述 (从 httpQuery + jsonBody 取值)
)

type AlertInfo struct {
	// (必填) Labels 告警标签
	Labels map[string]string
	// (必填) RawData 原始数据
	RawData map[string]interface{}
	// (必填) TrigUser 告警触发人
	TrigUser *User
	// (必填) Entrance 告警来源
	Entrance AlertEntrance
	// (可选) CCUsers 抄送通知的用户
	CCUsers []*User
	// (可选) Template 通知模版
	Template *Template
	// todo 这个 Extra 暂时这样抽象, 后续可以结合通知点做一次真正的重构(将通知配置挂在通知点上)
	// (可选) ExtraDataNotifyURLs 通知额外数据 NotifyType 为 (NotifyFeiShu | NotifyDingDing | NotifyQiYeWX | NotifyWebhook) 存在
	ExtraDataNotifyURLs map[string]string
	// (可选) 通知回调地址
	AlertWebhooks []AlertWebhook
}

// IAlertUseCase 管理 Alerter 根据 NotifyType 进行路由告警
type IAlertUseCase interface {
	// Alert 直接通知
	Alert(ctx context.Context, info *AlertInfo) error
	// RouteAlert 路由通知 (Template 为 nil, 根据路由规则动态获取)
	RouteAlert(ctx context.Context, info *AlertInfo) error
}

type IAlerterList []IAlerter

// IAlerter 告警器
type IAlerter interface {
	Alert(ctx context.Context, ac AlertContext) error
	Type() NotifyType
}

// AlertStatus 告警状态
type AlertStatus string

const (
	// AlertStatusResolved 已恢复
	AlertStatusResolved = AlertStatus("resolved")
	// AlertStatusFiring 告警发生
	AlertStatusFiring = AlertStatus("firing")
	// AlertStatusInactive 告警失效
	AlertStatusInactive = AlertStatus("inactive")
	// AlertStatusResolving 告警恢复中
	AlertStatusResolving = AlertStatus("resolving")
)

func (x AlertStatus) IsResolved() bool {
	return x == AlertStatusResolved
}

func (x AlertStatus) IsResolving() bool {
	return x == AlertStatusResolving
}

func (x AlertStatus) IsFiring() bool {
	return x == AlertStatusFiring
}

type AlertSeverity string

const (
	// AlertSeverityFatal 致命
	AlertSeverityFatal = AlertSeverity("fatal")
	// AlertSeverityError 严重
	AlertSeverityError = AlertSeverity("error")
	// AlertSeverityWarn 警告
	AlertSeverityWarn = AlertSeverity("warn")
	// AlertSeverityInfo 普通
	AlertSeverityInfo = AlertSeverity("info")
)

type AlertEntrance string

const (
	AlertEntranceTmpl   = AlertEntrance("template")
	AlertEntranceRouter = AlertEntrance("router")
)

type WebhookRequest struct {
	UserEmail string `json:"user_email"`
}

type WebhookResponse struct {
	Code int `json:"code"`
	// 开发看
	Msg string `json:"msg"`
	// 用户看
	Reason string `json:"reason"`
	// 服务器时间戳
	Timestamp uint32 `json:"timestamp"`
	// 结构化数据
	Result struct {
		HookResult string `json:"hook_result"`
	} `json:"result"`
}

type AlertWebhook struct {
	URL   string
	Alias string
	// 同意 or 拒绝
	IsApprove bool
	IsReject  bool
}

type AlertContext struct {
	// Status 告警状态, 只有在 AlertStatusFiring 时, AlertSeverity 才有意义
	Status AlertStatus
	// Severity 告警等级
	Severity AlertSeverity
	// Title 告警标题
	Title string
	// Message 告警消息
	Message string
	// Args 上下文参数
	Args map[string]string
	// TrigUser 告警触发人
	TrigUser *User
	// Entrance 告警来源
	Entrance AlertEntrance
	// 抄送人
	CC []*User
	// 负责人
	Manager []*User

	// webhook
	AlertWebhooks []AlertWebhook
	// Template 通知模版
	Template *Template
	// NotifyType 通知类型, 数据与通知模板内的通知类型一致
	NotifyType NotifyType
	// AlertNotifyData 通知端点数据
	AlertNotifyData AlertNotifyData
}

type AlertNotifyData struct {
	Email       AlertNotifyDataEmail
	FeiShu      AlertNotifyDataFeiShu
	FeiShuApp   AlertNotifyDataFeiShuApp
	AliyunDX    AlertNotifyDataAliyunDX
	AliyunDH    AlertNotifyDataAliyunDH
	TengXunDX   AlertNotifyDataTengXunDX
	TengXunDH   AlertNotifyDataTengXunDH
	DingDing    AlertNotifyDataDingDing
	DingDingApp AlertNotifyDataDingDingApp
	QiYeWX      AlertNotifyDataQiYeWX
	QiYeWXApp   AlertNotifyDataQiYeWXApp
	Webhook     AlertNotifyDataWebhook
}

type AlertNotifyDataEmail struct {
	// 邮件服务器配置
	Host     string
	Port     int
	User     string
	Password string
}

func (x *AlertNotifyDataEmail) IsValid() bool {
	return x.Host != "" && x.Port != 0 && x.User != "" && x.Password != ""
}

type AlertNotifyDataFeiShu struct {
	// URL 和 Secret 集合
	URLSet map[string]string
}

func (x *AlertNotifyDataFeiShu) IsValid() bool {
	return len(x.URLSet) > 0
}

type AlertNotifyDataFeiShuApp struct {
	// 通知账号密码配置
	AppID     string
	AppSecret string
}

func (x *AlertNotifyDataFeiShuApp) IsValid() bool {
	return x.AppID != "" && x.AppSecret != ""
}

type AlertNotifyDataAliyunDX struct {
	// 短信告警配置
	AccessId     string
	AccessSecret string
	SignName     string
	Template     string
}

func (x *AlertNotifyDataAliyunDX) IsValid() bool {
	return x.AccessId != "" && x.AccessSecret != ""
}

type AlertNotifyDataAliyunDH struct {
	// 电话告警配置
	AccessId         string
	AccessSecret     string
	TtsCode          string
	CalledShowNumber string
}

func (x *AlertNotifyDataAliyunDH) IsValid() bool {
	return x.AccessId != "" && x.AccessSecret != ""
}

type AlertNotifyDataTengXunDX struct {
	// 短信告警配置
	AccessId     string
	AccessSecret string
	SignName     string
	Template     string
	AppId        string
}

func (x *AlertNotifyDataTengXunDX) IsValid() bool {
	return x.AccessId != "" && x.AccessSecret != ""
}

type AlertNotifyDataTengXunDH struct {
	// 电话告警配置
	AccessId     string
	AccessSecret string
	Template     string
	AppId        string
}

func (x *AlertNotifyDataTengXunDH) IsValid() bool {
	return x.AccessId != "" && x.AccessSecret != ""
}

type AlertNotifyDataDingDing struct {
	// URL 和 Secret 集合
	URLSet map[string]string
}

func (x *AlertNotifyDataDingDing) IsValid() bool {
	return len(x.URLSet) > 0
}

type AlertNotifyDataDingDingApp struct {
	// 钉钉告警配置
	AppId     string
	AppSecret string
	AgentId   string
}

func (x *AlertNotifyDataDingDingApp) IsValid() bool {
	return x.AppId != "" && x.AppSecret != ""
}

type AlertNotifyDataQiYeWX struct {
	// URL 和 Secret 集合
	URLSet map[string]string
}

func (x *AlertNotifyDataQiYeWX) IsValid() bool {
	return len(x.URLSet) > 0
}

type AlertNotifyDataQiYeWXApp struct {
	// 企业微信告警配置
	AgentId     int64  `mapstructure:"agentId"`
	AgentSecret string `mapstructure:"agentSecret"`
	CropId      string `mapstructure:"cropId"`
}

func (x *AlertNotifyDataQiYeWXApp) IsValid() bool {
	return x.AgentId != 0 && x.AgentSecret != "" && x.CropId != ""
}

type AlertNotifyDataWebhook struct {
	// webhook URL 和 Secret 集合
	URLSet map[string]string
}

func (x *AlertNotifyDataWebhook) IsValid() bool {
	return len(x.URLSet) > 0
}

type AlertUseCase struct {
	logger *loggersdk.Helper

	userUC     IUserUseCase
	routerUC   IRouterUseCase
	channelUC  IChannelUseCase
	templateUC ITemplateUseCase

	nc *conf.NotifyConfig

	alerterMap map[NotifyType]IAlerter

	mu sync.Mutex

	routers RouterList
}

func NewIAlertUseCase(x *AlertUseCase) IAlertUseCase {
	return x
}

func NewAlertUseCase(userUC IUserUseCase, bc *conf.Bootstrap, list IAlerterList, logger loggersdk.Logger,
	routerUC IRouterUseCase,
	channelUC IChannelUseCase,
	templateUC ITemplateUseCase,
) (*AlertUseCase, func()) {
	x := &AlertUseCase{
		logger:     loggersdk.NewHelper(logger),
		userUC:     userUC,
		routerUC:   routerUC,
		channelUC:  channelUC,
		templateUC: templateUC,
		nc:         bc.NotifyConfig,
		alerterMap: arrayx.BuildMap(list, func(t IAlerter) NotifyType {
			return t.Type()
		}),
		mu:      sync.Mutex{},
		routers: nil,
	}

	ctx, cancelCause := context.WithCancelCause(context.Background())

	eg := errgroup.WithCancel(ctx)

	eg.Go(func(ctx context.Context) error {
		for {
			if err := x.refreshRouters(ctx); err != nil {
				x.logger.Errorf(ctx, "[AlertUseCase][refreshRouters] refresh routers failed: %s", err)
			}
			select {
			case <-ctx.Done():
				return nil
			case <-time.After(time.Second * 10):
			}
		}
	})

	return x, func() {
		cancelCause(fmt.Errorf("cleanup"))
		eg.Wait()
	}
}

func (x *AlertUseCase) Alert(ctx context.Context, info *AlertInfo) error {
	alertCtx, err := x.buildAlertContext(ctx, info)
	if err != nil {
		return err
	}
	return x.alert(ctx, *alertCtx)
}

func (x *AlertUseCase) RouteAlert(ctx context.Context, info *AlertInfo) error {
	routers, err := x.matchRouter(ctx, info)
	if err != nil {
		return err
	}

	var errs []error
	for _, router := range routers {
		channel, err := x.channelUC.Get(ctx, router.ChannelID)
		if err != nil {
			errs = append(errs, fmt.Errorf("get channel failed: %w, channel_id=%d", err, router.ChannelID))
			continue
		}

		ccUsers, _, err := x.userUC.List(ctx, UserQuery{
			PageSize: 999,
			PageNum:  1,
			FilterMap: map[string]interface{}{
				"user_id": channel.User,
			},
		})
		if err != nil {
			errs = append(errs, fmt.Errorf("get cc users failed: %w, channel_id=%d, users=%+v", err, router.ChannelID, channel.User))
			continue
		}
		if len(ccUsers) == 0 {
			x.logger.Warnf(ctx, "[AlertUseCase][RouteAlert] ccUsers is empty, channel_id=%d", router.ChannelID)
			continue
		}

		// 整理 points
		points := make([]*ContactPoint, 0, len(channel.ContactPoints)+len(channel.CustomItems))
		for _, point := range channel.ContactPoints {
			points = append(points, point)
		}
		for _, point := range channel.CustomItems {
			points = append(points, point)
		}

		for _, point := range points {
			tmpl, err := x.templateUC.Get(ctx, point.TemplateId)
			if err != nil {
				errs = append(errs, fmt.Errorf("get template failed: %w, template_id=%d", err, point.TemplateId))
				continue
			}
			// 配置模板
			info.Template = tmpl
			// 配置抄送
			info.CCUsers = ccUsers
			// 配置回调
			info.AlertWebhooks = point.AlertWebhooks
			// todo 这里后续整体优化成 通知点配置 转化成 AlerterContext
			if point.Addr != "" {
				info.ExtraDataNotifyURLs = map[string]string{
					point.Addr: point.Secret,
				}
			}
			alertCtx, err := x.buildAlertContext(ctx, info)
			if err != nil {
				errs = append(errs, fmt.Errorf("build alert context failed: %w", err))
				continue
			}

			// 告警等级过滤
			if !arrayx.ContainsAny(point.Severity, []string{string(alertCtx.Severity)}) {
				x.logger.Debug(ctx, "[AlertUseCase][RouteAlert] severity not match", "point", point, "alert", alertCtx)
				continue
			}

			// 告警
			if err := x.alert(ctx, *alertCtx); err != nil {
				errs = append(errs, fmt.Errorf("[AlertUseCase][RouteAlert] alert failed: %w", err))
				continue
			}
		}
	}

	if len(errs) > 0 {
		return multierr.Combine(errs...)
	}
	return nil
}

func (x *AlertUseCase) matchRouter(ctx context.Context, info *AlertInfo) ([]*Router, error) {
	x.mu.Lock()
	defer x.mu.Unlock()

	var routers []*Router
	isMatched := false
	for _, router := range x.routers {
		matched, err := router.ConditionList.MatchLabels(info.Labels)
		if err != nil {
			x.logger.Errorf(ctx, "[AlertUseCase][matchRouter] match router failed: %s", err)
		}
		// todo 这里可以优化成批量匹配批量发, 跟着需求走, 目前只要单个即可
		if matched {
			routers = append(routers, router)
			isMatched = true
			break
		}
	}

	if !isMatched {
		return nil, ErrNoRouterMatch
	}
	return routers, nil
}

func (x *AlertUseCase) refreshRouters(ctx context.Context) error {
	defer func() {
		if err := recover(); err != nil {
			x.logger.Errorf(ctx, "[AlertUseCase][refreshRouters] panic: %v", err)
		}
	}()
	var routers RouterList
	for i := 1; true; i++ {
		list, _, err := x.routerUC.List(ctx, RouterQuery{
			PageSize: 100,
			PageNum:  int32(i),
		})
		if err != nil {
			return err
		}
		if len(list) == 0 {
			break
		}
		routers = append(routers, list...)
	}
	x.mu.Lock()
	defer x.mu.Unlock()
	x.routers = routers
	return nil
}

func (x *AlertUseCase) alert(ctx context.Context, alertCtx AlertContext) error {
	alerter, err := x.getAlerter(alertCtx.NotifyType)
	if err != nil {
		return err
	}
	return alerter.Alert(ctx, alertCtx)
}

func (x *AlertUseCase) getAlerter(notifyType NotifyType) (IAlerter, error) {
	alerter, ok := x.alerterMap[notifyType]
	if !ok {
		return nil, fmt.Errorf("not found alerter for notify type: %s", notifyType)
	}
	return alerter, nil
}

func (x *AlertUseCase) buildAlertContext(ctx context.Context, info *AlertInfo) (*AlertContext, error) {
	// 构造 label & data
	labels := info.Labels
	rawData := info.RawData
	data := make(map[string]string, len(rawData))
	for k, v := range rawData {
		val := strings.TrimSpace(istr.GetString(v))
		if val == "" {
			continue
		}
		data[k] = val
		labels[k] = val
	}

	// 获取触发人
	trigUsr := info.TrigUser

	// 构造 managers
	managers, err := x.buildManagers(ctx, labels, rawData)
	if err != nil {
		return nil, err
	}

	// 构造 managers
	noticers, err := x.buildNoticers(ctx, labels, rawData)
	if err != nil {
		return nil, err
	}

	if len(noticers) == 0 {
		noticers = info.CCUsers
	}

	// 构造 notify data
	notifyData, err := x.buildAlertNotifyData(ctx, info)
	if err != nil {
		return nil, err
	}

	return &AlertContext{
		Status:          x.getStatus(labels),
		Severity:        x.getSeverity(labels),
		Title:           x.getTitle(labels),
		Message:         data[FieldsMessage],
		Args:            data,
		TrigUser:        trigUsr,
		Entrance:        info.Entrance,
		CC:              noticers,
		Manager:         managers,
		AlertWebhooks:   info.AlertWebhooks,
		Template:        info.Template,
		NotifyType:      info.Template.Type,
		AlertNotifyData: *notifyData,
	}, nil
}

// 获取告警级别
func (x *AlertUseCase) getSeverity(labels map[string]string) AlertSeverity {
	tag := FieldsSeverity
	// from labels (query & body)
	buildInGrade := []AlertSeverity{AlertSeverityFatal, AlertSeverityError, AlertSeverityWarn, AlertSeverityInfo}
	v := labels[tag]
	severity := AlertSeverity(v)
	if arrayx.ContainsAny(buildInGrade, []AlertSeverity{severity}) {
		return severity
	}
	// 如果没指定默认为 warn
	return AlertSeverityWarn
}

// getTitle 获取标题
func (x *AlertUseCase) getTitle(labels map[string]string) string {
	title := "Notification"
	if v, ok := labels[FieldsNativeTitle]; ok {
		// 原生标题
		return v
	}
	if v, ok := labels[FieldsTitle]; ok {
		// 存在 title
		title = v
		if v, ok := labels[FieldsAppCn]; ok {
			// 存在业务中文名
			title = fmt.Sprintf("%s - %s", v, title)
		}
	}
	if v, ok := labels[FieldsStatus]; ok && (AlertStatus(v)).IsResolved() {
		// 当前为恢复告警
		title = fmt.Sprintf("【告警恢复】%s", title)
	} else {
		// 当前为告警
		if v, ok := labels[FieldsSeverity]; ok {
			title = fmt.Sprintf("【%s】%s", strings.ToUpper(v), title)
		}
	}
	return title
}

// getStatus 获取状态
func (x *AlertUseCase) getStatus(labels map[string]string) AlertStatus {
	if v, ok := labels[FieldsStatus]; ok {
		return AlertStatus(v)
	}
	text := labels[FieldsMessage]
	firingCnt := strings.Count(text, "firing")
	resolvedCnt := strings.Count(text, "resolved")
	if firingCnt > 0 && resolvedCnt > 0 {
		return AlertStatusResolving
	}
	if firingCnt > 0 {
		return AlertStatusFiring
	}
	if resolvedCnt > 0 {
		return AlertStatusResolved
	}
	return AlertStatusFiring
}

func (x *AlertUseCase) buildManagers(ctx context.Context, labels map[string]string, data map[string]interface{}) ([]*User, error) {
	tag := FieldsManager
	delimiter := ","
	// from query
	result := make([]string, 0)
	if v, ok := labels[tag]; ok {
		// 逗号分隔
		result = append(result, strings.Split(v, delimiter)...)
	}
	// from body
	if v, ok := data[tag]; ok {
		if val, ok := v.(string); ok {
			result = append(result, strings.Split(val, delimiter)...)
		} else if val, ok := v.([]string); ok {
			result = append(result, val...)
		}
	}

	result = arrayx.UniqArray(arrayx.Filter(arrayx.Map(result, func(t string) string {
		return strings.TrimSpace(t)
	}), func(s string) bool {
		return s != ""
	}))
	if len(result) == 0 {
		return nil, nil
	}

	managers, _, err := x.userUC.List(ctx, UserQuery{
		ListAll:  false,
		PageSize: 999,
		PageNum:  1,
		FilterMap: map[string]interface{}{
			"username": result,
			"disable":  false,
		},
	})
	if err != nil {
		return nil, err
	}
	return managers, nil
}

func (x *AlertUseCase) buildNoticers(ctx context.Context, labels map[string]string, data map[string]interface{}) ([]*User, error) {
	tag := FieldsNoticer
	delimiter := ","
	// from query
	result := make([]string, 0)
	if v, ok := labels[tag]; ok {
		// 逗号分隔
		result = append(result, strings.Split(v, delimiter)...)
	}
	// from body
	if v, ok := data[tag]; ok {
		if val, ok := v.(string); ok {
			result = append(result, strings.Split(val, delimiter)...)
		} else if val, ok := v.([]string); ok {
			result = append(result, val...)
		}
	}

	result = arrayx.UniqArray(arrayx.Filter(arrayx.Map(result, func(t string) string {
		return strings.TrimSpace(t)
	}), func(s string) bool {
		return s != ""
	}))
	if len(result) == 0 {
		return nil, nil
	}

	users, _, err := x.userUC.List(ctx, UserQuery{
		ListAll:  false,
		PageSize: 999,
		PageNum:  1,
		FilterMap: map[string]interface{}{
			"username": result,
			"disable":  false,
		},
	})
	if err != nil {
		return nil, err
	}
	return users, nil
}

func (x *AlertUseCase) buildAlertNotifyData(ctx context.Context, info *AlertInfo) (*AlertNotifyData, error) {
	switch info.Template.Type {
	case NotifyEmail:
		ec := x.nc.GetEmail()
		return &AlertNotifyData{
			Email: AlertNotifyDataEmail{
				Host:     ec.Host,
				Port:     int(ec.Port),
				User:     ec.User,
				Password: ec.Password,
			},
		}, nil
	case NotifyFeiShu:
		return &AlertNotifyData{
			FeiShu: AlertNotifyDataFeiShu{
				URLSet: info.ExtraDataNotifyURLs,
			},
		}, nil

	case NotifyFeiShuApp:
		fc := x.nc.GetFsapp()
		return &AlertNotifyData{
			FeiShuApp: AlertNotifyDataFeiShuApp{
				AppID:     fc.AppId,
				AppSecret: fc.AppSecret,
			},
		}, nil
	case NotifyAliyunDX:
		ac := x.nc.GetAliyun()
		return &AlertNotifyData{
			AliyunDX: AlertNotifyDataAliyunDX{
				AccessId:     ac.DxAccessId,
				AccessSecret: ac.DxAccessSecret,
				SignName:     ac.DxSignName,
				Template:     ac.DxTemplate,
			},
		}, nil

	case NotifyAliyunDH:
		ac := x.nc.GetAliyun()
		return &AlertNotifyData{
			AliyunDH: AlertNotifyDataAliyunDH{
				AccessId:         ac.DhAccessId,
				AccessSecret:     ac.DhAccessSecret,
				TtsCode:          ac.DhTtsCode,
				CalledShowNumber: ac.DhCalledShowNumber,
			},
		}, nil
	case NotifyTengXunDX:
		tc := x.nc.GetTxyun()
		return &AlertNotifyData{
			TengXunDX: AlertNotifyDataTengXunDX{
				AccessId:     tc.DxAccessId,
				AccessSecret: tc.DxAccessSecret,
				SignName:     tc.DxSignName,
				Template:     tc.DxTemplate,
				AppId:        tc.DxAppId,
			},
		}, nil
	case NotifyTengXunDH:
		tc := x.nc.GetTxyun()
		return &AlertNotifyData{
			TengXunDH: AlertNotifyDataTengXunDH{
				AccessId:     tc.DhAccessId,
				AccessSecret: tc.DhAccessSecret,
				Template:     tc.DhTemplate,
				AppId:        tc.DhAppId,
			},
		}, nil

	case NotifyDingDing:
		return &AlertNotifyData{
			DingDing: AlertNotifyDataDingDing{
				URLSet: info.ExtraDataNotifyURLs,
			},
		}, nil
	case NotifyDingDingApp:
		dc := x.nc.GetDdapp()
		return &AlertNotifyData{
			DingDingApp: AlertNotifyDataDingDingApp{
				AppId:     dc.AppId,
				AppSecret: dc.AppSecret,
				AgentId:   dc.AgentId,
			},
		}, nil
	case NotifyQiYeWX:
		return &AlertNotifyData{
			QiYeWX: AlertNotifyDataQiYeWX{
				URLSet: info.ExtraDataNotifyURLs,
			},
		}, nil
	case NotifyQiYeWXApp:
		wc := x.nc.GetWxapp()
		return &AlertNotifyData{
			QiYeWXApp: AlertNotifyDataQiYeWXApp{
				AgentId:     wc.AgentId,
				AgentSecret: wc.AgentSecret,
				CropId:      wc.CropId,
			},
		}, nil
	case NotifyWebhook:
		return &AlertNotifyData{
			Webhook: AlertNotifyDataWebhook{
				URLSet: info.ExtraDataNotifyURLs,
			},
		}, nil
	}

	return nil, fmt.Errorf("not support notify type: %s", info.Template.Type)
}
