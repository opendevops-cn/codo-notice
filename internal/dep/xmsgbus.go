package dep

import (
	"context"

	"github.com/ccheers/xpkg/sync/try_lock"
	"github.com/ccheers/xpkg/xmsgbus"
	redisimpl "github.com/ccheers/xpkg/xmsgbus/impl/redis"
	"github.com/go-redis/redis/v8"
)

func NewMsgbus(client *redis.Client) xmsgbus.IMsgBus {
	return redisimpl.NewMsgBus(client)
}

func NewTopicManager(ctx context.Context, bus xmsgbus.IMsgBus,
	command try_lock.CASCommand, storage xmsgbus.ISharedStorage,
) xmsgbus.ITopicManager {
	return xmsgbus.NewTopicManager(ctx, bus, command, storage)
}

func NewSharedStorage(client *redis.Client) xmsgbus.ISharedStorage {
	return redisimpl.NewSharedStorage(client)
}

func NewOtelOptions() *xmsgbus.OTELOptions {
	return xmsgbus.NewOTELOptions()
}
