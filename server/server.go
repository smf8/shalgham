package server

import (
	"bufio"
	"net"

	"github.com/sirupsen/logrus"
	"github.com/smf8/shalgham/command"
	"github.com/smf8/shalgham/common"
	"github.com/smf8/shalgham/config"
	"github.com/smf8/shalgham/model"
)

const ClientBufferSize = 1024
const ServerBufferSize = 2048

type Server struct {
	connect      chan *Client
	disconnect   chan *Client
	Clients      map[*Client]*model.User
	routingTable map[string]*Client
	send         chan common.Msg
	ChatRepo     model.ChatRepo
	UserRepo     model.UserRepo
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
			SendQueue:      make(chan common.Msg, ClientBufferSize),
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
			logrus.Debugf("a client is trying to connect with Addr %s\n", c.Conn.RemoteAddr().String())
			s.routingTable[c.Conn.RemoteAddr().String()] = c
			s.Clients[c] = nil
		case c := <-s.disconnect:
			if user, ok := s.Clients[c]; ok {
				logrus.Infof("User %s with address %s disconnected", user.Username, c.Conn.RemoteAddr().String())
				delete(s.routingTable, c.Conn.RemoteAddr().String())
				close(c.SendQueue)
				delete(s.Clients, c)

				if user != nil {
					if err := s.UserRepo.Disconnect(user.Username); err != nil {
						logrus.Errorf("could not disconnect user in database: %s", err)
					}
				}
			}
		case msg := <-s.send:
			go s.handleMsg(msg)
		}
	}
}

func (s *Server) DisconnectUser(client *Client) {
	s.disconnect <- client
}

func (s *Server) getUserStatusMsg() *common.Msg {
	userStatusMsg := &common.Msg{}

	onlineUsers, err := s.UserRepo.FindOnline()
	if err != nil {
		logrus.Errorf("failed to get online users: %s", err)

		userStatusMsg = nil
	} else {
		userStatusCmd := command.UserStatus{Users: onlineUsers}
		m := userStatusCmd.GetMessage()
		userStatusMsg = &m
		userStatusMsg.CalculateChecksum()
	}

	return userStatusMsg
}

func (s *Server) getConvStatusMsg(username string) *common.Msg {
	convStatusMsg := &common.Msg{}
	//sorry for here :D
	conversations, err := s.ChatRepo.FindConversations(s.Clients[s.routingTable[username]].ID)
	if err != nil {
		logrus.Errorf("failed to get conversations list: %s", err)
		convStatusMsg = nil
	} else {
		convStatusCmd := command.ConversationStatus{Conversations: conversations}
		m := convStatusCmd.GetMessage()
		convStatusMsg = &m
		convStatusMsg.CalculateChecksum()
	}

	return convStatusMsg
}
func (s *Server) handleMsg(msg common.Msg) {
	logrus.Debugf("received msg %v\n", msg)

	if msg.Type == "login" {
		login, err := command.CreateLoginFromMsg(msg)
		if err != nil {
			logrus.Errorf("failed to handle login command: %s", err)
			return
		}

		err = HandleLogin(login, s, s.routingTable[msg.Sender])
		if err != nil {
			logrus.Errorf("failed to handle login command: %s", err)
			s.DisconnectUser(s.routingTable[msg.Sender])

			return
		}

		logrus.Infof("logged %s in", login.Username)
	} else if msg.Type == "signup" {
		signup, err := command.CreateSignupFromMsg(msg)
		if err != nil {
			logrus.Errorf("failed to handle signup command: %s", err)
			return
		}

		err = HandleSignup(signup, s, s.routingTable[msg.Sender])
		if err != nil {
			logrus.Errorf("failed to handle signup command: %s", err)
			s.DisconnectUser(s.routingTable[msg.Sender])

			return
		}

		logrus.Infof("signed %s up", signup.Username)
	}
}

func StartServer(chatRepo model.ChatRepo, userRepo model.UserRepo) *Server {
	server := &Server{
		connect:      make(chan *Client),
		disconnect:   make(chan *Client),
		send:         make(chan common.Msg, ServerBufferSize),
		Clients:      make(map[*Client]*model.User),
		routingTable: make(map[string]*Client),
		ChatRepo:     chatRepo,
		UserRepo:     userRepo,
	}

	return server
}
