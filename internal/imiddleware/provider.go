package imiddleware

import "github.com/google/wire"

var ProviderSet = wire.NewSet(NewJWTMiddleware)
