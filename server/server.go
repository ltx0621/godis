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
	"sync"
	"syscall"
)

type Server struct {
	sessions []session.Session
	wg       sync.WaitGroup
}

var defaultServer Server = *NewServer()

func NewServer() *Server {
	return &Server{
		sessions: []session.Session{},
		wg:       sync.WaitGroup{},
	}
}

func ListenAndServe(address string) {
	defaultServer.listenAndServe(address)
}

func (s *Server) listenAndServe(address string) error {
	// 监控系统信号，中断时进行退出
	closeChan := make(chan struct{})
	sigCh := make(chan os.Signal)
	signal.Notify(sigCh, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
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
			log.Errorln(err.Error())
			break
		}
		s.wg.Add(1)
		ctx, cancel := context.WithCancel(context.Background())
		redisHandler := handler.RedisHandler{}
		sess := session.Session{
			Conn:    conn,
			Ctx:     ctx,
			Cancel:  cancel,
			Handler: redisHandler,
		}
		s.sessions = append(s.sessions, sess)
		go func() {
			defer s.wg.Done()
			s.Handle(&sess)
		}()
	}
	s.wg.Wait()
	return nil
}

func (s *Server) Close() {
	for _, ss := range s.sessions {
		ss.Conn.Close()
		ss.Cancel()
	}
}

func (s *Server) Handle(sess *session.Session) {
	log.Infoln("accept request from ", sess.Conn.RemoteAddr())
	for {
		select {
		case <-sess.Ctx.Done():
			return
		default:
			sess.Handler.Handle(sess.Ctx, sess.Conn)
		}
	}
}
