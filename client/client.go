package client

import (
	"bufio"
	"fmt"
	"net"

	"github.com/jroimartin/gocui"
	"github.com/sirupsen/logrus"
	"github.com/smf8/shalgham/common"
	"github.com/smf8/shalgham/server"
)

func Connect(address string) (*net.Conn, *server.Client) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		logrus.Fatalf("failed connecting to server: %s\n", err)
	}

	defer conn.Close()

	client := &server.Client{
		Conn:           conn,
		OutputBuffer:   bufio.NewWriter(conn),
		ReadBuffer:     bufio.NewReader(conn),
		DisconnectChan: make(chan *server.Client),
		SendQueue:      make(chan common.Msg, server.ClientBufferSize),
		RecvQueue:      make(chan common.Msg, server.ClientBufferSize),
	}

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
				fmt.Println("Got new message")
				fmt.Println(msg)
			}
		}
	}()

	return &conn, client
}

func Login(g *gocui.Gui, v *gocui.View) error {
	fmt.Println(v.Buffer())
	v.Clear()

	v.Title = "password"
	//g.SetKeybinding("password", )

	return nil
}
