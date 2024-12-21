package alerts

import (
	"context"
	"crypto/tls"
	"fmt"

	"codo-notice/internal/biz"

	"github.com/ccheers/xpkg/generic/arrayx"
	"github.com/go-gomail/gomail"
)

type EmailAlerter struct {
	render *TemplateRender
}

func NewEmailAlerter(render *TemplateRender) *EmailAlerter {
	return &EmailAlerter{render: render}
}

func (x *EmailAlerter) Alert(ctx context.Context, ac biz.AlertContext) error {
	if !ac.AlertNotifyData.Email.IsValid() {
		return nil
	}

	content, err := x.render.Render(ctx, ac.Template.Content, ac.Args)
	if err != nil {
		return fmt.Errorf("[EmailAlerter][Alert][Render] err=%w, template=%s, args=%+v", err, ac.Template.Content, ac.Args)
	}
	cfg := ac.AlertNotifyData.Email
	title := ac.Title
	from := cfg.User
	subject := title
	serverHost := cfg.Host
	serverPort := cfg.Port

	m := gomail.NewMessage()
	emails := arrayx.Filter(arrayx.Map(ac.CC, func(t *biz.User) string {
		return t.Email
	}), func(s string) bool {
		return s != ""
	})

	m.SetHeader("To", emails...)
	m.SetAddressHeader("From", from, "ops-notice")
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", content)
	d := gomail.NewDialer(serverHost, serverPort, from, cfg.Password)
	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	err = d.DialAndSend(m)
	if err != nil {
		return fmt.Errorf("[EmailAlerter][Alert][SendEmail] err=%w", err)
	}
	return nil
}

func (x *EmailAlerter) Type() biz.NotifyType {
	return biz.NotifyEmail
}
