package alerts

import (
	"context"
	"fmt"

	"codo-notice/internal/biz"

	"github.com/ccheers/xpkg/generic/arrayx"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	sms "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/sms/v20210111"
)

type TencentCloudDXAlerter struct {
	render *TemplateRender
}

func NewTencentCloudDXAlerter(render *TemplateRender) *TencentCloudDXAlerter {
	return &TencentCloudDXAlerter{render: render}
}

func (x *TencentCloudDXAlerter) Alert(ctx context.Context, ac biz.AlertContext) error {
	if !ac.AlertNotifyData.TengXunDX.IsValid() {
		return nil
	}
	cfg := ac.AlertNotifyData.TengXunDX
	content, err := x.render.Render(ctx, ac.Template.Content, ac.Args)
	if err != nil {
		return fmt.Errorf("[TencentCloudDXAlerter][Alert][Render] err=%w, template=%s, args=%+v", err, ac.Template.Content, ac.Args)
	}

	credential := common.NewCredential(
		cfg.AccessId,
		cfg.AccessSecret,
	)
	cpf := profile.NewClientProfile()
	client, err := sms.NewClient(credential, cfg.RegionID, cpf)
	if err != nil {
		return fmt.Errorf("[TencentCloudDXAlerter][Alert][NewClient] err=%w", err)
	}
	request := sms.NewSendSmsRequest()
	request.SmsSdkAppId = common.StringPtr(cfg.AppId)
	request.SignName = common.StringPtr(cfg.SignName)
	request.TemplateId = common.StringPtr(cfg.Template)
	request.PhoneNumberSet = common.StringPtrs(arrayx.Filter(arrayx.Map(ac.CC, func(t *biz.User) string {
		return t.Tel
	}), func(s string) bool {
		return s != ""
	}))
	request.TemplateParamSet = common.StringPtrs([]string{content})
	// send
	_, err = client.SendSms(request)
	if err != nil {
		return fmt.Errorf("[TencentCloudDXAlerter][Alert][SendSms] phone=%v, err=%w", request.PhoneNumberSet, err)
	}
	return nil
}

func (x *TencentCloudDXAlerter) Type() biz.NotifyType {
	return biz.NotifyTengXunDX
}
