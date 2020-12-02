package command

import (
	"encoding/json"

	"github.com/sirupsen/logrus"
	"github.com/smf8/shalgham/common"
	"github.com/smf8/shalgham/server"
)

type Login struct {
	Sender   *server.Client `json:"-"`
	Username string         `json:"username"`
	Password string         `json:"password"`
}

func (l *Login) GetMessage() common.Msg {
	data, err := json.Marshal(l)
	if err != nil {
		logrus.Errorf("failed to create login message: %s", err)
		data = []byte("error")
	}

	return common.Msg{
		Sender:         l.Sender.Conn.RemoteAddr().String(),
		Digest:         "",
		NumberOfParts:  1,
		SequenceNumber: 1,
		Data:           data,
	}
}
