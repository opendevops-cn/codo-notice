package server

import (
	"context"
	"net"
	"net/http"
	"sync/atomic"

	"codo-notice/internal/conf"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type PrometheusServer struct {
	started  uint32
	listener net.Listener

	conf *conf.Server_Prometheus
}

func NewPrometheusServer(bc *conf.Bootstrap) (*PrometheusServer, error) {
	c := bc.Server
	svr := &PrometheusServer{
		conf: c.Prometheus,
	}
	addr := svr.conf.GetAddr()
	if svr.conf.GetEnable() {
		listener, err := net.Listen(svr.conf.GetNetwork(), addr)
		if err != nil {
			return nil, err
		}

		svr.listener = listener
	}
	return svr, nil
}

func (x *PrometheusServer) Start(ctx context.Context) error {
	if x.listener == nil {
		return nil
	}
	if !atomic.CompareAndSwapUint32(&x.started, 0, 1) {
		return nil
	}

	// 注册路由
	handler := http.NewServeMux()
	handler.Handle("/metrics", promhttp.HandlerFor(
		prometheus.DefaultGatherer,
		promhttp.HandlerOpts{
			EnableOpenMetrics: true,
		},
	))

	return http.Serve(x.listener, handler)
}

func (x *PrometheusServer) Stop(ctx context.Context) error {
	if atomic.CompareAndSwapUint32(&x.started, 1, 0) {
		return x.listener.Close()
	}
	return nil
}
