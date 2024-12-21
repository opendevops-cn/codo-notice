package dep

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"codo-notice/internal/conf"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-redis/redis/extra/redisotel/v8"
	"github.com/go-redis/redis/v8"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

func NewRedis(ctx context.Context, logger log.Logger, bc *conf.Bootstrap, provider trace.TracerProvider) (*redis.Client, func(), error) {
	helper := log.NewHelper(logger)

	redisConf := bc.Data.Redis
	client := redis.NewClient(&redis.Options{
		Network:      redisConf.Network,
		Addr:         redisConf.Addr,
		Password:     redisConf.Password,
		DB:           int(redisConf.Db),
		MaxRetries:   5,
		DialTimeout:  redisConf.DialTimeout.AsDuration(),
		ReadTimeout:  redisConf.ReadTimeout.AsDuration(),
		WriteTimeout: redisConf.WriteTimeout.AsDuration(),
		PoolSize:     128,
	})

	err := client.Ping(ctx).Err()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to ping the redis: %w", err)
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(3 * time.Second):
			}
			pingCtx, cancel := context.WithTimeout(ctx, redisConf.DialTimeout.AsDuration())
			client.Ping(pingCtx)
			cancel()
		}
	}()

	client.AddHook(redisotel.NewTracingHook(redisotel.WithTracerProvider(provider)))

	return client, func() {
		err = client.Close()
		if err != nil {
			helper.Errorf("failed to close the redis: %v", err)
		}
	}, nil
}

func NewMysql(bc *conf.Bootstrap, logger log.Logger, _ trace.TracerProvider) (*gorm.DB, func(), error) {
	helper := log.NewHelper(logger)

	// database/gdb/gdb_core_underlying.go:176 直接拿的全局 trace provider ， 需要保证全局 trace provider 已经初始化, 所以引入
	mysqlConf := bc.Data.Database
	sqlDB, err := sql.Open("mysql", mysqlConf.Link)
	if err != nil {
		return nil, nil, err
	}
	err = sqlDB.Ping()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to ping the database: %w", err)
	}
	sqlDB.SetConnMaxLifetime(mysqlConf.MaxLifetime.AsDuration())
	sqlDB.SetMaxIdleConns(int(mysqlConf.MaxIdleConns))
	sqlDB.SetMaxOpenConns(int(mysqlConf.MaxOpenConns))

	gormDB, err := gorm.Open(mysql.New(mysql.Config{
		Conn: sqlDB,
	}), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   mysqlConf.Prefix,
			SingularTable: true,
		},
	})
	if err != nil {
		sqlDB.Close()
		return nil, nil, err
	}

	if mysqlConf.Debug {
		gormDB = gormDB.Debug()
	}
	return gormDB, func() {
		helper.Info("closing the mysql")
		err := sqlDB.Close()
		if err != nil {
			helper.Errorf("failed to close the database: %v", err)
		}
	}, nil
}
