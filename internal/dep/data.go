package dep

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	"codo-notice/internal/conf"

	"github.com/ccheers/xpkg/sync/errgroup"
	"github.com/go-kratos/kratos/v2/log"
	mysql2 "github.com/go-sql-driver/mysql"
	loggersdk "github.com/opendevops-cn/codo-golang-sdk/logger"
	mysqlsdk "github.com/opendevops-cn/codo-golang-sdk/mysql"
	redissdk "github.com/opendevops-cn/codo-golang-sdk/redis"
	"go.opentelemetry.io/otel/metric"

	"github.com/go-redis/redis/v8"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

func NewRedis(ctx context.Context, logger loggersdk.Logger, bc *conf.Bootstrap, tp trace.TracerProvider, mp metric.MeterProvider) (*redis.Client, func(), error) {
	// 注意!!! 注入的 context 只是用于初始化，不要用于业务逻辑
	helper := loggersdk.NewHelper(logger)

	redisConf := bc.Data.Redis
	client, err := redissdk.NewRedisV8(func(c *redissdk.RedisConfig) {
		ss := strings.Split(redisConf.Addr, ":")
		if ss[0] != "" {
			c.Host = ss[0]
		}
		if len(ss) > 1 {
			port, _ := strconv.Atoi(ss[1])
			c.Port = uint32(port)
		}
		if redisConf.Password != "" {
			c.Pass = redisConf.Password
		}
		if redisConf.DialTimeout.AsDuration().Seconds() > 0 {
			c.DialTimeout = uint32(redisConf.DialTimeout.AsDuration().Seconds())
		}
		if redisConf.ReadTimeout.AsDuration().Seconds() > 0 {
			c.ReadTimeout = uint32(redisConf.ReadTimeout.AsDuration().Seconds())
		}
		if redisConf.WriteTimeout.AsDuration().Seconds() > 0 {
			c.WriteTimeout = uint32(redisConf.WriteTimeout.AsDuration().Seconds())
		}

		c.TracerProvider = tp
		c.MeterProvider = mp
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create the redis: %w", err)
	}

	err = client.Ping(ctx).Err()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to ping the redis: %w", err)
	}

	ctx, cancel := context.WithCancelCause(context.Background())
	eg := errgroup.WithCancel(ctx)
	eg.Go(func(ctx context.Context) error {
		for {
			select {
			case <-ctx.Done():
				return nil
			case <-time.After(3 * time.Second):
			}
			pingCtx, cancel := context.WithTimeout(ctx, redisConf.DialTimeout.AsDuration())
			client.Ping(pingCtx)
			cancel()
		}
	})

	return client, func() {
		cancel(fmt.Errorf("server shutdown"))
		_ = eg.Wait()

		err = client.Close()
		if err != nil {
			helper.Errorf(ctx, "failed to close the redis: %v", err)
		}
	}, nil
}

func NewMysql(ctx context.Context, bc *conf.Bootstrap, logger log.Logger, tp trace.TracerProvider, mp metric.MeterProvider) (*sql.DB, func(), error) {
	helper := log.NewHelper(logger)
	mysqlConf := bc.Data.Database
	dsnCfg, err := mysql2.ParseDSN(mysqlConf.Link)
	if err != nil {
		return nil, nil, err
	}

	sqlDB, closeDB, err := mysqlsdk.NewMysql(func(c *mysqlsdk.DBConfig) {
		ss := strings.Split(dsnCfg.Addr, ":")
		if ss[0] != "" {
			c.Host = ss[0]
		}
		if len(ss) > 1 {
			port, _ := strconv.Atoi(ss[1])
			c.Port = uint32(port)
		}
		if dsnCfg.User != "" {
			c.User = dsnCfg.User
		}
		if dsnCfg.Passwd != "" {
			c.Pass = dsnCfg.Passwd
		}
		if dsnCfg.DBName != "" {
			c.DBName = dsnCfg.DBName
		}
		if mysqlConf.MaxLifetime.AsDuration().Seconds() > 0 {
			c.ConnMaxLifetime = uint32(mysqlConf.MaxLifetime.AsDuration().Seconds())
		}
		if mysqlConf.MaxIdleConns > 0 {
			c.MaxIdleConns = mysqlConf.MaxIdleConns
		}
		if mysqlConf.MaxOpenConns > 0 {
			c.MaxOpenConns = mysqlConf.MaxOpenConns
		}

		c.TracerProvider = tp
		c.MeterProvider = mp
	})
	if err != nil {
		return nil, nil, err
	}
	err = sqlDB.PingContext(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to ping the database: %w", err)
	}
	sqlDB.SetConnMaxLifetime(mysqlConf.MaxLifetime.AsDuration())
	sqlDB.SetMaxIdleConns(int(mysqlConf.MaxIdleConns))
	sqlDB.SetMaxOpenConns(int(mysqlConf.MaxOpenConns))

	return sqlDB, func() {
		helper.Info("closing the mysql")
		closeDB()
	}, nil
}

func NewGORM(bc *conf.Bootstrap, sqlDB *sql.DB) (*gorm.DB, error) {
	mysqlConf := bc.Data.Database

	gormDB, err := gorm.Open(mysql.New(mysql.Config{
		Conn: sqlDB,
	}), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   mysqlConf.Prefix,
			SingularTable: true,
		},
	})
	if err != nil {
		return nil, err
	}

	if mysqlConf.Debug {
		gormDB = gormDB.Debug()
	}

	return gormDB, nil
}
