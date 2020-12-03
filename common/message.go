package common

import (
	"encoding/json"
)

type Msg struct {
	Sender         string          `json:"sender"`
	Digest         string          `json:"digest"`
	Type           string          `json:"type"`
	SequenceNumber int             `json:"sequence_number"`
	NumberOfParts  int             `json:"number_of_parts"`
	Data           json.RawMessage `json:"data"`
}

func SuccessDataMessage() json.RawMessage {
	m := map[string]string{
		"data": "success",
	}
	j, _ := json.Marshal(m)

	return j
}

func (m *Msg) ToJSON() []byte {
	bytes, _ := json.Marshal(m)
	return bytes
}
