package client

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"runtime"
	"sort"
	"strings"
	"sync"
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
	userLock      sync.RWMutex
	OnlineUsers   map[string]*model.User
	username      string
	convLock      sync.RWMutex
	Conversations map[string]*model.Conversations
	ui            *gocui.Gui
	currentConv   *model.Conversations
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
				} else if msg.Type == command.TypeConversationStatus {
					c.handleConversationStatusMsg(msg)
				} else if msg.Type == command.TypeSendText {
					c.handleTextMessage(msg)
				}
			}
		}
	}()

	return conn, c
}

func (c *Client) SetUI(g *gocui.Gui) {
	c.ui = g
}

func (c *Client) SubmitInput(g *gocui.Gui, v *gocui.View) error {
	input := v.Buffer()
	input = strings.TrimSpace(input)
	messages, _ := g.View("messages")

	if strings.HasPrefix(input, "/join") {
		cName := strings.TrimLeft(input, "/join ")

		v.Clear()
		v.SetCursor(0, 0)

		if _, ok := c.OnlineUsers[cName]; ok {
			return c.JoinConversation(g, v, cName, []string{c.username, cName})
		} else {
			return c.JoinConversation(g, v, cName, []string{c.username})
		}
	} else {
		if c.currentConv == nil {
			v.Clear()
			v.SetCursor(0, 0)

			g.Update(func(g *gocui.Gui) error {
				c.writeError("you must join a conversation first", messages)

				return nil
			})
		} else {
			c.SendTextMsg(input, g, v)
			v.Clear()

			return v.SetCursor(0, 0)
		}
	}

	return nil
}

func (c *Client) Login(g *gocui.Gui, v *gocui.View) error {
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
	c.username = splitedAuth[0]

	c.C.SendQueue <- msg

	g.SetViewOnTop("messages")
	g.SetViewOnTop("users")
	g.SetViewOnTop("input")
	g.SetViewOnTop("conversations")
	g.SetCurrentView("input")

	go c.LoadStatus(g)

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
	c.username = splitedAuth[0]

	c.C.SendQueue <- msg

	g.SetViewOnTop("messages")
	g.SetViewOnTop("users")
	g.SetViewOnTop("input")
	g.SetViewOnTop("conversations")
	g.SetCurrentView("input")

	go c.LoadStatus(g)

	return nil
}

func (c *Client) JoinConversation(g *gocui.Gui, v *gocui.View, cName string, participants []string) error {
	c.convLock.RLock()
	conv, ok := c.Conversations[cName]
	c.convLock.RUnlock()
	msgView, err := g.View("messages")

	if err != nil {
		logrus.Errorf("loading messages view: %s", err)
	}

	if ok {
		c.currentConv = conv
		g.Update(func(g *gocui.Gui) error {
			msgView.Title = conv.Name
			msgView.Clear()
			msgView.SetCursor(0, 0)

			return nil
		})
	} else {

	}

	joinConvCmd := command.CreateJoinConvCmd(cName, len(participants) != 2, participants)
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

func (c *Client) SendTextMsg(message string, g *gocui.Gui, v *gocui.View) error {
	textMsg := command.CreateTextMessageCommand(c.username, message, c.currentConv.ID)
	//messages, _ := g.View("messages")
	msg := textMsg.GetMessage()
	msg.Sender = c.C.Conn.LocalAddr().String()

	//msgModel := model.Message{
	//	Author:    c.username,
	//	Body:      "message",
	//	CreatedAT: time.Now(),
	//}

	g.Update(func(g *gocui.Gui) error {
		//c.writeMessage(msgModel, messages)

		return nil
	})

	c.C.SendQueue <- msg

	return nil
}

func (c *Client) HandleLogin(msg common.Msg) {
	fmt.Println(string(msg.Data), runtime.NumGoroutine())
}

func (c *Client) HandleSignUp(msg common.Msg) {
	//fmt.Println(string(msg.Data))
}

func (c *Client) updateStatus() ([]string, []string) {
	c.userLock.RLock()
	usernames := make([]string, len(c.OnlineUsers))
	c.userLock.RUnlock()
	c.convLock.RLock()
	conversations := make([]string, len(c.Conversations))
	c.convLock.RUnlock()

	i := 0
	c.userLock.RLock()
	for username, _ := range c.OnlineUsers {
		usernames[i] = username
		i++
	}
	c.userLock.RUnlock()

	c.convLock.RLock()
	i = 0
	for cName := range c.Conversations {
		conversations[i] = cName
		i++
	}

	c.convLock.RUnlock()
	sort.Strings(usernames)
	sort.Strings(conversations)

	return usernames, conversations
}
func (c *Client) LoadStatus(g *gocui.Gui) {

	conversationMenu, _ := g.View("conversations")
	usersMenu, _ := g.View("users")

	for {
		usernames, cNames := c.updateStatus()
		onlineCount := 0

		users := new(strings.Builder)
		conversations := new(strings.Builder)

		for _, username := range usernames {
			users.WriteString(username)
			users.WriteString("  ")

			c.userLock.RLock()
			if c.OnlineUsers[username].IsOnline {
				onlineCount++
				users.WriteString("\u001B[32;1mONLINE\u001B[0m")
			} else {
				users.WriteString("\u001B[31;1mOFFLINE\u001B[0m")
			}
			c.userLock.RUnlock()

			users.WriteString("\n")
		}

		for _, cname := range cNames {
			conversations.WriteString(cname)

			c.convLock.RLock()
			found := false
			for user := range c.OnlineUsers {
				if c.Conversations[cname].Name == user {
					found = true
				}
			}

			if !found {
				conversations.WriteString("-")
				conversations.WriteString("G")
				conversations.WriteString(fmt.Sprintf("%d", len(c.Conversations[cname].Participants)))
			}
			conversations.WriteString("\n")
			c.convLock.RUnlock()
		}

		g.Update(func(g *gocui.Gui) error {
			usersMenu.Title = fmt.Sprintf(" %d online users: ", onlineCount)
			conversationMenu.Title = "Conversations"

			usersMenu.Clear()
			conversationMenu.Clear()
			fmt.Fprintln(usersMenu, users.String())
			fmt.Fprintln(conversationMenu, conversations.String())

			return nil
		})
		<-time.After(250 * time.Millisecond)
	}
}

func (c *Client) HandleUserStatus(msg common.Msg) {
	users := command.UserStatus{
		Users: make([]*model.User, 0),
	}

	if err := json.Unmarshal(msg.Data, &users); err != nil {
		logrus.Errorf("failed to parse user status: %s", err)
	}

	c.userLock.RLock()
	if len(c.OnlineUsers) == 0 {
		c.userLock.RUnlock()
		for _, user := range users.Users {

			c.userLock.Lock()
			c.OnlineUsers[user.Username] = user
			c.userLock.Unlock()
			//fmt.Println(user)
		}
	} else {
		c.userLock.RLock()
		for username, user := range c.OnlineUsers {
			found := false

			for _, u := range users.Users {
				if u.Username == username {
					user.IsOnline = true
					found = true
				}
			}

			if !found {
				user.IsOnline = false
			}
		}
		c.userLock.RUnlock()
	}
}

func (c *Client) handleConversationStatusMsg(msg common.Msg) {
	convs := command.ConversationStatus{}
	if err := json.Unmarshal(msg.Data, &convs); err != nil {
		logrus.Errorf("failed to parse conversation status: %s", err)
	}

	//if len(c.Conversations) == 0 {
	for _, conv := range convs.Conversations {
		if strings.Contains(conv.Name, "#") {
			s := strings.Split(conv.Name, "#")
			if s[0] == c.username {
				conv.Name = s[1]
			} else {
				conv.Name = s[0]
			}
		}
		c.Conversations[conv.Name] = conv
	}
	//}

	msgView, _ := c.ui.View("messages")

	if len(convs.Conversations) == 1 && (c.currentConv != nil && strings.Contains(convs.Conversations[0].Name, c.currentConv.Name)) {

		c.ui.Update(func(g *gocui.Gui) error {
			for _, msg := range convs.Conversations[0].Messages {
				c.writeMessage(msg, msgView)
			}

			return nil
		})

	}
}

func (c *Client) handleTextMessage(msg common.Msg) {
	messageView, _ := c.ui.View("messages")
	txtMsg, err := command.CreateTextMessageFromMsg(msg)
	message := model.Message{
		CreatedAT: time.Now(),
	}

	if err != nil || txtMsg == nil {
		message.Body = "\u001B[31;1mERROR\u001B[0m"
		message.Author = "NO ONE"
		c.ui.Update(func(g *gocui.Gui) error {
			c.writeMessage(message, messageView)

			return nil
		})

		return
	}

	message.Author = txtMsg.Author
	message.ConversationID = txtMsg.ConversationID
	message.Body = txtMsg.Text

	if c.currentConv.ID == txtMsg.ConversationID {
		c.ui.Update(func(g *gocui.Gui) error {
			c.writeMessage(message, messageView)

			return nil
		})
	}
}

func (c *Client) writeMessage(msg model.Message, v *gocui.View) {
	fmt.Fprintf(v, "\u001B[3%d;%dm[%s]\u001B[0m  \u001B[3%d;%dm%s\u001B[0m: %s\n", 3, 1, msg.CreatedAT.Format("2006-01-02 15:04:05"), 2, 7, msg.Author, msg.Body)
}

func (c *Client) writeError(msg string, v *gocui.View) {
	fmt.Fprintf(v, "\u001B[3%d;%dm[%s]\u001B[0m\n", 1, 1, msg)
}
