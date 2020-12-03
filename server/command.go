package server

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/smf8/shalgham/command"
	"github.com/smf8/shalgham/common"
	"github.com/smf8/shalgham/model"
)

func HandleSignup(cmd *command.Signup, server *Server, client *Client) error {
	user := model.User{
		Username: cmd.Username,
		Password: cmd.Password,
		IsOnline: true,
	}

	if err := server.UserRepo.Save(user); err != nil {
		return fmt.Errorf("signup failed: %w", err)
	}

	response := cmd.GetMessage()
	response.Sender = client.Conn.LocalAddr().String()
	response.Data = common.SuccessDataMessage()

	server.Clients[client] = cmd.Username
	client.SendQueue <- response

	return nil
}
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
