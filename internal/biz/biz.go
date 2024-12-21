package biz

import "github.com/google/wire"

// ProviderSet is biz providers.
var ProviderSet = wire.NewSet(
	NewIChannelUseCase, NewChannelUseCase,
	NewIRouterUseCase, NewRouterUseCase,
	NewITemplateUseCase, NewTemplateUseCase,
	NewIUserUseCase, NewUserUseCase,
	NewIAlertUseCase, NewAlertUseCase,
	NewIHookUseCase, NewHookUseCase,
)
