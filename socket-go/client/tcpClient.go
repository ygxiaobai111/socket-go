package client

import (
	"io"
	"log"
	"net"
	"socket/protocol"
)

type TcpChatClient struct {
	conn      net.Conn
	cmdReader *protocol.CommandReader
	cmdWriter *protocol.CommandWriter
	name      string
	incoming  chan protocol.MessageCommand
}

func NewClient() *TcpChatClient {
	return &TcpChatClient{
		incoming: make(chan protocol.MessageCommand),
	}
}

// 建立连接并创建通讯协议的r和w
func (c *TcpChatClient) Dial(address string) error {
	conn, err := net.Dial("tcp", address)
	if err == nil {
		c.conn = conn
	}
	c.cmdReader = protocol.NewCommandReader(conn)
	c.cmdWriter = protocol.NewCommandWriter(conn)
	return err
}

// 将消息发送至服务端
func (c *TcpChatClient) Send(command interface{}) error {
	return c.cmdWriter.Write(command)
}

// 监听服务器端广播信息，并将它们发送回chan
func (c *TcpChatClient) Start() {
	for {
		cmd, err := c.cmdReader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			log.Printf("Read error %v", err)
		}
		if cmd != nil {
			switch v := cmd.(type) {
			case protocol.MessageCommand:
				c.incoming <- v
			default:
				log.Printf("Unknown command: %v", v)
			}
		}
	}
}

func (c *TcpChatClient) Close() {
	c.conn.Close()
}

func (c *TcpChatClient) Incoming() chan protocol.MessageCommand {
	return c.incoming
}

func (c *TcpChatClient) SetName(name string) error {
	return c.Send(protocol.NameCommand{name})
}

func (c *TcpChatClient) SendMessage(message string) error {
	return c.Send(protocol.SendCommand{
		Message: message,
	})
}
