package dep

import "github.com/google/wire"

var ProviderSet = wire.NewSet(
	NewRedis, NewMysql, NewGORM,
	NewMeterProvider, NewTracerProvider, NewTextMapPropagator,
	NewLogger, NewRegister, NewXHTTPClient, NewRedisCas,
	NewMsgbus, NewTopicManager, NewSharedStorage, NewOtelOptions,
)
