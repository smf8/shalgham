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
	connect    chan *Client
	disconnect chan *Client
	Clients    map[*Client]*model.User
	//routing table from <address:port> to clients
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
		go notifyOnlines(s)

		s.connect <- c
	}
}

func (s *Server) HandleClients() {
	for {
		select {
		case c := <-s.connect:
			logrus.Debugf("a client is trying to connect with Addr %s\n", c.Conn.RemoteAddr().String())
			s.routingTable[c.Conn.RemoteAddr().String()] = c
			s.Clients[c] = &model.User{Username: "undefined"}
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

func (s *Server) getUserStatusMsg(users []*model.User) *common.Msg {
	userStatusMsg := &common.Msg{}

	onlineUsers := users

	if users == nil {
		logrus.Errorf("failed to get online users")

		userStatusMsg = nil
	} else if len(onlineUsers) != 0 {
		for user := range onlineUsers {
			onlineUsers[user].Password = "zart!"

		}

		sender := ""
		for client, _ := range s.Clients {
			sender = client.Conn.LocalAddr().String()
			break
		}

		userStatusCmd := command.UserStatus{Users: onlineUsers}
		m := userStatusCmd.GetMessage()
		userStatusMsg = &m
		userStatusMsg.Sender = sender
		userStatusMsg.CalculateChecksum()
	} else {

		return nil
	}

	return userStatusMsg
}

func (s *Server) getConvStatusMsg(username string) *common.Msg {
	convStatusMsg := &common.Msg{}
	//sorry for here :D
	client, user := s.findUser(username)
	if user == nil {
		logrus.Errorf("failed getting conversation status, username not found in server")

		return nil
	}

	conversations, err := s.ChatRepo.FindConversations(user.ID)
	if err != nil {
		logrus.Errorf("failed to get conversations list: %s", err)
		convStatusMsg = nil
	} else {
		convStatusCmd := command.ConversationStatus{Conversations: conversations}
		m := convStatusCmd.GetMessage()
		convStatusMsg = &m
		convStatusMsg.Sender = client.Conn.LocalAddr().String()
		convStatusMsg.CalculateChecksum()
	}

	return convStatusMsg
}
func (s *Server) handleMsg(msg common.Msg) {
	//logrus.Debugf("received msg %v\n", msg)
	if msg.Type == command.TypeLogin {
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
	} else if msg.Type == command.TypeSignup {
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
	} else if msg.Type == command.TypeJoinConversation {
		join, err := command.CreateJoinConvFromMsg(msg)
		if err != nil {
			logrus.Errorf("failed to handle join conversation command: %s", err)
			return
		}

		uids := make([]int, len(join.Participants))

		for i, username := range join.Participants {
			if user, err := s.UserRepo.FindUser(username); err != nil {
				logrus.Errorf("could not find user in server")

				return
			} else {
				uids[i] = user.ID
			}
		}

		if err := HandleJoinConv(join, uids, s, s.routingTable[msg.Sender]); err != nil {
			logrus.Errorf("failed joining conversation: %s", err)
		}

		logrus.Infof("user %s joined conversation %s successfully\n", join.Participants[0], join.ConversationName)
	} else if msg.Type == command.TypeSendText {
		send, err := command.CreateTextMessageFromMsg(msg)
		if err != nil {
			logrus.Errorf("failed to create message command: %s", err)
		}

		if err = HandleTextMessage(send, msg, s, s.routingTable[msg.Sender]); err != nil {
			logrus.Errorf("failed to send message to peers in server: %s", err)
		}

		logrus.Infof("sent text message successfully")
	} else if msg.Type == command.TypeChangeUsername {
		changeUsername, err := command.CreateChangeUsernameFromMsg(msg)
		if err != nil {
			logrus.Errorf("failed to change username: %s", err)
		}

		if err = HandleChangeUsername(changeUsername, s, s.routingTable[msg.Sender]); err != nil {
			logrus.Errorf("failed to change username: %s", err)
		}
	} else if msg.Type == command.TypeFileMessage {
		fileCmd, err := command.GetInfoFromMsg(msg)
		if err != nil {
			logrus.Errorf("failed to send file message: %s", err)
		}

		if err = HandleFileMessage(fileCmd.ConversationID, msg, s, s.routingTable[msg.Sender]); err != nil {
			logrus.Errorf("failed to send file message: %s", err)
		}
	}
}

func (s *Server) findUser(username string) (*Client, *model.User) {
	for client, user := range s.Clients {
		if user.Username == username {
			return client, user
		}
	}

	return nil, nil
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
