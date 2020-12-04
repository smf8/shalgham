package command

import (
	"encoding/json"
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/smf8/shalgham/common"
)

type ChangeUsername struct {
	OldUsername string `json:"old_username"`
	NewUsername string `json:"new_username"`
	Status      bool   `json:"status"`
}

func CreateChangeUsernameCmd(old, new string) *ChangeUsername {
	return &ChangeUsername{
		OldUsername: old,
		NewUsername: new,
	}
}

func CreateChangeUsernameFromMsg(msg common.Msg) (*ChangeUsername, error) {
	chUsername := &ChangeUsername{}
	if err := json.Unmarshal(msg.Data, chUsername); err != nil {
		return nil, fmt.Errorf("failed to parse data to change username cmd: %s", err)
	}

	return chUsername, nil
}

func (t *ChangeUsername) GetMessage() common.Msg {
	data, err := json.Marshal(t)
	if err != nil {
		logrus.Errorf("failed to create change username message: %s", err)

		data = common.ErrorMessageData(err)
	}

	return common.Msg{
		NumberOfParts:  1,
		SequenceNumber: 1,
		Type:           TypeChangeUsername,
		Data:           data,
	}
}
