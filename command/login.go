package command

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/smf8/shalgham/common"
)

type Login struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func CreateLoginCommand(username, password string) *Login {
	cmd := &Login{
		Username: username,
	}

	hash := sha256.Sum256([]byte(password))
	cmd.Password = string(hash[:])

	return cmd
}

func CreateLoginFromMsg(msg common.Msg) (*Login, error) {
	login := &Login{}
	if err := json.Unmarshal(msg.Data, login); err != nil {
		return nil, fmt.Errorf("failed to parse data to login cmd: %s", err)
	}

	return login, nil
}

func (l *Login) GetMessage() common.Msg {
	data, err := json.Marshal(l)
	if err != nil {
		logrus.Errorf("failed to create login message: %s", err)

		data = []byte("error")
	}

	return common.Msg{
		Digest:         "",
		NumberOfParts:  1,
		SequenceNumber: 1,
		Type:           "login",
		Data:           data,
	}
}
