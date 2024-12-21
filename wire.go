//go:build wireinject
// +build wireinject

// The build tag makes sure the stub is not built in the final build.

package main

import (
	"context"

	"codo-notice/internal/imiddleware"

	"codo-notice/internal/biz"
	"codo-notice/internal/conf"
	"codo-notice/internal/dep"
	"codo-notice/internal/impl"
	"codo-notice/internal/server"
	"codo-notice/internal/service"

	"github.com/go-kratos/kratos/v2"
	loggersdk "github.com/opendevops-cn/codo-golang-sdk/logger"

	"github.com/google/wire"
)

// wireApp init kratos application.
func wireApp(context.Context, loggersdk.Logger, *conf.Bootstrap) (*kratos.App, func(), error) {
	panic(wire.Build(server.ProviderSet, impl.ProviderSet, biz.ProviderSet,
		imiddleware.ProviderSet, service.ProviderSet, dep.ProviderSet, newApp))
}
