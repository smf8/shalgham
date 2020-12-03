package command

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/smf8/shalgham/common"
)

type Signup struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func CreateSignupCommand(username, password string) *Signup {
	cmd := &Signup{
		Username: username,
	}

	hash := sha256.Sum256([]byte(password))
	cmd.Password = fmt.Sprintf("%x", hash)

	return cmd
}

func CreateSignupFromMsg(msg common.Msg) (*Signup, error) {
	signup := &Signup{}
	if err := json.Unmarshal(msg.Data, signup); err != nil {
		return nil, fmt.Errorf("failed to parse data to signup cmd: %s", err)
	}

	return signup, nil
}

func (s *Signup) GetMessage() common.Msg {
	data, err := json.Marshal(s)
	if err != nil {
		logrus.Errorf("failed to create signup message: %s", err)

		data = common.ErrorMessageData(err)
	}

	return common.Msg{
		NumberOfParts:  1,
		SequenceNumber: 1,
		Type:           TypeSignup,
		Data:           data,
	}
}
