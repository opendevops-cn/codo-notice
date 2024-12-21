package alerts

import (
	"context"
	"fmt"

	"codo-notice/internal/biz"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	vms "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/vms/v20200902"
	"go.uber.org/multierr"
)

type TencentCloudDHAlerter struct {
	render *TemplateRender
}

func NewTencentCloudDHAlerter(render *TemplateRender) *TencentCloudDHAlerter {
	return &TencentCloudDHAlerter{render: render}
}

func (x *TencentCloudDHAlerter) Alert(ctx context.Context, ac biz.AlertContext) error {
	if !ac.AlertNotifyData.TengXunDH.IsValid() {
		return nil
	}
	cfg := ac.AlertNotifyData.TengXunDH
	credential := common.NewCredential(
		cfg.AccessId,
		cfg.AccessSecret,
	)
	cpf := profile.NewClientProfile()

	content, err := x.render.Render(ctx, ac.Template.Content, ac.Args)
	if err != nil {
		return fmt.Errorf("[TencentCloudDHAlerter][Alert][Render] err=%w, template=%s, args=%+v", err, ac.Template.Content, ac.Args)
	}

	var errs []error
	for _, usr := range ac.CC {
		number := usr.Tel
		client, _ := vms.NewClient(credential, "ap-guangzhou", cpf)
		request := vms.NewSendTtsVoiceRequest()
		request.TemplateId = common.StringPtr(cfg.Template)
		request.TemplateParamSet = common.StringPtrs([]string{content})
		request.CalledNumber = common.StringPtr(number)
		request.VoiceSdkAppid = common.StringPtr(cfg.AppId)
		request.PlayTimes = common.Uint64Ptr(2)
		// send
		_, err := client.SendTtsVoice(request)
		if err != nil {
			errs = append(errs, fmt.Errorf("[TencentCloudDHAlerter][Alert][SendTtsVoice] phone=%s, err=%s", number, err.Error()))
			continue
		}
	}

	if len(errs) > 0 {
		return multierr.Combine(errs...)
	}

	return nil
}

func (x *TencentCloudDHAlerter) Type() biz.NotifyType {
	return biz.NotifyTengXunDH
}
