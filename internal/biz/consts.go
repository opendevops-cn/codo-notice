package biz

type NotifyType string

// 通知渠道标识 英文小写
const (
	// NotifyEmail 邮件
	NotifyEmail = NotifyType("email")
	// NotifyFeiShu 飞书机器人
	NotifyFeiShu = NotifyType("fs")
	// NotifyFeiShuApp 飞书机器人应用
	NotifyFeiShuApp = NotifyType("fsapp")
	// NotifyAliyunDX 阿里云短信
	NotifyAliyunDX = NotifyType("alydx")
	// NotifyAliyunDH 阿里云电话
	NotifyAliyunDH = NotifyType("alydh")
	// NotifyTengXunDX 腾讯云短信
	NotifyTengXunDX = NotifyType("txdx")
	// NotifyTengXunDH 腾讯云电话
	NotifyTengXunDH = NotifyType("txdh")
	// NotifyDingDing 钉钉机器人
	NotifyDingDing = NotifyType("dd")
	// NotifyDingDingApp 钉钉机器人应用
	NotifyDingDingApp = NotifyType("ddapp")
	// NotifyQiYeWX 企业微信
	NotifyQiYeWX = NotifyType("wx")
	// NotifyQiYeWXApp 企业微信应用
	NotifyQiYeWXApp = NotifyType("wxapp")
	// NotifyWebhook Webhook
	NotifyWebhook = NotifyType("webhook")
)
