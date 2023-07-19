package handler

import (
	"context"
	"net"
)

type Handler interface {
	Handle(context.Context, net.Conn)
	Close() error
}

type RedisHandler struct{}

func (r RedisHandler) Handle(ctx context.Context, conn net.Conn) {

}

func (r RedisHandler) Close() error {
	return nil
}

var _ Handler = (*RedisHandler)(nil)
