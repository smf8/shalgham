package server

import (
	"bufio"
	"net"

	"github.com/sirupsen/logrus"
	"github.com/smf8/shalgham/common"
	"github.com/smf8/shalgham/config"
)

type Server struct {
	connect    chan *Client
	disconnect chan *Client
	clients    map[*Client]string
	send       chan common.Msg
}

func (s *Server) Listen(cfg config.Server) {
	server, err := net.Listen("tcp", cfg.Address)
	if err != nil {
		logrus.Fatalf("failed to start server: %s", err)
	}

	for {
		conn, err := server.Accept()
		if err != nil {
			logrus.Errorf("failed accepting client connection: %s", err)
		}

		c := &Client{
			Conn:           conn,
			OutputBuffer:   bufio.NewWriter(conn),
			ReadBuffer:     bufio.NewReader(conn),
			SendQueue:      make(chan common.Msg, 1024),
			RecvQueue:      s.send,
			DisconnectChan: s.disconnect,
		}

		go c.ReadMessage()
		go c.SendMessage()

		s.connect <- c
	}
}

func (s *Server) HandleClients() {
	for {
		select {
		case c := <-s.connect:
			// a client is trying to connect
			logrus.Debugf("a client is trying to connect with Addr %s\n", c.Conn.RemoteAddr().String())
			s.clients[c] = "undefined"
		case c := <-s.disconnect:
			if _, ok := s.clients[c]; ok {
				close(c.SendQueue)
				delete(s.clients, c)
			}
		case msg := <-s.send:
			s.handleMsg(msg)

		}
	}
}

func (s *Server) handleMsg(msg common.Msg) {
	logrus.Debugf("received msg %v\n", msg)
	//send to ALL
	for client := range s.clients {
		select {
		case client.SendQueue <- msg:
		default:
			break
		}
	}
}

func StartServer() *Server {
	server := &Server{
		connect:    make(chan *Client),
		disconnect: make(chan *Client),
		send:       make(chan common.Msg, 2048),
		clients:    make(map[*Client]string),
	}

	return server
}
