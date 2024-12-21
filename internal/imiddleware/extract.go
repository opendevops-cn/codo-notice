package imiddleware

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-kratos/kratos/v2/transport"
	khttp "github.com/go-kratos/kratos/v2/transport/http"
)

var (
	ErrKratosTransportNotFound         = fmt.Errorf("kratos: transport not found in ctx")
	ErrKratosHTTPContextNotFound       = fmt.Errorf("kratos: http.context not found in ctx")
	ErrKratosTransportNotHTTPTransport = fmt.Errorf("kratos: transport not a http.Transporter")
)

func ExtraHTTPRequestFromKratosContext(ctx context.Context) (*http.Request, error) {
	httpTr, err := extraKratosHTTPTransport(ctx)
	if err != nil {
		return nil, err
	}
	return httpTr.Request(), nil
}

func extraKratosHTTPTransport(ctx context.Context) (khttp.Transporter, error) {
	tr, ok := transport.FromServerContext(ctx)
	if !ok {
		return nil, ErrKratosTransportNotFound
	}
	httpTr, ok := tr.(khttp.Transporter)
	if !ok {
		return nil, ErrKratosTransportNotHTTPTransport
	}
	return httpTr, nil
}

func extraKratosHTTPContext(ctx context.Context) (khttp.Context, error) {
	httpCtx, ok := ctx.(khttp.Context)
	if !ok {
		return nil, ErrKratosHTTPContextNotFound
	}
	return httpCtx, nil
}
