package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"codo-notice/internal/biz"
	"codo-notice/internal/imiddleware"
	templatespb "codo-notice/pb/templates"

	"github.com/ccheers/xpkg/generic/arrayx"
	"github.com/opendevops-cn/codo-golang-sdk/cerr"
)

type TemplateService struct {
	tmplUC  biz.ITemplateUseCase
	usrUC   biz.IUserUseCase
	alertUC biz.IAlertUseCase
}

func NewTemplateService(uc biz.ITemplateUseCase,
	usrUC biz.IUserUseCase,
	alertUC biz.IAlertUseCase,
) *TemplateService {
	return &TemplateService{
		tmplUC:  uc,
		usrUC:   usrUC,
		alertUC: alertUC,
	}
}

func (x *TemplateService) ListTemplate(ctx context.Context, request *templatespb.ListTemplateRequest) (*templatespb.ListTemplateReply, error) {
	list, cnt, err := x.tmplUC.List(ctx, biz.TemplateQuery{
		PageSize:    request.PageSize,
		PageNum:     request.PageNum,
		Order:       request.Order,
		SearchText:  request.SearchText,
		SearchField: request.SearchField,
		Field:       request.Field,
		Cache:       request.Cache,
		FilterMap:   request.FilterMap.AsMap(),
	})
	if err != nil {
		return nil, err
	}

	return &templatespb.ListTemplateReply{
		Data: arrayx.Map(list, func(t *biz.Template) *templatespb.TemplateDTO {
			return x.convertTemplateDTO(t)
		}),
		Count: int32(cnt),
	}, nil
}

func (x *TemplateService) GetTemplate(ctx context.Context, request *templatespb.GetTemplateRequest) (*templatespb.TemplateDTO, error) {
	t, err := x.tmplUC.Get(ctx, request.Id)
	if err != nil {
		return nil, err
	}

	return x.convertTemplateDTO(t), nil
}

func (x *TemplateService) CreateTemplate(ctx context.Context, request *templatespb.CreateTemplateRequest) (*templatespb.TemplateDTO, error) {
	usr, err := imiddleware.GetUserFromContext(ctx)
	if err != nil {
		return nil, cerr.New(cerr.EUnAuthCode, err)
	}

	data, err := x.tmplUC.Create(ctx, &biz.Template{
		Name:      request.Name,
		Content:   request.Content,
		Type:      biz.NotifyType(request.Type),
		Use:       request.Use,
		CreatedBy: usr.FullName(),
		UpdatedBy: usr.FullName(),
	})
	if err != nil {
		return nil, err
	}

	return x.convertTemplateDTO(data), nil
}

func (x *TemplateService) UpdateTemplate(ctx context.Context, request *templatespb.UpdateTemplateRequest) (*templatespb.UpdateTemplateReply, error) {
	usr, err := imiddleware.GetUserFromContext(ctx)
	if err != nil {
		return nil, cerr.New(cerr.EUnAuthCode, err)
	}

	err = x.tmplUC.Update(ctx, &biz.Template{
		ID:        request.Id,
		Name:      request.Name,
		Content:   request.Content,
		Type:      biz.NotifyType(request.Type),
		Use:       request.Use,
		UpdatedBy: usr.FullName(),
	}, biz.TemplateUpdateOptions{
		Name:    true,
		Content: true,
		Type:    true,
		Use:     true,
	})
	if err != nil {
		return nil, err
	}

	return &templatespb.UpdateTemplateReply{}, nil
}

func (x *TemplateService) UpdateTemplateBatch(ctx context.Context, request *templatespb.UpdateTemplateBatchRequest) (*templatespb.UpdateTemplateBatchReply, error) {
	usr, err := imiddleware.GetUserFromContext(ctx)
	if err != nil {
		return nil, cerr.New(cerr.EUnAuthCode, err)
	}

	for _, id := range request.Ids {
		err := x.tmplUC.Update(ctx, &biz.Template{
			ID:        id,
			Name:      request.Name,
			Content:   request.Content,
			Type:      biz.NotifyType(request.Type),
			Use:       request.Use,
			UpdatedBy: usr.FullName(),
		}, biz.TemplateUpdateOptions{
			Name:    true,
			Content: true,
			Type:    true,
			Use:     true,
		})
		if err != nil {
			return nil, err
		}
	}

	return &templatespb.UpdateTemplateBatchReply{}, nil
}

func (x *TemplateService) DeleteTemplate(ctx context.Context, request *templatespb.DeleteTemplateRequest) (*templatespb.DeleteTemplateReply, error) {
	err := x.tmplUC.Delete(ctx, request.Ids)
	if err != nil {
		return nil, err
	}

	return &templatespb.DeleteTemplateReply{}, nil
}

func (x *TemplateService) AlertTemplate(ctx context.Context, request *templatespb.AlertTemplateRequest) (*templatespb.AlertTemplateReply, error) {
	httpReq, err := imiddleware.ExtraHTTPRequestFromKratosContext(ctx)
	if err != nil {
		return nil, err
	}

	// 获取触发用户
	usr, err := imiddleware.GetUserFromContext(ctx)
	if err != nil {
		return nil, cerr.New(cerr.EUnAuthCode, err)
	}
	usrID, _ := strconv.Atoi(usr.UserId)
	trigUser, err := x.usrUC.Get(ctx, uint32(usrID))
	if err != nil {
		return nil, cerr.New(cerr.EUnAuthCode, err)
	}

	// 获取标签和原始数据
	query := httpReq.URL.Query()
	labels := make(map[string]string, len(query))
	for k, v := range query {
		var str string
		if len(v) > 0 {
			str = v[0]
		}
		labels[k] = str
	}
	var rawData map[string]interface{}
	err = json.NewDecoder(httpReq.Body).Decode(&rawData)
	if err != nil {
		return nil, cerr.New(cerr.EParamUnparsedCode, err)
	}

	// 获取模板
	tplName := request.Tpl
	filter := map[string]interface{}{
		"name": tplName,
	}
	if tplName == "" {
		filter = map[string]interface{}{
			"type":    request.Type,
			"use":     request.Use,
			"default": "yes",
		}
	}
	list, _, err := x.tmplUC.List(ctx, biz.TemplateQuery{
		ListAll:   false,
		PageSize:  1,
		PageNum:   1,
		FilterMap: filter,
	})
	if err != nil {
		return nil, err
	}
	if len(list) == 0 {
		return nil, cerr.New(cerr.EDataNotFoundCode, fmt.Errorf("no template found, filter=%+v", filter))
	}
	tmpl := list[0]

	// 获取抄送用户
	var ccUsers []*biz.User
	if request.Phone != "" {
		usrs, _, err := x.usrUC.List(ctx, biz.UserQuery{
			ListAll:  false,
			PageSize: 999,
			PageNum:  1,
			FilterMap: map[string]interface{}{
				"tel": strings.Split(request.Phone, ","),
			},
		})
		if err != nil {
			return nil, cerr.New(cerr.EDBErrorCode, err)
		}
		ccUsers = append(ccUsers, usrs...)
	}
	if request.Email != "" {
		usrs, _, err := x.usrUC.List(ctx, biz.UserQuery{
			ListAll:  false,
			PageSize: 999,
			PageNum:  1,
			FilterMap: map[string]interface{}{
				"email": strings.Split(request.Email, ","),
			},
		})
		if err != nil {
			return nil, cerr.New(cerr.EDBErrorCode, err)
		}
		ccUsers = append(ccUsers, usrs...)
	}
	if request.At != "" {
		usrs, _, err := x.usrUC.List(ctx, biz.UserQuery{
			ListAll:  false,
			PageSize: 999,
			PageNum:  1,
			FilterMap: map[string]interface{}{
				"username": strings.Split(request.At, ","),
			},
		})
		if err != nil {
			return nil, cerr.New(cerr.EDBErrorCode, err)
		}
		ccUsers = append(ccUsers, usrs...)
	}
	if request.Phone != "" {
		usrs, _, err := x.usrUC.List(ctx, biz.UserQuery{
			ListAll:  false,
			PageSize: 999,
			PageNum:  1,
			FilterMap: map[string]interface{}{
				"tel": strings.Split(request.Phone, ","),
			},
		})
		if err != nil {
			return nil, cerr.New(cerr.EDBErrorCode, err)
		}
		ccUsers = append(ccUsers, usrs...)
	}

	err = x.alertUC.Alert(ctx, &biz.AlertInfo{
		Template: tmpl,
		Labels:   labels,
		RawData:  rawData,
		TrigUser: trigUser,
		Entrance: biz.AlertEntranceTmpl,
		CCUsers:  ccUsers,
		ExtraDataNotifyURLs: map[string]string{
			request.Url: request.Secret,
		},
	})
	if err != nil {
		return nil, err
	}

	return &templatespb.AlertTemplateReply{}, nil
}

func (x *TemplateService) convertTemplateDTO(template *biz.Template) *templatespb.TemplateDTO {
	return &templatespb.TemplateDTO{
		Id:        template.ID,
		CreatedAt: template.CreatedAt.Format(time.RFC3339Nano),
		UpdatedAt: template.UpdatedAt.Format(time.RFC3339Nano),
		CreatedBy: template.CreatedBy,
		UpdatedBy: template.UpdatedBy,
		Name:      template.Name,
		Content:   template.Content,
		Type:      string(template.Type),
		Use:       template.Use,
		Default:   template.Default,
		Path:      template.Path,
	}
}

func (x *TemplateService) convertTemplateDO(template *templatespb.TemplateDTO) *biz.Template {
	return &biz.Template{
		ID:      template.Id,
		Name:    template.Name,
		Content: template.Content,
		Type:    biz.NotifyType(template.Type),
		Use:     template.Use,
		Default: template.Default,
		Path:    template.Path,
	}
}
