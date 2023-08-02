package handler

import (
	"godis/log"
	"godis/parser"
	"godis/protocol"
	"godis/session"
	"godis/store/inmem"
	"io"
	"strings"
)

type Handler interface {
	Handle(*session.Session)
	Close() error
}

type redisHandler struct {
	db *inmem.Inmem
}

var _ Handler = (*redisHandler)(nil)
var RedisHandler = redisHandler{
	db: inmem.NewInmem(),
}

func (r redisHandler) Handle(sess *session.Session) {
	ch := parser.ParseStream(sess.Conn)
	for payload := range ch {
		if payload.Err != nil {
			if payload.Err == io.EOF ||
				payload.Err == io.ErrUnexpectedEOF ||
				strings.Contains(payload.Err.Error(), "use of closed network connection") {
				// connection closed
				sess.Cancel()
				_ = sess.Conn.Close()
				log.Infoln("connection closed: " + sess.RemoteAddr)
				return
			}
			// protocol err
			errReply := protocol.MakeErrReply(payload.Err.Error())
			_, err := sess.Conn.Write(errReply.ToBytes())
			if err != nil {
				sess.Cancel()
				_ = sess.Conn.Close()
				log.Infoln("connection closed: " + sess.RemoteAddr)
				return
			}
			continue
		}
		if payload.Data == nil {
			log.Errorln("empty payload")
			sess.Conn.Write(protocol.MakeErrReply("empty payload").ToBytes())
			continue
		}
		cmd, ok := payload.Data.(*protocol.MultiBulkReply)
		if !ok {
			log.Errorln("require multi bulk protocol")
			continue
		}
		result := r.Exec(cmd.Args)
		if result != nil {
			_, _ = sess.Conn.Write(result.ToBytes())
		} else {
			_, _ = sess.Conn.Write(protocol.MakeErrReply("unknow cmd").ToBytes())
		}
		// switch typ := reflect.TypeOf(payload.Data){
		// case
		// }
		// sess.Conn.Write(payload.Data.ToBytes())
	}
}

func (r redisHandler) Exec(args [][]byte) protocol.Reply {
	cmdName := strings.ToLower(string(args[0]))
	switch cmdName {
	case "set":
		return r.Insert(args[1:])
	case "key":
		return r.Find(args[1:])
	}
	return nil

}

func (r redisHandler) Close() error {
	return nil
}

func (r redisHandler) Insert(args [][]byte) protocol.Reply {
	l := len(args)
	if l == 0 {
		return protocol.MakeErrReply("empty k/v")
	}
	if l%2 != 0 {
		return protocol.MakeErrReply("k/v nums don't be in pairs")
	}
	for i := 0; i < l; i = i + 2 {
		r.db.Insert(string(args[i]), string(args[i+1]))
	}
	return protocol.MakeStatusReply("insert ok")
}

func (r redisHandler) Find(args [][]byte) protocol.Reply {
	l := len(args)
	if l == 0 || l > 1 {
		return protocol.MakeErrReply("wrong num of input key")
	}
	v := r.db.Find(string(args[0]))
	if v == nil {
		return protocol.MakeErrReply("key not find")
	}
	switch vv := v.(type) {
	case int:
		return protocol.MakeIntReply(int64(vv))
	case string:
		return protocol.MakeBulkReply([]byte(vv))
	default:
		return protocol.MakeErrReply("not supported type")
	}
}
