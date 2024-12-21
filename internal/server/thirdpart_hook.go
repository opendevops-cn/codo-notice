package server

import (
	"context"

	"codo-notice/internal/conf"
	"codo-notice/internal/service"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/logging"
	"github.com/go-kratos/kratos/v2/middleware/metrics"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	"github.com/go-kratos/kratos/v2/transport/http"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

type ThirdPartHookServer struct {
	svr *http.Server
}

func NewThirdPartHookServer(bc *conf.Bootstrap, logger log.Logger,
	mp metric.MeterProvider, tp trace.TracerProvider,
	service *service.HookService,
) (*ThirdPartHookServer, error) {
	c := bc.Server

	meter := mp.Meter("server.thirdpart_hook")

	counter, err := metrics.DefaultRequestsCounter(meter, metrics.DefaultServerRequestsCounterName)
	if err != nil {
		return nil, err
	}
	seconds, err := metrics.DefaultSecondsHistogram(meter, metrics.DefaultServerSecondsHistogramName)
	if err != nil {
		return nil, err
	}

	opts := []http.ServerOption{
		http.Middleware(
			recovery.Recovery(),
			tracing.Server(
				tracing.WithTracerProvider(tp),
			),
			logging.Server(logger),
			metrics.Server(
				metrics.WithRequests(counter),
				metrics.WithSeconds(seconds),
			),
		),
	}

	tph := c.ThirdPartHook
	if tph.Network != "" {
		opts = append(opts, http.Network(tph.Network))
	}
	if c.ThirdPartHook.Addr != "" {
		opts = append(opts, http.Address(tph.Addr))
	}
	if tph.Timeout != nil {
		opts = append(opts, http.Timeout(tph.Timeout.AsDuration()))
	}
	svr := http.NewServer(opts...)

	// 注册路由信息
	routes := service.Routes()
	router := svr.Route("/")
	for _, route := range routes {
		router.Handle(route.Method, route.Path, route.Handle)
	}

	return &ThirdPartHookServer{
		svr: svr,
	}, nil
}

func (x *ThirdPartHookServer) Start(ctx context.Context) error {
	return x.svr.Start(ctx)
}

func (x *ThirdPartHookServer) Stop(ctx context.Context) error {
	return x.svr.Stop(ctx)
}
