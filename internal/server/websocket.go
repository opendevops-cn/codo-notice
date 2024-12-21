package server

import (
	"codo-notice/internal/conf"

	"github.com/tx7do/kratos-transport/transport/websocket"
)

func NewWsServer(bc *conf.Bootstrap, handler websocket.MessageHandler, binder websocket.Binder) *websocket.Server {
	wsConf := bc.Server.Websocket
	srv := websocket.NewServer(
		websocket.WithNetwork(wsConf.Network),
		websocket.WithAddress(wsConf.Addr),
		websocket.WithTimeout(wsConf.Timeout.AsDuration()),
	)

	srv.RegisterMessageHandler(websocket.PayloadTypeText, handler, binder)
	return srv
}
