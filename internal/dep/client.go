package dep

import (
	"github.com/ccheers/xpkg/sync/try_lock"
	"github.com/go-redis/redis/v8"
	"github.com/opendevops-cn/codo-golang-sdk/client/xhttp"
	"github.com/opendevops-cn/codo-golang-sdk/tools/cascmd"
)

func NewXHTTPClient() (xhttp.IClient, error) {
	return xhttp.NewClient()
}

func NewRedisCas(client *redis.Client) try_lock.CASCommand {
	return cascmd.NewCasCmd(client)
}
