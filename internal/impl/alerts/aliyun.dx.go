package alerts

import (
	"context"
	"fmt"
	"strings"

	"codo-notice/internal/biz"

	openapi "github.com/alibabacloud-go/darabonba-openapi/client"
	dysmsapi20170525 "github.com/alibabacloud-go/dysmsapi-20170525/v2/client"
	"github.com/alibabacloud-go/tea-utils/service"
	"github.com/alibabacloud-go/tea/tea"
	"github.com/ccheers/xpkg/generic/arrayx"
)

type AliyunDuanxinAlerter struct {
	render *TemplateRender
}

func NewAliyunDuanxinAlerter(render *TemplateRender) *AliyunDuanxinAlerter {
	return &AliyunDuanxinAlerter{render: render}
}

func (x *AliyunDuanxinAlerter) Alert(ctx context.Context, ac biz.AlertContext) error {
	if !ac.AlertNotifyData.AliyunDX.IsValid() {
		return nil
	}

	dxConf := ac.AlertNotifyData.AliyunDX

	content, err := x.render.Render(ctx, ac.Template.Content, ac.Args)
	if err != nil {
		return err
	}

	phoneStr := strings.Join(arrayx.Map(ac.CC, func(t *biz.User) string {
		return t.Tel
	}), ",")

	client, err := x.createClient(tea.String(dxConf.AccessId), tea.String(dxConf.AccessSecret))
	if err != nil {
		return err
	}
	sendSmsRequest := &dysmsapi20170525.SendSmsRequest{
		// 新版批量发送还是存在 多个里面如果有失败时 返回结果只有 success的问题
		// todo 如果需要的话 改成循环发送
		PhoneNumbers: tea.String(phoneStr),
		SignName:     tea.String(dxConf.SignName),
		TemplateCode: tea.String(dxConf.Template),
		// TemplateParam: tea.String(param.Content),

		// 内部直接对接云模版
		// {"msg":"xxx内容"}
		TemplateParam: tea.String(fmt.Sprintf("{\"msg\":\"%s\"}", content)),
	}
	runtime := &service.RuntimeOptions{}
	// send
	resp, err := client.SendSmsWithOptions(sendSmsRequest, runtime)
	if err != nil {
		return fmt.Errorf("[AliyunDuanxinAlerter][Alert][SendSms] phone=%v, err=%w", phoneStr, err)
	}
	if *resp.Body.Code != "OK" {
		return fmt.Errorf("[AliyunDuanxinAlerter][Alert][SendSms] phone=%v, code=%s, msg=%s", phoneStr, *resp.Body.Code, *resp.Body.Message)
	}
	return nil
}

func (x *AliyunDuanxinAlerter) createClient(accessKeyId *string, accessKeySecret *string) (_result *dysmsapi20170525.Client, _err error) {
	config := &openapi.Config{
		// 您的 AccessKey ID
		AccessKeyId: accessKeyId,
		// 您的 AccessKey Secret
		AccessKeySecret: accessKeySecret,
	}
	// 访问的域名
	config.Endpoint = tea.String("dysmsapi.aliyuncs.com")
	_result = &dysmsapi20170525.Client{}
	_result, _err = dysmsapi20170525.NewClient(config)
	return _result, _err
}

func (x *AliyunDuanxinAlerter) Type() biz.NotifyType {
	return biz.NotifyAliyunDX
}
