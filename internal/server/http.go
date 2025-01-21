package server

import (
	"codo-notice/internal/conf"
	"codo-notice/internal/imiddleware"
	"codo-notice/internal/service"
	channelpb "codo-notice/pb/channel"
	healthypb "codo-notice/pb/healthy"
	routerpb "codo-notice/pb/router"
	templatespb "codo-notice/pb/templates"
	userpb "codo-notice/pb/user"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/logging"
	"github.com/go-kratos/kratos/v2/middleware/metrics"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/transport/http"
	"github.com/opendevops-cn/codo-golang-sdk/adapter/kratos/middleware/ktracing"
	"github.com/opendevops-cn/codo-golang-sdk/transport/chttp"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

// NewHTTPServer new an HTTP server.
func NewHTTPServer(bc *conf.Bootstrap, logger log.Logger,
	mp metric.MeterProvider, tp trace.TracerProvider,
	cs *service.ChannelService,
	rs *service.RouterService,
	ts *service.TemplateService,
	us *service.UserService,
	hs *service.HealthyService,
	jwtMiddleware *imiddleware.JWTMiddleware,
) (*http.Server, error) {
	c := bc.Server
	meter := mp.Meter("server.http")

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
			ktracing.Server(
				ktracing.WithTracerProvider(tp),
			),
			logging.Server(logger),
			metrics.Server(
				metrics.WithRequests(counter),
				metrics.WithSeconds(seconds),
			),
			channelpb.NewChannelHTTPServerMiddleware(
				channelpb.ChannelJWTMiddlewareMiddleware(jwtMiddleware.Server()),
			),
			routerpb.NewRouterHTTPServerMiddleware(
				routerpb.RouterJWTMiddlewareMiddleware(jwtMiddleware.Server()),
			),
			userpb.NewUserHTTPServerMiddleware(
				userpb.UserJWTMiddlewareMiddleware(jwtMiddleware.Server()),
			),
			templatespb.NewTemplatesHTTPServerMiddleware(
				templatespb.TemplatesJWTMiddlewareMiddleware(jwtMiddleware.Server()),
			),
		),
		http.ResponseEncoder(chttp.ResponseEncoder),
		http.ErrorEncoder(chttp.ErrorEncoder),
		http.RequestDecoder(chttp.RequestBodyDecoder),
	}
	if c.Http.Network != "" {
		opts = append(opts, http.Network(c.Http.Network))
	}
	if c.Http.Addr != "" {
		opts = append(opts, http.Address(c.Http.Addr))
	}
	if c.Http.Timeout != nil {
		opts = append(opts, http.Timeout(c.Http.Timeout.AsDuration()))
	}
	srv := http.NewServer(opts...)
	channelpb.RegisterChannelHTTPServer(srv, cs)
	routerpb.RegisterRouterHTTPServer(srv, rs)
	userpb.RegisterUserHTTPServer(srv, us)
	templatespb.RegisterTemplatesHTTPServer(srv, ts)
	healthypb.RegisterHealthyHTTPServer(srv, hs)
	return srv, nil
}
