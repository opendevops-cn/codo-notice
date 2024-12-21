package alerts

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"codo-notice/internal/biz"

	"github.com/ccheers/xpkg/generic/arrayx"
	"github.com/opendevops-cn/codo-golang-sdk/client/xhttp"
	loggersdk "github.com/opendevops-cn/codo-golang-sdk/logger"
	"github.com/tidwall/sjson"
	"go.uber.org/multierr"
)

type FeishuWebhookAlerter struct {
	logger *loggersdk.Helper

	render *TemplateRender
	client xhttp.IClient
}

func NewFeishuWebhookAlerter(render *TemplateRender, client xhttp.IClient) *FeishuWebhookAlerter {
	return &FeishuWebhookAlerter{render: render, client: client}
}

func (x *FeishuWebhookAlerter) Alert(ctx context.Context, ac biz.AlertContext) error {
	if !ac.AlertNotifyData.FeiShu.IsValid() {
		return nil
	}
	fsCfg := ac.AlertNotifyData.FeiShu
	jsonDoc := `{"msg_type":"interactive"}`
	// 配置信息
	if _, err := sjson.Set(jsonDoc, "card.config", map[string]bool{
		"update_multi":     true,
		"wide_screen_mode": true,
	}); err != nil {
		return err
	}

	text, err := x.render.Render(ctx, ac.Template.Content, ac.Args)
	if err != nil {
		return err
	}
	// 特征信息
	color := getTitleColor(text, ac)
	jsonDoc, err = sjson.Set(jsonDoc, "card.header.template", color)
	if err != nil {
		return err
	}

	// 标题信息
	titleInfo := map[string]string{
		"tag":     "plain_text",
		"content": ac.Title,
	}
	jsonDoc, err = sjson.Set(jsonDoc, "card.header.title", titleInfo)
	if err != nil {
		return err
	}

	noticer, manager := ac.CC, ac.Manager
	// 正文
	detail := []map[string]interface{}{
		{
			"tag":     "markdown",
			"content": text,
		},
	}
	if len(noticer) > 0 || len(manager) > 0 {
		// 添加通知人信息
		fields := make([]map[string]interface{}, 0)
		if len(noticer) > 0 {
			fields = append(fields, map[string]interface{}{
				"is_short": false,
				"text": map[string]string{
					"tag": "lark_md",
					"content": fmt.Sprintf("**通知人：**%s", strings.Join(arrayx.Map(noticer, func(t *biz.User) string {
						return t.Nickname
					}), ", ")),
				},
			})
		}
		if len(manager) > 0 {
			fields = append(fields, map[string]interface{}{
				"is_short": false,
				"text": map[string]string{
					"tag": "lark_md",
					"content": fmt.Sprintf("**负责人：**%s", strings.Join(arrayx.Map(manager, func(t *biz.User) string {
						return t.Nickname
					}), ", ")),
				},
			})
		}
		if len(fields) > 0 {
			detail = append(detail, map[string]interface{}{"tag": "hr"})
			detail = append(detail, map[string]interface{}{
				"tag":    "div",
				"fields": fields,
			})
		}
	}
	jsonDoc, err = sjson.Set(jsonDoc, "card.elements", detail)
	if err != nil {
		return err
	}

	// send
	var errs []error
	for url, secret := range fsCfg.URLSet {
		err = x.doAlert(ctx, jsonDoc, url, secret)
		if err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return multierr.Combine(errs...)
	}
	return nil
}

func (x *FeishuWebhookAlerter) Type() biz.NotifyType {
	return biz.NotifyFeiShu
}

func (x *FeishuWebhookAlerter) genSign(secret string, timestamp int64) (string, error) {
	// timestamp + key 做sha256, 再进行base64 encode
	stringToSign := fmt.Sprintf("%v", timestamp) + "\n" + secret
	var data []byte
	hm := hmac.New(sha256.New, []byte(stringToSign))
	_, err := hm.Write(data)
	if err != nil {
		return "", err
	}
	signature := base64.StdEncoding.EncodeToString(hm.Sum(nil))
	return signature, nil
}

func (x *FeishuWebhookAlerter) doAlert(ctx context.Context, jsonDoc string, url, secret string) error {
	type Resp struct {
		StatusCode int    `json:"StatusCode"`
		Code       int    `json:"code"`
		Message    string `json:"msg"`
	}

	header := make(http.Header)
	header.Set("Content-Type", "application/json")
	// 签名
	timestamp := time.Now().Unix()
	sign, err := x.genSign(secret, timestamp)
	if err != nil {
		return fmt.Errorf("[FeishuWebhookAlerter][doAlert][genSign] err=%w", err)
	}
	jsonDoc, err = sjson.Set(jsonDoc, "timestamp", fmt.Sprintf("%d", timestamp))
	if err != nil {
		return fmt.Errorf("[FeishuWebhookAlerter][doAlert][sjson.Set] err=%w, path=timestamp", err)
	}

	jsonDoc, err = sjson.Set(jsonDoc, "sign", sign)
	if err != nil {
		return fmt.Errorf("[FeishuWebhookAlerter][doAlert][sjson.Set] err=%w, path=sign", err)
	}

	req, err := http.NewRequest(http.MethodPost, url, strings.NewReader(jsonDoc))
	if err != nil {
		return fmt.Errorf("[FeishuWebhookAlerter][doAlert][NewRequest] err=%w, url=%s, json=%s", err, url, jsonDoc)
	}
	req.Header = header

	resp, err := x.client.Do(ctx, req)
	if err != nil {
		return fmt.Errorf("[FeishuWebhookAlerter][doAlert][Do] err=%w", err)
	}
	defer resp.Body.Close()
	var dst Resp
	err = json.NewDecoder(resp.Body).Decode(&dst)
	if err != nil {
		return fmt.Errorf("[FeishuWebhookAlerter][doAlert][Decode] err=%w", err)
	}
	if dst.StatusCode != 0 {
		return fmt.Errorf("[FeishuWebhookAlerter][doAlert] code=%d, msg=%s", dst.Code, dst.Message)
	}
	return nil
}

// getTitleColor
func getTitleColor(content string, ac biz.AlertContext) string {
	if ac.Status.IsResolved() {
		return "green"
	}
	if ac.Status != biz.AlertStatusFiring {
		text := content
		if strings.Count(text, "resolved") > 0 && strings.Count(text, "firing") > 0 {
			return "orange"
		} else if strings.Count(text, "resolved") > 0 {
			return "green"
		} else {
			return "red"
		}
	}
	switch ac.Severity {
	case biz.AlertSeverityWarn, biz.AlertSeverityInfo:
		return "orange"
	case biz.AlertSeverityError:
		return "red"
	case biz.AlertSeverityFatal:
		return "carmine"
	}
	return "red"
}
