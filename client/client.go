package client

import (
	"bufio"
	"fmt"
	"godis/log"
	"godis/parser"
	"godis/protocol"
	"io"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/peterh/liner"
)

type Client struct {
	Conn      net.Conn
	Line      *liner.State
	OsSignals chan os.Signal
	closeChan chan struct{}
}

func NewClient(addr string) (*Client, error) {

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	return &Client{
		Conn:      conn,
		Line:      liner.NewLiner(),
		OsSignals: make(chan os.Signal, 1),
		closeChan: make(chan struct{}, 1),
	}, err
}

func Run(addr string) {
	cli, err := NewClient(addr)
	if err != nil {
		log.Errorln(err.Error())
	}
	go cli.Read()
	_ = cli.run()
}

// TODO client在server断联的情况下不能gently quit，因为c.Read协程没有退出
func (c *Client) run() error {
	signal.Notify(c.OsSignals, syscall.SIGTERM, syscall.SIGINT)
	go c.Read()
	for {
		select {
		case <-c.OsSignals:
			c.exit()
			close(c.closeChan)
			return nil
		case <-c.closeChan:
			_ = c.Line.Close()
			return nil
		default:
			cmd, e := c.Line.Prompt("> ")
			if e != nil {
				c.exit()
				return e
			}
			err := c.Execute(cmd)
			if err != nil {
				log.Errorln(err)
			}
		}
	}
}

func (c *Client) exit() {
	close(c.closeChan)
}

func (c *Client) Execute(cmd string) error {
	cmd = strings.TrimSpace(cmd)
	cmd = strings.ToLower(cmd)
	if strings.HasPrefix(cmd, "quit") {
		c.exit()
		return nil
	}
	_, err := c.Conn.Write(stringToRESP(cmd).ToBytes())
	return err
}

func (c *Client) Read() error {
	rd := bufio.NewReader(c.Conn)
	defer c.Conn.Close()
	ch := parser.ParseStream(rd)
	for payload := range ch {
		fmt.Println(payload.Data.ToString())
		if payload.Err != nil {
			if payload.Err != io.EOF {
				fmt.Println(payload.Err.Error())
			}
			fmt.Print("> ")
			return payload.Err
		}
		fmt.Print("> ")
	}
	return nil
}

// 这里是将所有的输入的设定为字符串，并将该字符串转换成protocol.MultiBulkReply
// TODO 如果后续出现需要再数组中插入的不同类型的输入则想办法修改成生成protocol.MultiRawReply
func stringToRESP(line string) protocol.Reply {
	words := strings.Split(line, " ")
	bytes := make([][]byte, 0)
	for _, word := range words {
		bytes = append(bytes, []byte(word))
	}
	return protocol.MakeMultiBulkReply(bytes)
}
