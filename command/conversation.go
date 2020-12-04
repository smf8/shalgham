package command

import (
	"encoding/json"
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/smf8/shalgham/common"
)

type JoinConversation struct {
	ConversationName string   `json:"conversation_name"`
	IsGroup          bool     `json:"is_group"`
	Participants     []string `json:"participants"`
}

func CreateJoinConvCmd(cName string, isGroup bool, participants []string) *JoinConversation {
	return &JoinConversation{
		ConversationName: cName,
		Participants:     participants,
		IsGroup:          isGroup,
	}
}

func CreateJoinConvFromMsg(msg common.Msg) (*JoinConversation, error) {
	convCmd := &JoinConversation{}
	if err := json.Unmarshal(msg.Data, convCmd); err != nil {
		return nil, fmt.Errorf("failed to parse data to join conversation cmd: %s", err)
	}

	return convCmd, nil
}

func (j *JoinConversation) GetMessage() common.Msg {
	data, err := json.Marshal(j)
	if err != nil {
		logrus.Errorf("failed to create join conversation message: %s", err)

		data = common.ErrorMessageData(err)
	}

	return common.Msg{
		NumberOfParts:  1,
		SequenceNumber: 1,
		Type:           TypeJoinConversation,
		Data:           data,
	}
}
