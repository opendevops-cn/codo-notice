package alerts

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"codo-notice/internal/biz"

	"github.com/opendevops-cn/codo-golang-sdk/client/xhttp"
	"go.uber.org/multierr"
)

type QiYeWeiXinAlerter struct {
	client xhttp.IClient
	render *TemplateRender
}

func NewQiYeWeiXinAlerter(client xhttp.IClient, render *TemplateRender) *QiYeWeiXinAlerter {
	return &QiYeWeiXinAlerter{client: client, render: render}
}

func (x *QiYeWeiXinAlerter) Alert(ctx context.Context, ac biz.AlertContext) error {
	if !ac.AlertNotifyData.QiYeWX.IsValid() {
		return nil
	}
	content, err := x.render.Render(ctx, ac.Template.Content, ac.Args)
	if err != nil {
		return fmt.Errorf("[QiYeWeiXinAlerter][Alert][Render] err=%w, template=%s, args=%+v", err, ac.Template.Content, ac.Args)
	}

	// send
	var errs []error
	cfg := ac.AlertNotifyData.QiYeWX
	for addr, secret := range cfg.URLSet {
		err := x.doAlert(ctx, content, addr, secret)
		if err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return multierr.Combine(errs...)
	}

	return nil
}

func (x *QiYeWeiXinAlerter) Type() biz.NotifyType {
	return biz.NotifyQiYeWX
}

func (x *QiYeWeiXinAlerter) doAlert(ctx context.Context, content string, url string, secret string) error {
	header := make(http.Header)
	header.Set("Content-Type", "application/json")

	// content
	data := map[string]interface{}{
		"msgtype": "markdown",
		"markdown": map[string]string{
			"content": content,
		},
	}

	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(data)
	if err != nil {
		return fmt.Errorf("[QiYeWeiXinAlerter][doAlert][Encode] err=%w, data=%+v", err, data)
	}

	r, err := http.NewRequest(http.MethodPost, url, &buf)
	if err != nil {
		return fmt.Errorf("[QiYeWeiXinAlerter][doAlert][NewRequest] err=%w, url=%s, content=%s", err, url, content)
	}
	r.Header = header

	resp, err := x.client.Do(ctx, r)
	if err != nil {
		return fmt.Errorf("[QiYeWeiXinAlerter][doAlert][Do] err=%w, url=%s, content=%s", err, url, content)
	}
	defer resp.Body.Close()

	type Resp struct {
		ErrCode    int    `json:"errcode"`
		ErrMessage string `json:"errmsg"`
	}
	var dst Resp
	if err := json.NewDecoder(resp.Body).Decode(&dst); err != nil {
		return fmt.Errorf("[QiYeWeiXinAlerter][doAlert][Decode] err=%w, url=%s, content=%s", err, url, content)
	}
	if dst.ErrCode != 0 {
		return fmt.Errorf("[QiYeWeiXinAlerter][doAlert] errcode=%d, errmsg=%s", dst.ErrCode, dst.ErrMessage)
	}

	return nil
}
