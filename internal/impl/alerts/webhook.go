package alerts

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"codo-notice/internal/biz"

	"github.com/opendevops-cn/codo-golang-sdk/client/xhttp"
	"go.uber.org/multierr"
)

type WebhookAlerter struct {
	render *TemplateRender

	client xhttp.IClient
}

func NewWebhookAlerter(render *TemplateRender, client xhttp.IClient) *WebhookAlerter {
	return &WebhookAlerter{render: render, client: client}
}

func (x *WebhookAlerter) Alert(ctx context.Context, ac biz.AlertContext) error {
	if !ac.AlertNotifyData.Webhook.IsValid() {
		return nil
	}
	content, err := x.render.Render(ctx, ac.Template.Content, ac.Args)
	if err != nil {
		return fmt.Errorf("[WebhookAlerter][Alert][Render] err=%w, template=%s, args=%+v", err, ac.Template.Content, ac.Args)
	}
	header := make(http.Header)
	header.Set("Content-Type", "application/json")
	cfg := ac.AlertNotifyData.Webhook
	var errs []error
	for addr, secret := range cfg.URLSet {
		err = x.doAlert(ctx, content, addr, secret)
		if err != nil {
			errs = append(errs, fmt.Errorf("[WebhookAlerter][Alert][doAlert] err=%w, addr=%s, secret=%s", err, addr, secret))
		}
	}
	if len(errs) > 0 {
		return multierr.Combine(errs...)
	}

	return nil
}

func (x *WebhookAlerter) Type() biz.NotifyType {
	return biz.NotifyWebhook
}

func (x *WebhookAlerter) doAlert(ctx context.Context, content string, url string, secret string) error {
	header := make(http.Header)
	header.Set("Content-Type", "application/json")

	r, err := http.NewRequest(http.MethodPost, url, strings.NewReader(content))
	if err != nil {
		return fmt.Errorf("[WebhookAlerter][doAlert][NewRequest] err=%w, url=%s, content=%s", err, url, content)
	}
	r.Header = header
	query := r.URL.Query()
	query.Set("timestamp", strconv.Itoa(int(time.Now().Unix())))
	r.URL.RawQuery = query.Encode()
	x.setSign(ctx, r, secret)

	resp, err := x.client.Do(ctx, r)
	if err != nil {
		return fmt.Errorf("[WebhookAlerter][doAlert][Do] err=%w, url=%s, content=%s", err, url, content)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("[WebhookAlerter][doAlert][StatusCode] code=%d, url=%s, content=%s", resp.StatusCode, url, content)
	}
	return nil
}

func (x *WebhookAlerter) setSign(ctx context.Context, r *http.Request, secret string) {
	query := r.URL.Query()
	for k := range query {
		if k == "x-sign" {
			query.Del(k)
		}
	}
	md5BS := md5.Sum([]byte(query.Encode() + secret))
	query.Set("x-sign", hex.EncodeToString(md5BS[:]))
	r.URL.RawQuery = query.Encode()
}
