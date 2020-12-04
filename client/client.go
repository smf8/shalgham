package client

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"

	"github.com/jroimartin/gocui"
	"github.com/sirupsen/logrus"
	"github.com/smf8/shalgham/command"
	"github.com/smf8/shalgham/common"
	"github.com/smf8/shalgham/model"
	"github.com/smf8/shalgham/server"
)

type Client struct {
	C             *server.Client
	OnlineUsers   map[string]*model.User
	Conversations map[string]*model.Conversations
}

func Connect(address string) (net.Conn, *Client) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		logrus.Fatalf("failed connecting to server: %s\n", err)
	}

	client := &server.Client{
		Conn:           conn,
		OutputBuffer:   bufio.NewWriter(conn),
		ReadBuffer:     bufio.NewReader(conn),
		DisconnectChan: make(chan *server.Client),
		SendQueue:      make(chan common.Msg, server.ClientBufferSize),
		RecvQueue:      make(chan common.Msg, server.ClientBufferSize),
	}

	c := &Client{C: client}

	c.OnlineUsers = make(map[string]*model.User)
	c.Conversations = make(map[string]*model.Conversations)

	go client.ReadMessage()
	go client.SendMessage()
	go func() {
		for {
			select {
			case c := <-client.DisconnectChan:
				close(c.SendQueue)

				if err := client.Conn.Close(); err != nil {
					logrus.Errorf("failed to close connection to server: %s\n", err)
				}

				conn.Close()

				return
			case msg := <-client.RecvQueue:
				fmt.Println("Got message", msg)
				fmt.Println("Message Data is: ", string(msg.Data))

				if msg.Type == command.TypeLogin {
					c.HandleLogin(msg)
				} else if msg.Type == command.TypeSignup {
					c.HandleSignUp(msg)
				} else if msg.Type == command.TypeUserStatus {
					c.HandleUserStatus(msg)
				}
			}
		}
	}()

	return conn, c
}

//func (c Client) SendText(g *gocui.Gui, v *gocui.View) error {
//	txtCmd :=
//}

func (c Client) Login(g *gocui.Gui, v *gocui.View) error {
	loginCmd := command.CreateLoginCommand("jigar", "joon")
	msg := loginCmd.GetMessage()
	msg.Sender = c.C.Conn.LocalAddr().String()
	msg.Digest = msg.CalculateChecksum()
	c.C.SendQueue <- msg
	//fmt.Println(v.Buffer())
	//v.Clear()
	//
	//v.Title = "password"
	////g.SetKeybinding("password", )

	return nil
}

func (c *Client) Signup(g *gocui.Gui, v *gocui.View) error {
	signUpCmd := command.CreateSignupCommand("goosfand", "joon")
	msg := signUpCmd.GetMessage()
	msg.Sender = c.C.Conn.LocalAddr().String()
	msg.Digest = msg.CalculateChecksum()

	c.C.SendQueue <- msg
	//fmt.Println(v.Buffer())
	//v.Clear()
	//
	//v.Title = "password"
	////g.SetKeybinding("password", )

	return nil
}

func (c *Client) JoinConversation(g *gocui.Gui, v *gocui.View) error {
	joinConvCmd := command.CreateJoinConvCmd("", false, []string{"jigar", "goosfand"})
	msg := joinConvCmd.GetMessage()
	msg.Sender = c.C.Conn.LocalAddr().String()
	msg.Digest = msg.CalculateChecksum()

	c.C.SendQueue <- msg
	//fmt.Println(v.Buffer())
	//v.Clear()
	//
	//v.Title = "password"
	////g.SetKeybinding("password", )

	return nil
}

func (c *Client) HandleLogin(msg common.Msg) {
	fmt.Println(string(msg.Data))
}

func (c *Client) HandleSignUp(msg common.Msg) {
	fmt.Println(string(msg.Data))
}

func (c *Client) HandleConvStatus(msg common.Msg) {
	conversations := make([]*model.Conversations, 0)
	if err := json.Unmarshal(msg.Data, &conversations); err != nil {
		logrus.Errorf("failed to parse conversation status: %s", err)
	}

	for _, conv := range conversations {
		c.Conversations[conv.Name] = conv
		fmt.Println(conv)
	}
}

func (c *Client) HandleUserStatus(msg common.Msg) {
	users := command.UserStatus{
		Users: make([]*model.User, 0),
	}

	if err := json.Unmarshal(msg.Data, &users); err != nil {
		logrus.Errorf("failed to parse user status: %s", err)
	}

	if len(c.OnlineUsers) == 0 {
		for _, user := range users.Users {
			c.OnlineUsers[user.Username] = user
			fmt.Println(user)
		}
	} else {
		for username, user := range c.OnlineUsers {
			found := false

			for _, u := range users.Users {
				if u.Username == username {
					found = true
				}
			}

			if !found {
				user.IsOnline = false
			}
		}
	}
}
