package command

import (
	"encoding/json"
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/smf8/shalgham/common"
)

type TextMessage struct {
	Author         string `json:"author"`
	ConversationID int    `json:"conversation_id"`
	Text           string `json:"text"`
}

func CreateTextMessageCommand(author, text string, cid int) *TextMessage {
	return &TextMessage{
		Author:         author,
		ConversationID: cid,
		Text:           text,
	}
}

func CreateTextMessageFromMsg(msg common.Msg) (*TextMessage, error) {
	txtMsg := &TextMessage{}
	if err := json.Unmarshal(msg.Data, txtMsg); err != nil {
		return nil, fmt.Errorf("failed to parse data to textMessage cmd: %s", err)
	}

	return txtMsg, nil
}

func (t *TextMessage) GetMessage() common.Msg {
	data, err := json.Marshal(t)
	if err != nil {
		logrus.Errorf("failed to create send text message: %s", err)

		data = common.ErrorMessageData(err)
	}

	return common.Msg{
		NumberOfParts:  1,
		SequenceNumber: 1,
		Type:           TypeSendText,
		Data:           data,
	}
}
