package main

import (
	"context"
	"flag"
	"os"

	"codo-notice/internal/conf"
	"codo-notice/internal/server"

	"github.com/opendevops-cn/codo-golang-sdk/config"
	loggersdk "github.com/opendevops-cn/codo-golang-sdk/logger"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/registry"
	"github.com/go-kratos/kratos/v2/transport/http"

	_ "go.uber.org/automaxprocs"
)

// go build -ldflags "-X main.Version=x.y.z"
var (
	// Version is the version of the compiled software.
	Version string
	// flagconf is the config flag.
	flagconf string

	id, _ = os.Hostname()
)

func init() {
	flag.StringVar(&flagconf, "conf", "conf/config.yaml", "config path, eg: -conf config.yaml")
}

func newApp(bc *conf.Bootstrap, logger log.Logger, hs *http.Server, pprof *server.PprofServer,
	prom *server.PrometheusServer, r registry.Registrar, hookServer *server.ThirdPartHookServer,
) (*kratos.App, error) {
	meta := bc.Metadata

	return kratos.New(
		kratos.ID(id),
		kratos.Name(meta.Name),
		kratos.Version(Version),
		kratos.Metadata(map[string]string{}),
		kratos.Logger(logger),
		kratos.Server(
			hs,
			pprof,
			prom,
			hookServer,
		),
		kratos.Registrar(r),
	), nil
}

func main() {
	flag.Parse()

	var bc conf.Bootstrap
	err := config.LoadConfig(
		&bc,
		config.WithYaml(flagconf),
		config.WithEnv("CODO"),
	)
	if err != nil {
		panic(err)
	}

	// 初始化日志组件
	logger, err := loggersdk.NewLogger(
		func(logConfig *loggersdk.LogConfig) {
			logConfig.Level = bc.Otel.Log.Level
		},
	)
	if err != nil {
		panic(err)
	}
	loggersdk.SetLogger(logger)

	lh := loggersdk.NewHelper(logger)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	lh.Debugf(ctx, "config loaded ====== %+v", &bc)

	app, cleanup, err := wireApp(ctx, logger, &bc)
	if err != nil {
		panic(err)
	}
	defer cleanup()

	// start and wait for stop signal
	if err := app.Run(); err != nil {
		panic(err)
	}
}
