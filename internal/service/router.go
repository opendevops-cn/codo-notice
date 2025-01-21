package service

import (
	"context"
	"encoding/json"
	"time"

	"codo-notice/internal/biz"
	"codo-notice/internal/imiddleware"
	routerpb "codo-notice/pb/router"

	"github.com/ccheers/xpkg/generic/arrayx"
	"github.com/opendevops-cn/codo-golang-sdk/cerr"
	loggersdk "github.com/opendevops-cn/codo-golang-sdk/logger"
	"google.golang.org/protobuf/types/known/structpb"
)

type RouterService struct {
	uc      biz.IRouterUseCase
	alertUC biz.IAlertUseCase
	usrUC   biz.IUserUseCase
}

func NewRouterService(uc biz.IRouterUseCase, alertUC biz.IAlertUseCase, usrUC biz.IUserUseCase) *RouterService {
	return &RouterService{
		uc:      uc,
		alertUC: alertUC,
		usrUC:   usrUC,
	}
}

func (x *RouterService) ListRouter(ctx context.Context, request *routerpb.ListRouterRequest) (*routerpb.ListRouterReply, error) {
	list, cnt, err := x.uc.List(ctx, biz.RouterQuery{
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

	return &routerpb.ListRouterReply{
		Data: arrayx.Map(list, func(t *biz.Router) *routerpb.RouterDTO {
			return x.convertRouterDTO(t)
		}),
		Count: int32(cnt),
	}, nil
}

func (x *RouterService) GetRouter(ctx context.Context, request *routerpb.GetRouterRequest) (*routerpb.RouterDTO, error) {
	item, err := x.uc.Get(ctx, request.Id)
	if err != nil {
		return nil, err
	}

	return x.convertRouterDTO(item), nil
}

func (x *RouterService) CreateRouter(ctx context.Context, request *routerpb.CreateRouterRequest) (*routerpb.RouterDTO, error) {
	usr, err := imiddleware.GetUserFromContext(ctx)
	if err != nil {
		return nil, cerr.New(cerr.EUnAuthCode, err)
	}

	data, err := x.uc.Create(ctx, &biz.Router{
		Name:          request.Name,
		Description:   request.Description,
		Status:        request.Status,
		ChannelID:     request.ChannelId,
		ConditionList: x.convertConditionListDO(request.ConditionList),
		CreatedBy:     usr.FullName(),
		UpdatedBy:     usr.FullName(),
	})
	if err != nil {
		return nil, err
	}

	return x.convertRouterDTO(data), nil
}

func (x *RouterService) UpdateRouter(ctx context.Context, request *routerpb.UpdateRouterRequest) (*routerpb.UpdateRouterReply, error) {
	usr, err := imiddleware.GetUserFromContext(ctx)
	if err != nil {
		return nil, cerr.New(cerr.EUnAuthCode, err)
	}

	err = x.uc.Update(ctx, &biz.Router{
		ID:            request.Id,
		Name:          request.Name,
		Description:   request.Description,
		Status:        request.Status,
		ChannelID:     request.ChannelId,
		ConditionList: x.convertConditionListDO(request.ConditionList),
		UpdatedBy:     usr.FullName(),
	}, biz.RouterUpdateOptions{
		Name:          true,
		Description:   true,
		Status:        true,
		ChannelId:     true,
		ConditionList: true,
	})
	if err != nil {
		return nil, err
	}

	return &routerpb.UpdateRouterReply{}, nil
}

func (x *RouterService) UpdateRouterBatch(ctx context.Context, request *routerpb.UpdateRouterBatchRequest) (*routerpb.UpdateRouterBatchReply, error) {
	usr, err := imiddleware.GetUserFromContext(ctx)
	if err != nil {
		return nil, cerr.New(cerr.EUnAuthCode, err)
	}

	for _, id := range request.Ids {
		err := x.uc.Update(ctx, &biz.Router{
			ID:            id,
			Name:          request.Name,
			Description:   request.Description,
			Status:        request.Status,
			ChannelID:     request.ChannelId,
			ConditionList: x.convertConditionListDO(request.ConditionList),
			UpdatedBy:     usr.FullName(),
		}, biz.RouterUpdateOptions{
			Name:          true,
			Description:   true,
			Status:        true,
			ChannelId:     true,
			ConditionList: true,
		})
		if err != nil {
			return nil, err
		}
	}

	return &routerpb.UpdateRouterBatchReply{}, nil
}

func (x *RouterService) DeleteRouter(ctx context.Context, request *routerpb.DeleteRouterRequest) (*routerpb.DeleteRouterReply, error) {
	err := x.uc.Delete(ctx, request.Ids)
	if err != nil {
		return nil, err
	}

	return &routerpb.DeleteRouterReply{}, nil
}

func (x *RouterService) AlertRouterGET(ctx context.Context, request *routerpb.AlertRouterRequest) (*routerpb.AlertRouterReply, error) {
	return x.alertRouter(ctx, request)
}

func (x *RouterService) AlertRouterPOST(ctx context.Context, request *routerpb.AlertRouterRequest) (*routerpb.AlertRouterReply, error) {
	return x.alertRouter(ctx, request)
}

func (x *RouterService) alertRouter(ctx context.Context, _ *routerpb.AlertRouterRequest) (*routerpb.AlertRouterReply, error) {
	httpReq, err := imiddleware.ExtraHTTPRequestFromKratosContext(ctx)
	if err != nil {
		return nil, err
	}
	// 获取触发用户
	usr, err := imiddleware.GetUserFromContext(ctx)
	if err != nil {
		return nil, cerr.New(cerr.EUnAuthCode, err)
	}
	trigUser, err := x.usrUC.Get(ctx, usr.UserId)
	if err != nil {
		return nil, cerr.New(cerr.EDataNotFoundCode, err)
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

	// 获取原始BODY数据
	rawData := make(map[string]interface{})
	_ = json.NewDecoder(httpReq.Body).Decode(&rawData)

	err = x.alertUC.RouteAlert(ctx, &biz.AlertInfo{
		Labels:   labels,
		RawData:  rawData,
		TrigUser: trigUser,
		Entrance: biz.AlertEntranceRouter,
	})
	if err != nil {
		return nil, cerr.New(cerr.EUnknownCode, err)
	}

	return &routerpb.AlertRouterReply{}, nil
}

func (x *RouterService) convertRouterDTO(item *biz.Router) *routerpb.RouterDTO {
	return &routerpb.RouterDTO{
		Id:            item.ID,
		CreatedAt:     item.CreatedAt.Format(time.RFC3339Nano),
		UpdatedAt:     item.UpdatedAt.Format(time.RFC3339Nano),
		CreatedBy:     item.CreatedBy,
		UpdatedBy:     item.UpdatedBy,
		Name:          item.Name,
		Description:   item.Description,
		Status:        item.Status,
		ChannelId:     item.ChannelID,
		ConditionList: x.convertConditionListDTO(item.ConditionList),
	}
}

func (x *RouterService) convertConditionListDTO(group biz.QueryGroup) *structpb.ListValue {
	var dtos [][]interface{}
	// todo 这里兼容的 OR 逻辑, 如果后续需要转换 AND OR 嵌套, 这里需要修改
	for _, g := range group.Groups {
		var queries []interface{}
		for _, q := range g.Queries {
			var dst map[string]interface{}
			bs, _ := json.Marshal(QueryDTO{
				Label:    q.Label,
				Operator: q.Operator,
				Value:    q.Value,
				Index:    int(q.Index),
				Status:   int(q.Status),
			})

			_ = json.Unmarshal(bs, &dst)
			queries = append(queries, dst)
		}
		dtos = append(dtos, queries)
	}

	// 转换成 ListValue
	var data []any
	for _, item := range dtos {
		data = append(data, item)
	}
	result, err := structpb.NewList(data)
	if err != nil {
		loggersdk.Errorf(context.TODO(), "convertConditionListDTO error: %v", err)
	}
	return result
}

func (x *RouterService) convertRouterDO(item *routerpb.RouterDTO) *biz.Router {
	return &biz.Router{
		ID:            item.Id,
		CreatedBy:     item.CreatedBy,
		UpdatedBy:     item.UpdatedBy,
		Name:          item.Name,
		Description:   item.Description,
		Status:        item.Status,
		ChannelID:     item.ChannelId,
		ConditionList: x.convertConditionListDO(item.ConditionList),
	}
}

type QueryDTO struct {
	Label    string `json:"label"`
	Operator string `json:"operator"`
	Value    string `json:"value"`
	Index    int    `json:"index"`
	Status   int    `json:"status"`
}

func (x *RouterService) convertConditionListDO(dto *structpb.ListValue) biz.QueryGroup {
	bs, _ := dto.MarshalJSON()

	var data [][]QueryDTO
	_ = json.Unmarshal(bs, &data)

	// todo 这里兼容的 OR 逻辑, 如果后续需要转换 AND OR 嵌套, 这里需要修改
	group := biz.QueryGroup{
		CombineType: biz.CombineTypeOR,
	}
	for _, item := range data {
		group.Groups = append(group.Groups, &biz.QueryGroup{
			Queries: arrayx.Map(item, func(t QueryDTO) *biz.Query {
				return &biz.Query{
					Label:    t.Label,
					Operator: t.Operator,
					Value:    t.Value,
					Index:    int32(t.Index),
					Status:   int32(t.Status),
				}
			}),
			CombineType: biz.CombineTypeAND,
		})
	}
	return group
}
