package impl

import (
	"codo-notice/internal/impl/alerts"
	"codo-notice/internal/impl/data"

	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(data.ProviderSet, alerts.ProviderSet)
