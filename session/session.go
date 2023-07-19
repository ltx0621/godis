package session

import (
	"context"
	"godis/handler"
	"net"
)

type Session struct {
	Conn    net.Conn
	Ctx     context.Context
	Cancel  context.CancelFunc
	Handler handler.Handler
}
