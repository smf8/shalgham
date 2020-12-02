package client

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/smf8/shalgham/command"
	"github.com/smf8/shalgham/common"
	"github.com/smf8/shalgham/server"
	"github.com/spf13/cobra"
)

func main(addr string) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		logrus.Fatalf("failed connecting to server: %s\n", err)
	}

	defer conn.Close()

	client := &server.Client{
		Conn:           conn,
		OutputBuffer:   bufio.NewWriter(conn),
		ReadBuffer:     bufio.NewReader(conn),
		DisconnectChan: make(chan *server.Client),
		SendQueue:      make(chan common.Msg, 128),
		RecvQueue:      make(chan common.Msg, 128),
	}

	go client.ReadMessage()
	go client.SendMessage()

	loginCmd := command.Login{
		Sender:   client,
		Username: "Joon nanat kar kon",
		Password: "an",
	}

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

	time.Sleep(1 * time.Second)

	client.SendQueue <- loginCmd.GetMessage()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	s := <-sig

	client.DisconnectChan <- client
	logrus.Infof("signal %s received", s)
	time.Sleep(1 * time.Second)
}

// Register client command
func Register(root *cobra.Command) {
	root.AddCommand(
		&cobra.Command{
			Use:   "client",
			Short: "Shalgham TCP chat client with a mighty TUI",
			RunE: func(cmd *cobra.Command, args []string) error {
				if len(args) != 1 {
					return fmt.Errorf("usage: %s <address:port>\n", cmd.ValidArgs[0])
				}

				main(args[0])

				return nil
			},
		},
	)
}
