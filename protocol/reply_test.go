package protocol_test

import (
	"fmt"
	"godis/protocol"
	"testing"
)

/*
	resp协议可以将byte类型的输入，转换成resp字符串。
	通过不同类型reply来进行。
*/

// MultiBulkReply只负责全是字符串数组的输入
func TestMultiBulkReply(t *testing.T) {
	source := [][]byte{
		[]byte("set"),
		[]byte("key"),
		[]byte("value"),
	}
	fmt.Println(string(protocol.MakeMultiBulkReply(source).ToBytes()))
	//expect
	// *3
	// $3
	// set
	// $3
	// key
	// $5
	// value
}

func TestMakeMultiRawReply(t *testing.T) {
	source := [][]byte{
		[]byte("set"),
		[]byte("key"),
		[]byte("value"),
	}
	replys := make([]protocol.Reply, 0)
	replys = append(replys, protocol.MakeBulkReply([]byte("123")))
	replys = append(replys, protocol.MakeIntReply(200))
	replys = append(replys, protocol.MakeErrReply("xxxx"))
	replys = append(replys, protocol.MakeMultiBulkReply(source))
	fmt.Println(string(protocol.MakeMultiRawReply(replys).ToBytes()))
	// *4
	// $3
	// 123
	// :200
	// -xxxx
	// *3
	// $3
	// set
	// $3
	// key
	// $5
	// value
}
