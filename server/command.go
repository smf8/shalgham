package server

import (
	"errors"
	"fmt"
	"sort"
	"strings"

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

	server.Clients[client] = &user
	client.SendQueue <- response

	if userStatus := server.getUserStatusMsg(); userStatus != nil {
		client.SendQueue <- *userStatus
	}

	if convStatus := server.getUserStatusMsg(); convStatus != nil {
		client.SendQueue <- *convStatus
	}

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

	server.Clients[client] = user

	response := cmd.GetMessage()
	response.Sender = client.Conn.LocalAddr().String()
	response.Data = common.SuccessDataMessage()

	client.SendQueue <- response
	if userStatus := server.getUserStatusMsg(); userStatus != nil {
		client.SendQueue <- *userStatus
	}

	if convStatus := server.getUserStatusMsg(); convStatus != nil {
		client.SendQueue <- *convStatus
	}

	if err := server.UserRepo.Connect(cmd.Username); err != nil {
		return fmt.Errorf("login failed, could not update database")
	}

	return nil
}

func HandleTextMessage(cmd *command.TextMessage, msg common.Msg,
	server *Server, client *Client) error {
	if !msg.ValidateCheckSum() {
		return fmt.Errorf("sent and received checksums are not equal")
	}

	go func() {
		msgModel := model.Message{
			ConversationID: cmd.ConversationID,
			Body:           cmd.Text,
			FromID:         server.Clients[client].ID,
		}

		if err := server.ChatRepo.SaveMessage(msgModel); err != nil {
			logrus.Errorf("failed to save message in database: %s", err)
		}
	}()

	participants, err := server.ChatRepo.FindParticipants(cmd.ConversationID)
	if err != nil {
		return fmt.Errorf("no participants found: %w", err)
	}

	// not efficient
	for c, user := range server.Clients {
		for _, p := range participants {
			if user.ID == p.UserID {
				select {
				case c.SendQueue <- msg:
				}
			}
		}
	}

	return nil
}

func HandleJoinConv(cmd *command.JoinConversation, userIDs []int, server *Server, client *Client) error {
	conversation := model.Conversations{Name: cmd.ConversationName}
	conversation.Participants = make([]model.Participants, len(userIDs))

	if !cmd.IsGroup {
		usernames := []string{cmd.Participants[0], cmd.Participants[1]}
		sort.Strings(usernames)
		conversation.Name = strings.Join(usernames, "#")
	}

	for i, uid := range userIDs {
		participant := model.Participants{
			UserID: uid,
		}
		conversation.Participants[i] = participant
	}

	c, err := server.ChatRepo.FindConversation(conversation.Name)
	if err != nil {
		if errors.Is(err, model.ErrConversationNotFound) {
			c = &conversation

			if err := server.ChatRepo.SaveConversation(conversation); err != nil {
				return fmt.Errorf("failed to create conversation: %w", err)
			}
		}

		return fmt.Errorf("failed to get conversation with given name: %w", err)
	} else if cmd.IsGroup {
		conversation.Participants[0].ConversationID = int(c.ID)
		c.Participants = append(c.Participants, conversation.Participants[0])

		if err := server.ChatRepo.SaveParticipant(conversation.Participants[0]); err != nil {
			return fmt.Errorf("failed to add participant to conversation: %w", err)
		}
	}

	response := server.getConvStatusMsg(cmd.Participants[0])
	client.SendQueue <- *response

	return nil
}
