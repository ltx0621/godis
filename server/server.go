package server

import (
	"context"
	"fmt"
	"godis/handler"
	"godis/log"
	"godis/session"
	"net"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
)

type Server struct {
	sessions sync.Map
	wg       sync.WaitGroup
}

var defaultServer Server = *NewServer()

func NewServer() *Server {
	return &Server{
		sessions: sync.Map{},
		wg:       sync.WaitGroup{},
	}
}

func ListenAndServe(address string) {
	defaultServer.listenAndServe(address, handler.RedisHandler)
}

func (s *Server) listenAndServe(address string, handler handler.Handler) error {
	closeChan := make(chan struct{})
	sigCh := make(chan os.Signal)
	signal.Notify(sigCh, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	// 监控系统信号，中断时进行退出
	go func() {
		sig := <-sigCh
		switch sig {
		case syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			closeChan <- struct{}{}
		}
	}()
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}
	log.Infoln(fmt.Sprintf("bind: %s, start listening...", address))
	//接受closeChan信号，停止服务，安全退出
	go func() {
		<-closeChan
		log.Infoln("accept close signal,closing")
		listener.Close()
		s.Close()
	}()

	defer func() {
		s.Close()
		listener.Close()
	}()
	for {
		conn, err := listener.Accept()
		if err != nil {
			if !strings.Contains(err.Error(), "use of closed network connection") {
				log.Errorln(err.Error())
			}
			break
		}
		s.wg.Add(1)
		ctx, cancel := context.WithCancel(context.Background())
		sess := &session.Session{
			Conn:       conn,
			Ctx:        ctx,
			Cancel:     cancel,
			RemoteAddr: conn.RemoteAddr().String(),
		}
		s.sessions.Store(sess, struct{}{})
		go func() {
			defer func() {
				s.sessions.Delete(sess)
				s.wg.Done()
			}()
			s.Handle(sess, handler)
		}()
	}
	s.wg.Wait()
	return nil
}

func (s *Server) Close() {
	s.sessions.Range(func(key, value any) bool {
		ses := key.(*session.Session)
		_ = ses.Conn.Close()
		ses.Cancel()
		return true
	})
}

func (s *Server) Handle(sess *session.Session, handler handler.Handler) {
	log.Infoln("accept request from ", sess.Conn.RemoteAddr())
	//TODO 这里好像没啥用，需要修改
	select {
	case <-sess.Ctx.Done():
		return
	default:
		handler.Handle(sess)
	}
}
