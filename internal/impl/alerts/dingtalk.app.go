package alerts

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"codo-notice/internal/biz"

	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	dingtalkoauth2_1_0 "github.com/alibabacloud-go/dingtalk/oauth2_1_0"
	"github.com/alibabacloud-go/tea/tea"
	"github.com/ccheers/xpkg/generic/arrayx"
	"github.com/opendevops-cn/codo-golang-sdk/client/xhttp"
)

type ddAppWidget struct {
	endpoint string
}

type DingTalkAppAlerter struct {
	render *TemplateRender
	widget ddAppWidget
	client xhttp.IClient

	mu sync.Mutex
	// appID=>accessToken
	accessTokenMap map[string]string
}

func NewDingTalkAppAlerter(render *TemplateRender, client xhttp.IClient) *DingTalkAppAlerter {
	return &DingTalkAppAlerter{
		render: render,
		widget: ddAppWidget{
			endpoint: "https://oapi.dingtalk.com",
		},
		client:         client,
		mu:             sync.Mutex{},
		accessTokenMap: make(map[string]string),
	}
}

func (x *DingTalkAppAlerter) Alert(ctx context.Context, ac biz.AlertContext) error {
	err := x.doAlert(ctx, ac)
	if err != nil {
		if strings.Contains(err.Error(), "不合法的access_token") {
			_, err = x.getAccessToken(true, ac.AlertNotifyData.DingDingApp)
			if err != nil {
				return err
			}
			return x.doAlert(ctx, ac)
		}
		return err
	}

	return nil
}

func (x *DingTalkAppAlerter) doAlert(ctx context.Context, ac biz.AlertContext) error {
	type Resp struct {
		ErrCode    int    `json:"errcode"`
		ErrMessage string `json:"errmsg"`
	}
	cfg := ac.AlertNotifyData.DingDingApp

	if !cfg.IsValid() {
		return nil
	}

	content, err := x.render.Render(ctx, ac.Template.Content, ac.Args)
	if err != nil {
		return fmt.Errorf("[DingTalkAppAlerter][Alert][Render] err=%w, template=%s, args=%+v", err, ac.Template.Content, ac.Args)
	}
	x.mu.Lock()
	token := x.accessTokenMap[cfg.AppId]
	x.mu.Unlock()

	if token == "" {
		if t, err := x.getAccessToken(false, ac.AlertNotifyData.DingDingApp); err != nil {
			return fmt.Errorf("[DingTalkAppAlerter][Alert][getAccessToken] err=%w", err)
		} else {
			token = t
		}
	}

	// content
	data := map[string]interface{}{
		"to_all_user": "false",
		"agent_id":    cfg.AgentId,
		"userid_list": strings.Join(arrayx.Map(ac.CC, func(t *biz.User) string {
			return t.DdId
		}), ","),
		"msg": map[string]interface{}{
			"msgtype": "markdown",
			"markdown": map[string]string{
				"title": ac.Title,
				"text":  content,
			},
		},
	}
	// send
	addr := fmt.Sprintf("%s/topapi/message/corpconversation/asyncsend_v2?access_token=%s", x.widget.endpoint, token)

	var buf bytes.Buffer
	err = json.NewEncoder(&buf).Encode(data)
	if err != nil {
		return fmt.Errorf("[DingTalkAppAlerter][Alert][json.Encode] err=%w", err)
	}
	req, err := http.NewRequest(http.MethodPost, addr, &buf)
	if err != nil {
		return fmt.Errorf("[DingTalkAppAlerter][Alert][http.NewRequest] err=%w, addr=%s, json=%s", err, addr, buf.String())
	}
	resp, err := x.client.Do(ctx, req)
	if err != nil {
		return fmt.Errorf("[DingTalkAppAlerter][Alert][http.Do] err=%w", err)
	}
	defer resp.Body.Close()
	var dst Resp
	err = json.NewDecoder(resp.Body).Decode(&dst)
	if err != nil {
		return fmt.Errorf("[DingTalkAppAlerter][Alert][json.Decode] err=%w", err)
	}
	if dst.ErrCode != 0 {
		return fmt.Errorf("[error]: at, code=%d, msg=%s", dst.ErrCode, dst.ErrMessage)
	}

	return nil
}

func (x *DingTalkAppAlerter) Type() biz.NotifyType {
	return biz.NotifyDingDingApp
}

func (x *DingTalkAppAlerter) getAccessToken(reset bool, cfg biz.AlertNotifyDataDingDingApp) (string, error) {
	x.mu.Lock()
	defer x.mu.Unlock()

	if reset {
		x.accessTokenMap[cfg.AppId] = ""
	}
	if x.accessTokenMap[cfg.AppId] != "" {
		return x.accessTokenMap[cfg.AppId], nil
	}
	client, err := x.CreateAuthClient()
	if err != nil {
		return "", err
	}

	getAccessTokenRequest := &dingtalkoauth2_1_0.GetAccessTokenRequest{
		AppKey:    tea.String(cfg.AppId),
		AppSecret: tea.String(cfg.AppSecret),
	}
	resp, err := client.GetAccessToken(getAccessTokenRequest)
	if err != nil {
		return "", err
	}
	if resp.Body != nil && resp.Body.AccessToken != nil {
		x.accessTokenMap[cfg.AppId] = *resp.Body.AccessToken
		return x.accessTokenMap[cfg.AppId], nil
	} else {
		return "", fmt.Errorf("get access token failed")
	}
}

func (x *DingTalkAppAlerter) CreateAuthClient() (_result *dingtalkoauth2_1_0.Client, _err error) {
	config := &openapi.Config{}
	config.Protocol = tea.String("https")
	config.RegionId = tea.String("central")
	_result = &dingtalkoauth2_1_0.Client{}
	_result, _err = dingtalkoauth2_1_0.NewClient(config)
	return _result, _err
}
