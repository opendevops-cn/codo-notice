package service

import (
	"github.com/google/wire"
)

// ProviderSet is service providers.
var ProviderSet = wire.NewSet(
	NewChannelService,
	NewHealthyService,
	NewRouterService,
	NewTemplateService,
	NewUserService,
	NewHookService,
)
