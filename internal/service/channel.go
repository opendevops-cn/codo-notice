package service

import (
	"context"
	"time"

	"codo-notice/internal/biz"
	"codo-notice/internal/imiddleware"
	channelpb "codo-notice/pb/channel"

	"github.com/ccheers/xpkg/generic/arrayx"
	"github.com/opendevops-cn/codo-golang-sdk/cerr"
)

type ChannelService struct {
	channelUC biz.IChannelUseCase
}

func NewChannelService(channelUC biz.IChannelUseCase) *ChannelService {
	return &ChannelService{channelUC: channelUC}
}

func (x *ChannelService) ListChannel(ctx context.Context, request *channelpb.ListChannelRequest) (*channelpb.ListChannelReply, error) {
	data, cnt, err := x.channelUC.List(ctx, biz.ChannelQuery{
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

	return &channelpb.ListChannelReply{
		Data: arrayx.Map(data, func(t *biz.Channel) *channelpb.ChannelDTO {
			return x.convertChannelDTO(t)
		}),
		Count: int32(cnt),
	}, nil
}

func (x *ChannelService) GetChannel(ctx context.Context, request *channelpb.GetChannelRequest) (*channelpb.ChannelDTO, error) {
	data, err := x.channelUC.Get(ctx, request.Id)
	if err != nil {
		return nil, err
	}
	return x.convertChannelDTO(data), nil
}

func (x *ChannelService) CreateChannel(ctx context.Context, request *channelpb.CreateChannelRequest) (*channelpb.ChannelDTO, error) {
	usr, err := imiddleware.GetUserFromContext(ctx)
	if err != nil {
		return nil, cerr.New(cerr.EUnAuthCode, err)
	}

	data, err := x.channelUC.Create(ctx, &biz.Channel{
		Name: request.Name,
		Use:  request.Use,
		User: request.User,
		ContactPoints: arrayx.Map(request.ContactPoints, func(t *channelpb.ContactPointDTO) *biz.ContactPoint {
			return x.convertContactPointDO(t)
		}),
		CustomItems: arrayx.Map(request.CustomItems, func(t *channelpb.ContactPointDTO) *biz.ContactPoint {
			return x.convertContactPointDO(t)
		}),
		CreatedBy: usr.FullName(),
		UpdatedBy: usr.FullName(),
	})
	if err != nil {
		return nil, err
	}

	return x.convertChannelDTO(data), nil
}

func (x *ChannelService) UpdateChannel(ctx context.Context, request *channelpb.UpdateChannelRequest) (*channelpb.UpdateChannelReply, error) {
	usr, err := imiddleware.GetUserFromContext(ctx)
	if err != nil {
		return nil, cerr.New(cerr.EUnAuthCode, err)
	}

	err = x.channelUC.Update(ctx, &biz.Channel{
		ID:   request.Id,
		Name: request.Name,
		Use:  request.Use,
		User: request.User,
		ContactPoints: arrayx.Map(request.ContactPoints, func(t *channelpb.ContactPointDTO) *biz.ContactPoint {
			return x.convertContactPointDO(t)
		}),
		CustomItems: arrayx.Map(request.CustomItems, func(t *channelpb.ContactPointDTO) *biz.ContactPoint {
			return x.convertContactPointDO(t)
		}),
		UpdatedBy: usr.FullName(),
	}, biz.ChannelUpdateOptions{
		Name:          true,
		Use:           true,
		User:          true,
		ContactPoints: true,
		CustomItems:   true,
	})
	if err != nil {
		return nil, err
	}

	return &channelpb.UpdateChannelReply{}, nil
}

func (x *ChannelService) UpdateChannelBatch(ctx context.Context, request *channelpb.UpdateChannelBatchRequest) (*channelpb.UpdateChannelBatchReply, error) {
	usr, err := imiddleware.GetUserFromContext(ctx)
	if err != nil {
		return nil, cerr.New(cerr.EUnAuthCode, err)
	}

	for _, id := range request.Ids {
		err := x.channelUC.Update(ctx, &biz.Channel{
			ID:   id,
			Name: request.Name,
			Use:  request.Use,
			User: request.User,
			ContactPoints: arrayx.Map(request.ContactPoints, func(t *channelpb.ContactPointDTO) *biz.ContactPoint {
				return x.convertContactPointDO(t)
			}),
			CustomItems: arrayx.Map(request.CustomItems, func(t *channelpb.ContactPointDTO) *biz.ContactPoint {
				return x.convertContactPointDO(t)
			}),
			UpdatedBy: usr.FullName(),
		}, biz.ChannelUpdateOptions{
			Name:          true,
			Use:           true,
			User:          true,
			ContactPoints: true,
			CustomItems:   true,
		})
		if err != nil {
			return nil, err
		}
	}

	return &channelpb.UpdateChannelBatchReply{}, nil
}

func (x *ChannelService) DeleteChannel(ctx context.Context, request *channelpb.DeleteChannelRequest) (*channelpb.DeleteChannelReply, error) {
	err := x.channelUC.Delete(ctx, request.Ids)
	if err != nil {
		return nil, err
	}
	return &channelpb.DeleteChannelReply{}, nil
}

func (x *ChannelService) convertChannelDTO(item *biz.Channel) *channelpb.ChannelDTO {
	return &channelpb.ChannelDTO{
		Id:        item.ID,
		CreatedAt: item.CreatedAt.Format(time.RFC3339Nano),
		UpdatedAt: item.UpdatedAt.Format(time.RFC3339Nano),
		CreatedBy: item.CreatedBy,
		UpdatedBy: item.UpdatedBy,
		Name:      item.Name,
		Use:       item.Use,
		User:      item.User,
		Group:     item.Group,
		ContactPoints: arrayx.Map(item.ContactPoints, func(t *biz.ContactPoint) *channelpb.ContactPointDTO {
			return x.convertContactPointDTO(t)
		}),
		CustomItems: arrayx.Map(item.CustomItems, func(t *biz.ContactPoint) *channelpb.ContactPointDTO {
			return x.convertContactPointDTO(t)
		}),
		DefaultRule: item.DefaultRule,
	}
}

func (x *ChannelService) convertContactPointDTO(item *biz.ContactPoint) *channelpb.ContactPointDTO {
	return &channelpb.ContactPointDTO{
		Id:           item.Id,
		CreatedAt:    item.CreatedAt.Format(time.RFC3339Nano),
		UpdatedAt:    item.UpdatedAt.Format(time.RFC3339Nano),
		Type:         item.Type,
		ChannelId:    item.ChannelId,
		TplId:        item.TemplateId,
		Webhook:      item.Addr,
		Secret:       item.Secret,
		Severity:     item.Severity,
		SeverityDesc: item.SeverityDesc,
		Show:         item.Show,
		Fixed:        item.Fixed,
		AlertWebhooks: arrayx.Map(item.AlertWebhooks, func(t biz.AlertWebhook) *channelpb.AlertWebhookDTO {
			return &channelpb.AlertWebhookDTO{
				Url:       t.URL,
				Alias:     t.Alias,
				IsApprove: t.IsApprove,
				IsReject:  t.IsReject,
			}
		}),
	}
}

func (x *ChannelService) convertChannelDO(item *channelpb.ChannelDTO) *biz.Channel {
	return &biz.Channel{
		ID:    item.Id,
		Name:  item.Name,
		Use:   item.Use,
		User:  item.User,
		Group: item.Group,
		ContactPoints: arrayx.Map(item.ContactPoints, func(t *channelpb.ContactPointDTO) *biz.ContactPoint {
			return x.convertContactPointDO(t)
		}),
		CustomItems: arrayx.Map(item.CustomItems, func(t *channelpb.ContactPointDTO) *biz.ContactPoint {
			return x.convertContactPointDO(t)
		}),
		DefaultRule: item.DefaultRule,
	}
}

func (x *ChannelService) convertContactPointDO(item *channelpb.ContactPointDTO) *biz.ContactPoint {
	return &biz.ContactPoint{
		Id:           item.Id,
		Type:         item.Type,
		ChannelId:    item.ChannelId,
		TemplateId:   item.TplId,
		Addr:         item.Webhook,
		Secret:       item.Secret,
		Severity:     item.Severity,
		SeverityDesc: item.SeverityDesc,
		Show:         item.Show,
		Fixed:        item.Fixed,
		AlertWebhooks: arrayx.Map(item.AlertWebhooks, func(t *channelpb.AlertWebhookDTO) biz.AlertWebhook {
			return biz.AlertWebhook{
				URL:       t.Url,
				Alias:     t.Alias,
				IsApprove: t.IsApprove,
				IsReject:  t.IsReject,
			}
		}),
	}
}
