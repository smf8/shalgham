package client

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"sort"
	"strings"
	"time"

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
	ui            *gocui.Gui
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
			case cl := <-client.DisconnectChan:
				close(cl.SendQueue)

				if err := client.Conn.Close(); err != nil {
					logrus.Errorf("failed to close connection to server: %s\n", err)
				}

				conn.Close()
				return
			case msg := <-client.RecvQueue:
				//fmt.Println("Got message", msg)
				//fmt.Println("Message Data is: ", string(msg.Data))

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

func (c *Client) SetUI(g *gocui.Gui) {
	c.ui = g
}

func (c Client) Login(g *gocui.Gui, v *gocui.View) error {
	auth := v.Buffer()

	splitedAuth := strings.Split(auth, ":")

	//logrus.Infof("\n\n%s - %s\n\n", splitedAuth[0], splitedAuth[1])
	if len(splitedAuth) != 2 {
		fmt.Fprintln(v, "invalid login format")
	}

	loginCmd := command.CreateLoginCommand(splitedAuth[0], splitedAuth[1])
	msg := loginCmd.GetMessage()
	msg.Sender = c.C.Conn.LocalAddr().String()
	msg.Digest = msg.CalculateChecksum()
	c.C.SendQueue <- msg

	g.SetViewOnTop("messages")
	g.SetViewOnTop("users")
	g.SetViewOnTop("input")
	g.SetViewOnTop("conversations")
	g.SetCurrentView("input")

	return nil
}

func (c *Client) Signup(g *gocui.Gui, v *gocui.View) error {
	auth := v.Buffer()

	splitedAuth := strings.Split(auth, ":")

	if len(splitedAuth) != 2 {
		fmt.Fprintln(v, "invalid login format")
	}

	signUpCmd := command.CreateSignupCommand(splitedAuth[0], splitedAuth[1])
	msg := signUpCmd.GetMessage()
	msg.Sender = c.C.Conn.LocalAddr().String()
	msg.Digest = msg.CalculateChecksum()

	c.C.SendQueue <- msg

	g.SetViewOnTop("messages")
	g.SetViewOnTop("users")
	g.SetViewOnTop("input")
	g.SetViewOnTop("conversations")
	g.SetCurrentView("input")

	return nil
}

func (c *Client) JoinConversation(g *gocui.Gui, v *gocui.View) error {
	joinConvCmd := command.CreateJoinConvCmd("", false, []string{"jigar", "goosfand"})
	msg := joinConvCmd.GetMessage()
	msg.Sender = c.C.Conn.LocalAddr().String()
	msg.Digest = msg.CalculateChecksum()

	c.C.SendQueue <- msg

	return nil
}

func (c *Client) Disconnect(g *gocui.Gui, v *gocui.View) error {
	c.C.DisconnectChan <- c.C
	return gocui.ErrQuit
}
func (c *Client) HandleLogin(msg common.Msg) {
	//fmt.Println(string(msg.Data))
}

func (c *Client) HandleSignUp(msg common.Msg) {
	//fmt.Println(string(msg.Data))
}

func (c *Client) updateStatus() ([]string, []string) {
	usernames := make([]string, len(c.OnlineUsers))
	conversations := make([]string, len(c.Conversations))

	i := 0
	for username, _ := range c.OnlineUsers {
		usernames[i] = username
		i++
	}

	i = 0
	for cName := range c.Conversations {
		conversations[i] = cName
		i++
	}

	sort.Strings(usernames)
	sort.Strings(conversations)

	return usernames, conversations
}
func (c *Client) LoadStatus() {
	conversationMenu, _ := c.ui.View("conversations")
	usersMenu, _ := c.ui.View("users")

	for {
		usernames, cNames := c.updateStatus()
		onlineCount := 0

		users := new(strings.Builder)
		conversations := new(strings.Builder)

		for _, username := range usernames {
			users.WriteString(username)
			users.WriteString("  ")

			if c.OnlineUsers[username].IsOnline {
				onlineCount++
				users.WriteString("ONLINE")
			} else {
				users.WriteString("OFFLINE")
			}

			users.WriteString("\n")
		}

		for _, cname := range cNames {
			conversations.WriteString(cname)
			conversations.WriteString("  ")

			if !strings.Contains(c.Conversations[cname].Name, "#") {
				conversations.WriteString("G ")
				conversations.WriteString(fmt.Sprintf("%d", len(c.Conversations[cname].Participants)))
			}
		}

		c.ui.Update(func(g *gocui.Gui) error {
			usersMenu.Title = fmt.Sprintf(" %d online users: ", onlineCount)
			conversationMenu.Title = "Conversations"

			usersMenu.Clear()
			conversationMenu.Clear()
			fmt.Fprintln(usersMenu, users.String())
			fmt.Fprintln(conversationMenu, conversations.String())

			return nil
		})
		time.After(250 * time.Millisecond)
	}
}
func (c *Client) HandleConvStatus(msg common.Msg) {
	conversations := make([]*model.Conversations, 0)
	if err := json.Unmarshal(msg.Data, &conversations); err != nil {
		logrus.Errorf("failed to parse conversation status: %s", err)
	}

	for _, conv := range conversations {
		c.Conversations[conv.Name] = conv
		//fmt.Println(conv)
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
			//fmt.Println(user)
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
