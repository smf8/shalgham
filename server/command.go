package server

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/smf8/shalgham/command"
	"github.com/smf8/shalgham/common"
)

func HandleLogin(cmd *command.Login, server *Server, client *Client) error {
	user, err := server.UserRepo.FindUser(cmd.Username)
	if err != nil {
		logrus.Errorf("login cmd failed, user not found: %s", err)
		return fmt.Errorf("user not found")
	}

	if user.Password != cmd.Password {
		return fmt.Errorf("invalid password")
	}

	server.Clients[client] = cmd.Username

	response := cmd.GetMessage()

	response.Sender = client.Conn.LocalAddr().String()

	response.Data = common.SuccessDataMessage()

	client.SendQueue <- response

	if err := server.UserRepo.Connect(cmd.Username); err != nil {
		return fmt.Errorf("login failed, could not update database")
	}

	return nil
}
