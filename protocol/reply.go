package protocol

import (
	"bytes"
	"strconv"
)

var (

	// CRLF is the line separator of redis serialization protocol
	CRLF = "\r\n"
)

type Reply interface {
	ToBytes() []byte
	ToString() string
}

var _ Reply = (*BulkReply)(nil)
var _ Reply = (*IntReply)(nil)
var _ Reply = (*MultiBulkReply)(nil)
var _ Reply = (*MultiRawReply)(nil)
var _ Reply = (*StandardErrReply)(nil)
var _ Reply = (*StatusReply)(nil)

/* ---- Bulk Reply ---- */

// BulkReply stores a binary-safe string
type BulkReply struct {
	Arg []byte
}

// MakeBulkReply creates  BulkReply
func MakeBulkReply(arg []byte) *BulkReply {
	return &BulkReply{
		Arg: arg,
	}
}

// ToBytes marshal redis.Reply
func (r *BulkReply) ToBytes() []byte {
	if r.Arg == nil {
		return nullBulkBytes
	}
	return []byte("$" + strconv.Itoa(len(r.Arg)) + CRLF + string(r.Arg) + CRLF)
}

func (r *BulkReply) ToString() string {
	return string(r.Arg)
}

/* ---- Multi Bulk Reply ---- */

// MultiBulkReply stores a list of string
type MultiBulkReply struct {
	Args [][]byte
}

// MakeMultiBulkReply creates MultiBulkReply
func MakeMultiBulkReply(args [][]byte) *MultiBulkReply {
	return &MultiBulkReply{
		Args: args,
	}
}

// ToBytes marshal redis.Reply
func (r *MultiBulkReply) ToBytes() []byte {
	argLen := len(r.Args)
	var buf bytes.Buffer
	buf.WriteString("*" + strconv.Itoa(argLen) + CRLF)
	for _, arg := range r.Args {
		if arg == nil {
			buf.WriteString("$-1" + CRLF)
		} else {
			buf.WriteString("$" + strconv.Itoa(len(arg)) + CRLF + string(arg) + CRLF)
		}
	}
	return buf.Bytes()
}

func (r *MultiBulkReply) ToString() string {
	s := ""
	if len(r.Args) == 0 {
		return s
	}
	for _, arg := range r.Args {
		s += string(arg)
		s += " "
	}
	return s[:len(s)-1]

}

/* ---- Multi Raw Reply ---- */

// MultiRawReply store complex list structure, for example GeoPos command
type MultiRawReply struct {
	Replies []Reply
}

// MakeMultiRawReply creates MultiRawReply
func MakeMultiRawReply(replies []Reply) *MultiRawReply {
	return &MultiRawReply{
		Replies: replies,
	}
}

// ToBytes marshal redis.Reply
func (r *MultiRawReply) ToBytes() []byte {
	argLen := len(r.Replies)
	var buf bytes.Buffer
	buf.WriteString("*" + strconv.Itoa(argLen) + CRLF)
	for _, arg := range r.Replies {
		buf.Write(arg.ToBytes())
	}
	return buf.Bytes()
}

func (r *MultiRawReply) ToString() string {
	argLen := len(r.Replies)
	s := ""
	if argLen == 0 {
		return s
	}
	for _, reply := range r.Replies {
		s += reply.ToString()
		s += " "
	}
	return s[:len(s)-1]
}

/* ---- Status Reply ---- */

// StatusReply stores a simple status string
type StatusReply struct {
	Status string
}

// MakeStatusReply creates StatusReply
func MakeStatusReply(status string) *StatusReply {
	return &StatusReply{
		Status: status,
	}
}

// ToBytes marshal redis.Reply
func (r *StatusReply) ToBytes() []byte {
	return []byte("+" + r.Status + CRLF)
}

func (r *StatusReply) ToString() string {
	return r.Status
}

// IsOKReply returns true if the given protocol is +OK
func IsOKReply(reply Reply) bool {
	return string(reply.ToBytes()) == "+OK\r\n"
}

/* ---- Int Reply ---- */

// IntReply stores an int64 number
type IntReply struct {
	Code int64
}

// MakeIntReply creates int protocol
func MakeIntReply(code int64) *IntReply {
	return &IntReply{
		Code: code,
	}
}

// ToBytes marshal redis.Reply
func (r *IntReply) ToBytes() []byte {
	return []byte(":" + strconv.FormatInt(r.Code, 10) + CRLF)
}

func (r *IntReply) ToString() string {
	return string(r.Code)
}

/* ---- Error Reply ---- */

// // ErrorReply is an error and redis.Reply
// type ErrorReply interface {
//     ToString() string
//     ToBytes() []byte
// }

// StandardErrReply represents server error
type StandardErrReply struct {
	Status string
}

// MakeErrReply creates StandardErrReply
func MakeErrReply(status string) *StandardErrReply {
	return &StandardErrReply{
		Status: status,
	}
}

// IsErrorReply returns true if the given protocol is error
func IsErrorReply(reply Reply) bool {
	return reply.ToBytes()[0] == '-'
}

// ToBytes marshal redis.Reply
func (r *StandardErrReply) ToBytes() []byte {
	return []byte("-" + r.Status + CRLF)
}

func (r *StandardErrReply) ToString() string {
	return r.Status
}
