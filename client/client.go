package client

import (
	"bufio"
	"fmt"
	"net"

	"github.com/jroimartin/gocui"
	"github.com/sirupsen/logrus"
	"github.com/smf8/shalgham/command"
	"github.com/smf8/shalgham/common"
	"github.com/smf8/shalgham/server"
)

type Client struct {
	C *server.Client
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

	c := &Client{client}

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

				if msg.Type == "login" {
					c.HandleLogin(msg)
				} else if msg.Type == "signup" {
					c.HandleSignUp(msg)
				}
			}
		}
	}()

	return conn, c
}

func (c Client) Login(g *gocui.Gui, v *gocui.View) error {
	loginCmd := command.CreateLoginCommand("test", "test")
	msg := loginCmd.GetMessage()
	msg.Sender = c.C.Conn.LocalAddr().String()
	msg.Digest = msg.CheckSum()
	c.C.SendQueue <- msg
	//fmt.Println(v.Buffer())
	//v.Clear()
	//
	//v.Title = "password"
	////g.SetKeybinding("password", )

	return nil
}

func (c Client) Signup(g *gocui.Gui, v *gocui.View) error {
	loginCmd := command.CreateSignupCommand("test", "test")
	msg := loginCmd.GetMessage()
	msg.Sender = c.C.Conn.LocalAddr().String()
	msg.Digest = msg.CheckSum()

	c.C.SendQueue <- msg
	//fmt.Println(v.Buffer())
	//v.Clear()
	//
	//v.Title = "password"
	////g.SetKeybinding("password", )

	return nil
}

func (c Client) HandleLogin(msg common.Msg) {
	fmt.Println(string(msg.Data))
}

func (c Client) HandleSignUp(msg common.Msg) {
	fmt.Println(string(msg.Data))
}
