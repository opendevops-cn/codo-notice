package dep

import (
	"context"
	"time"

	"codo-notice/internal/conf"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/registry"
	"github.com/opendevops-cn/codo-golang-sdk/adapter/kratos/discovery/dummy"
	"github.com/opendevops-cn/codo-golang-sdk/adapter/kratos/discovery/etcd"

	clientv3 "go.etcd.io/etcd/client/v3"
)

func NewRegister(ctx context.Context, logger log.Logger, bc *conf.Bootstrap) (registry.Registrar, error) {
	registryConf := bc.EtcdRegistry
	if !registryConf.Enabled {
		return dummy.NewRegistrar(), nil
	}
	client, err := clientv3.New(clientv3.Config{
		Endpoints:           registryConf.Endpoints,
		AutoSyncInterval:    time.Minute,
		DialTimeout:         5 * time.Second,
		RejectOldCluster:    true,
		PermitWithoutStream: true,
	})
	if err != nil {
		return nil, err
	}
	return etcd.New(client, logger, etcd.Context(ctx)), nil
}
