package data

import (
	"context"

	"codo-notice/internal/impl/data/models"

	"github.com/go-redis/redis/v8"
	"github.com/google/wire"
	"gorm.io/gorm"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(
	NewData,
	NewUserRepo, NewIUserRepo,
	NewTemplateRepo, NewITemplateRepo,
	NewRouterRepo, NewIRouterRepo,
	NewChannelRepo, NewIChannelRepo,
	NewILarkCardCallbackRepo, NewLarkCardCallbackRepo,
)

// Data .
type Data struct {
	db    *gorm.DB
	redis *redis.Client
}

// NewData .
func NewData(gormDB *gorm.DB, redisClient *redis.Client) (*Data, error) {
	err := models.Migrate(gormDB)
	if err != nil {
		return nil, err
	}
	// 初始化数据库连接
	return &Data{
		db:    gormDB,
		redis: redisClient,
	}, nil
}

func (x *Data) DBWithContext(ctx context.Context) *gorm.DB {
	return x.db.WithContext(ctx)
}
