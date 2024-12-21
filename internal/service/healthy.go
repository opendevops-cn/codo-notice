package service

import (
	"context"

	"codo-notice/pb/healthy"
)

type HealthyService struct{}

func NewHealthyService() *HealthyService {
	return &HealthyService{}
}

func (x *HealthyService) Healthy(ctx context.Context, request *healthy.HealthyRequest) (*healthy.HealthyReply, error) {
	return &healthy.HealthyReply{
		Mysql: "yes",
		Redis: "yes",
	}, nil
}
