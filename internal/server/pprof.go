package server

import (
	"context"
	"net"
	"net/http"
	_ "net/http/pprof"
	"sync/atomic"

	"codo-notice/internal/conf"
)

type PprofServer struct {
	conf *conf.Server_Pprof

	started  uint32
	listener net.Listener
}

func NewPprofServer(bc *conf.Bootstrap) (*PprofServer, error) {
	c := bc.Server
	svr := &PprofServer{
		conf: c.Pprof,
	}
	if svr.conf.GetEnable() {
		addr := svr.conf.GetAddr()
		listener, err := net.Listen(svr.conf.GetNetwork(), addr)
		if err != nil {
			return nil, err
		}
		svr.listener = listener
	}
	return svr, nil
}

func (x *PprofServer) Start(ctx context.Context) error {
	if x.listener == nil {
		return nil
	}
	if !atomic.CompareAndSwapUint32(&x.started, 0, 1) {
		return nil
	}
	return http.Serve(x.listener, nil)
}

func (x *PprofServer) Stop(ctx context.Context) error {
	if atomic.CompareAndSwapUint32(&x.started, 1, 0) {
		return x.listener.Close()
	}
	return nil
}
