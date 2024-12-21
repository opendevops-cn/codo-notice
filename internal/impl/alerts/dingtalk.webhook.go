package alerts

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"codo-notice/internal/biz"

	"github.com/opendevops-cn/codo-golang-sdk/client/xhttp"
	"go.uber.org/multierr"
)

type DingTalkWebhookAlerter struct {
	render *TemplateRender

	client xhttp.IClient
}

func NewDingTalkWebhookAlerter(render *TemplateRender, client xhttp.IClient) *DingTalkWebhookAlerter {
	return &DingTalkWebhookAlerter{render: render, client: client}
}

func (x *DingTalkWebhookAlerter) Alert(ctx context.Context, ac biz.AlertContext) error {
	if !ac.AlertNotifyData.DingDing.IsValid() {
		return nil
	}

	content, err := x.render.Render(ctx, ac.Template.Content, ac.Args)
	if err != nil {
		return err
	}
	// content
	data := map[string]interface{}{
		"msgtype": "markdown",
		"markdown": map[string]string{
			"title": ac.Title,
			"text":  content,
		},
	}

	var errs []error
	header := make(http.Header)
	header.Set("Content-Type", "application/json")
	for addr, secret := range ac.AlertNotifyData.DingDing.URLSet {
		err := x.doAlert(ctx, addr, secret, data, header)
		if err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return multierr.Combine(errs...)
	}
	return nil
}

func (x *DingTalkWebhookAlerter) Type() biz.NotifyType {
	return biz.NotifyDingDing
}

func (x *DingTalkWebhookAlerter) doAlert(ctx context.Context, addr string, secret string, data map[string]interface{}, header http.Header) error {
	type Resp struct {
		ErrCode    int    `json:"errcode"`
		ErrMessage string `json:"errmsg"`
	}

	u := addr
	if secret != "" {
		// 签名
		uri, err := url.Parse(addr)
		if err != nil {
			return fmt.Errorf("[DingTalkWebhookAlerter][Alert][url.Parse]: url=%s, err=%s", addr, err.Error())
		}

		timestamp := time.Now().UnixNano() / 1e6
		sign, err := x.genSign(secret, timestamp)
		if err != nil {
			return fmt.Errorf("[DingTalkWebhookAlerter][Alert][genSign]: url=%s, err=%s", addr, err.Error())
		}
		query := uri.Query()
		query.Set("timestamp", fmt.Sprintf("%d", timestamp))
		query.Set("sign", sign)
		uri.RawQuery = query.Encode()
		u = uri.String()
	}

	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(data)
	if err != nil {
		return fmt.Errorf("[DingTalkWebhookAlerter][Alert][json.Encode]: url=%s, err=%s", addr, err.Error())
	}
	req, err := http.NewRequest(http.MethodPost, u, &buf)
	if err != nil {
		return fmt.Errorf("[DingTalkWebhookAlerter][Alert][http.NewRequest]: url=%s, err=%s", addr, err.Error())
	}
	req.Header = header
	resp, err := x.client.Do(ctx, req)
	if err != nil {
		return fmt.Errorf("[DingTalkWebhookAlerter][Alert][client.Do]: url=%s, err=%s", addr, err.Error())
	}
	defer resp.Body.Close()
	var dst Resp
	err = json.NewDecoder(resp.Body).Decode(&dst)
	if err != nil {
		return fmt.Errorf("[DingTalkWebhookAlerter][Alert][json.Decode]: url=%s, err=%s", addr, err.Error())
	}
	if dst.ErrCode != 0 {
		return fmt.Errorf("[DingTalkWebhookAlerter][Alert]: url=%s, errcode=%d, errmsg=%s", addr, dst.ErrCode, dst.ErrMessage)
	}

	return nil
}

func (x *DingTalkWebhookAlerter) genSign(secret string, timestamp int64) (string, error) {
	stringToSign := fmt.Sprintf("%d\n%s", timestamp, secret)
	hm := hmac.New(sha256.New, []byte(secret))
	_, err := hm.Write([]byte(stringToSign))
	if err != nil {
		return "", err
	}
	signature := url.QueryEscape(base64.StdEncoding.EncodeToString(hm.Sum(nil)))
	return signature, nil
}
