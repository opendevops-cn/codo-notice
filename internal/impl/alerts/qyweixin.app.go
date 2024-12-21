package alerts

import (
	"context"
	"fmt"
	"strings"

	"codo-notice/internal/biz"

	"github.com/ccheers/xpkg/generic/arrayx"
	"github.com/opendevops-cn/codo-golang-sdk/client/xhttp"

	"github.com/ysicing/workwxbot"
)

type QiYeWeiXinAppAlerter struct {
	client xhttp.IClient
	render *TemplateRender
}

func NewQiYeWeiXinAppAlerter(client xhttp.IClient, render *TemplateRender) *QiYeWeiXinAppAlerter {
	return &QiYeWeiXinAppAlerter{client: client, render: render}
}

func (x *QiYeWeiXinAppAlerter) Alert(ctx context.Context, ac biz.AlertContext) error {
	if !ac.AlertNotifyData.QiYeWXApp.IsValid() {
		return nil
	}
	content, err := x.render.Render(ctx, ac.Template.Content, ac.Args)
	if err != nil {
		return fmt.Errorf("[QiYeWeiXinAppAlerter][Alert][Render] err=%w, template=%s, args=%+v", err, ac.Template.Content, ac.Args)
	}

	cfg := ac.AlertNotifyData.QiYeWXApp
	client := workwxbot.Client{
		CropID:      cfg.CropId,
		AgentID:     cfg.AgentId,
		AgentSecret: cfg.AgentSecret,
	}
	request := workwxbot.Message{
		ToUser: strings.Join(arrayx.Filter(arrayx.Map(ac.CC, func(t *biz.User) string {
			return t.Username
		}), func(m string) bool {
			return m != ""
		}), ","),
		// ToParty:  toparty,
		// ToTag:    totag,
		MsgType:  "markdown",
		Markdown: workwxbot.Content{Content: content},
	}

	if err := client.Send(request); err != nil {
		return fmt.Errorf("[QiYeWeiXinAppAlerter][Alert][Send] err=%w, at=%+v, content=%s", err, ac.CC, content)
	}
	return nil
}

func (x *QiYeWeiXinAppAlerter) Type() biz.NotifyType {
	return biz.NotifyQiYeWXApp
}
