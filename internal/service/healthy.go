package service

import (
	"context"
	"database/sql"
	"fmt"

	healthypb "codo-notice/pb/healthy"

	"github.com/go-redis/redis/v8"
)

type HealthyService struct {
	db    *sql.DB
	redis *redis.Client
}

func NewHealthyService(db *sql.DB, redisClient *redis.Client) *HealthyService {
	return &HealthyService{
		db:    db,
		redis: redisClient,
	}
}

func (x *HealthyService) Healthy(ctx context.Context, request *healthypb.HealthyRequest) (*healthypb.HealthyReply, error) {
	err := x.db.Ping()
	if err != nil {
		return nil, fmt.Errorf("mysql ping failed: %w", err)
	}
	err = x.redis.Ping(ctx).Err()
	if err != nil {
		return nil, fmt.Errorf("redis ping failed: %w", err)
	}
	return &healthypb.HealthyReply{
		Mysql: "yes",
		Redis: "yes",
	}, nil
}
