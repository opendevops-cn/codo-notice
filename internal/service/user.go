package service

import (
	"context"
	"encoding/json"
	"time"

	"codo-notice/internal/biz"
	userpb "codo-notice/pb/user"

	"github.com/ccheers/xpkg/generic/arrayx"
	"github.com/opendevops-cn/codo-golang-sdk/cerr"
	"google.golang.org/protobuf/types/known/structpb"
)

type UserService struct {
	uc biz.IUserUseCase
}

func NewUserService(uc biz.IUserUseCase) *UserService {
	return &UserService{uc: uc}
}

func (x *UserService) ListUser(ctx context.Context, request *userpb.ListUserRequest) (*userpb.ListUserReply, error) {
	list, cnt, err := x.uc.List(ctx, biz.UserQuery{
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

	return &userpb.ListUserReply{
		Data: arrayx.Map(list, func(t *biz.User) *userpb.UserDTO {
			return x.convertDTO(t)
		}),
		Count: int32(cnt),
	}, nil
}

func (x *UserService) GetUser(ctx context.Context, request *userpb.GetUserRequest) (*userpb.UserDTO, error) {
	t, err := x.uc.Get(ctx, request.Id)
	if err != nil {
		return nil, cerr.New(cerr.EDataNotFoundCode, err)
	}

	return x.convertDTO(t), nil
}

func (x *UserService) convertDTO(item *biz.User) *userpb.UserDTO {
	var dst map[string]any
	_ = json.Unmarshal(item.DataSource, &dst)
	ss, _ := structpb.NewStruct(dst)
	return &userpb.UserDTO{
		Id:         item.ID,
		CreatedAt:  item.CreatedAt.Format(time.RFC3339Nano),
		UpdatedAt:  item.UpdatedAt.Format(time.RFC3339Nano),
		Username:   item.Username,
		Nickname:   item.Nickname,
		UserId:     item.UserId,
		DepId:      item.DepId,
		Dep:        item.Dep,
		Manager:    item.Manager,
		Avatar:     item.Avatar,
		Active:     item.Active,
		Tel:        item.Tel,
		Email:      item.Email,
		DataSource: ss,
		Disable:    item.Disable,
		DdId:       item.DdId,
		FsId:       item.FsId,
	}
}
