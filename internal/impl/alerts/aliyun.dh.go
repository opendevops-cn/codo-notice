package alerts

import (
	"context"
	"fmt"

	"codo-notice/internal/biz"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/dyvmsapi"
	"go.uber.org/multierr"
)

type AliyunDianhuaAlerter struct {
	render *TemplateRender
}

func NewAliyunDianhuaAlerter(render *TemplateRender) *AliyunDianhuaAlerter {
	return &AliyunDianhuaAlerter{render: render}
}

func (x *AliyunDianhuaAlerter) Alert(ctx context.Context, ac biz.AlertContext) error {
	if !ac.AlertNotifyData.AliyunDH.IsValid() {
		return nil
	}

	dhCfg := ac.AlertNotifyData.AliyunDH
	content, err := x.render.Render(ctx, ac.Template.Content, ac.Args)
	if err != nil {
		return fmt.Errorf("[AliyunDianhuaAlerter][Alert][Render] err=%w, template=%s, args=%+v", err, ac.Template.Content, ac.Args)
	}

	var errs []error

	for _, user := range ac.CC {
		number := user.Tel
		client, err := dyvmsapi.NewClientWithAccessKey("cn-hangzhou", dhCfg.AccessId, dhCfg.AccessSecret)
		if err != nil {
			errs = append(errs, fmt.Errorf("[AliyunDianhuaAlerter][Alert] phone=%s, access_id=%s, err=%w", number, dhCfg.AccessId, err))
		}
		request := dyvmsapi.CreateSingleCallByTtsRequest()
		request.Scheme = "https"
		request.CalledNumber = number
		request.CalledShowNumber = dhCfg.CalledShowNumber
		request.TtsCode = dhCfg.TtsCode

		// request.TtsParam = param.Content
		// 内部直接对接云模版
		// {"msg":"xxx内容"}
		request.TtsParam = fmt.Sprintf("{\"msg\":\"%s\"}", content)
		request.PlayTimes = requests.NewInteger(2)
		resp, err := client.SingleCallByTts(request)
		if err != nil {
			errs = append(errs, fmt.Errorf("[AliyunDianhuaAlerter][Alert] phone=%s, err=%w", number, err))
			continue
		}
		if resp.Code != "OK" {
			errs = append(errs, fmt.Errorf("[AliyunDianhuaAlerter][Alert] phone=%s, msg=%s", number, resp.Message))
			continue
		}
	}
	if len(errs) > 0 {
		return multierr.Combine(errs...)
	}
	return nil
}

func (x *AliyunDianhuaAlerter) Type() biz.NotifyType {
	return biz.NotifyAliyunDH
}
