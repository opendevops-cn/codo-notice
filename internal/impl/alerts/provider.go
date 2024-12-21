package alerts

import (
	"codo-notice/internal/biz"

	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(
	NewAlerterList,
	NewTemplateRender,

	// alerter list
	NewAliyunDianhuaAlerter,
	NewAliyunDuanxinAlerter,
	NewDingTalkWebhookAlerter,
	NewDingTalkAppAlerter,
	NewEmailAlerter,
	NewFeishuAppAlerter,
	NewFeishuWebhookAlerter,
	NewWebhookAlerter,
	NewQiYeWeiXinAlerter,
	NewQiYeWeiXinAppAlerter,
	NewTencentCloudDHAlerter,
	NewTencentCloudDXAlerter,
)

func NewAlerterList(
	a1 *AliyunDianhuaAlerter,
	a2 *AliyunDuanxinAlerter,
	a3 *DingTalkWebhookAlerter,
	a4 *DingTalkAppAlerter,
	a5 *EmailAlerter,
	a6 *FeishuWebhookAlerter,
	a7 *WebhookAlerter,
	a8 *QiYeWeiXinAlerter,
	a9 *QiYeWeiXinAppAlerter,
	a10 *TencentCloudDHAlerter,
	a11 *TencentCloudDXAlerter,
	a12 *FeishuAppAlerter,
) biz.IAlerterList {
	return biz.IAlerterList{
		a1,
		a2,
		a3,
		a4,
		a5,
		a6,
		a7,
		a8,
		a9,
		a10,
		a11,
		a12,
	}
}
