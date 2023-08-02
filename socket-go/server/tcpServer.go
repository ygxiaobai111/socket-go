package server

import (
	"io"
	"log"
	"net"
	"socket/protocol"
	"sync"
)

type TcpChatServer struct {
	listener net.Listener
	clients  []*client //该属性用于追踪连接的用户
	mutex    *sync.Mutex
}

type client struct {
	conn   net.Conn
	name   string
	writer *protocol.CommandWriter
}

// 启动
func NewServer() *TcpChatServer {
	return &TcpChatServer{
		mutex: &sync.Mutex{},
	}
}

// 监听服务
func (s *TcpChatServer) Listen(address string) error {
	l, err := net.Listen("tcp", address)
	if err == nil {
		s.listener = l
	}
	log.Printf("Listening on %v", address)
	return err
}

// 关闭服务
func (s *TcpChatServer) Close() {
	s.listener.Close()
}

// 启动服务
func (s *TcpChatServer) Start() {
	for {
		// need a way to break the loop
		conn, err := s.listener.Accept()
		if err != nil {
			log.Print(err)
		} else {
			// handle connection
			client := s.accept(conn)
			go s.serve(client)
		}
	}
}

// 将用户加入监听组
func (s *TcpChatServer) accept(conn net.Conn) *client {
	log.Printf("Accepting connection from %v, total clients: %v", conn.RemoteAddr().String(), len(s.clients)+1)
	s.mutex.Lock()
	defer s.mutex.Unlock()
	client := &client{
		conn:   conn,
		writer: protocol.NewCommandWriter(conn),
	}
	s.clients = append(s.clients, client)
	return client
}

// 将用户移除监听组
func (s *TcpChatServer) remove(client *client) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	// remove the connections from clients array
	for i, check := range s.clients {
		if check == client {
			s.clients = append(s.clients[:i], s.clients[i+1:]...)
		}
	}
	log.Printf("Closing connection from %v", client.conn.RemoteAddr().String())
	client.conn.Close()
}

// 接收从客户端传输过来的数据，并根据指令不同，用不同方式处理
func (s *TcpChatServer) serve(client *client) {
	cmdReader := protocol.NewCommandReader(client.conn)
	defer s.remove(client)
	for {
		cmd, err := cmdReader.Read()
		if err != nil && err != io.EOF {
			log.Printf("Read error: %v", err)
		}
		if cmd != nil {
			switch v := cmd.(type) {
			case protocol.SendCommand:
				go s.Broadcast(protocol.MessageCommand{
					Message: v.Message,
					Name:    client.name,
				})
			case protocol.NameCommand:
				client.name = v.Name
			}
		}
		if err == io.EOF {
			break
		}
	}
}

func (s *TcpChatServer) Broadcast(command interface{}) error {
	for _, client := range s.clients {
		// TODO: handle error here?
		client.writer.Write(command)
	}
	return nil
}
