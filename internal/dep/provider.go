package dep

import "github.com/google/wire"

var ProviderSet = wire.NewSet(
	NewRedis, NewMysql, NewMeterProvider, NewTracerProvider, NewTextMapPropagator,
	NewLogger, NewRegister, NewXHTTPClient, NewRedisCas,
	NewMsgbus, NewTopicManager, NewSharedStorage, NewOtelOptions,
)
