package command

import (
	"encoding/json"
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/smf8/shalgham/common"
	"github.com/smf8/shalgham/model"
)

type UserStatus struct {
	Users []*model.User `json:"users"`
}
type ConversationStatus struct {
	Conversations []*model.Conversations `json:"conversations"`
}

func CreateConversationStatusFromMsg(msg common.Msg) (*ConversationStatus, error) {
	status := &ConversationStatus{}
	if err := json.Unmarshal(msg.Data, status); err != nil {
		return nil, fmt.Errorf("failed to parse data to conversation status cmd: %s", err)
	}

	return status, nil
}

func (c *ConversationStatus) GetMessage() common.Msg {
	data, err := json.Marshal(c)
	if err != nil {
		logrus.Errorf("failed to create conversation status message: %s", err)

		data = common.ErrorMessageData(err)
	}

	return common.Msg{
		NumberOfParts:  1,
		SequenceNumber: 1,
		Type:           TypeConversationStatus,
		Data:           data,
	}
}

func CreateUserStatusFromMsg(msg common.Msg) (*UserStatus, error) {
	status := &UserStatus{}
	if err := json.Unmarshal(msg.Data, status); err != nil {
		return nil, fmt.Errorf("failed to parse data to user status cmd: %s", err)
	}

	return status, nil
}

func (s *UserStatus) GetMessage() common.Msg {
	data, err := json.Marshal(s)
	if err != nil {
		logrus.Errorf("failed to create user status message: %s", err)

		data = common.ErrorMessageData(err)
	}

	return common.Msg{
		NumberOfParts:  1,
		SequenceNumber: 1,
		Type:           TypeUserStatus,
		Data:           data,
	}
}
