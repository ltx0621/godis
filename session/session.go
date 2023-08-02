package session

import (
	"context"
	"net"
)

type Session struct {
	Conn       net.Conn
	Ctx        context.Context
	Cancel     context.CancelFunc
	RemoteAddr string
}
