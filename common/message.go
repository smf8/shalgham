package common

import "encoding/json"

type Msg struct {
	Sender         string `json:"sender"`
	Digest         string `json:"digest"`
	SequenceNumber int    `json:"sequence_number"`
	NumberOfParts  int    `json:"number_of_parts"`
	Data           []byte `json:"data"`
}

func (m *Msg) ToJson() []byte {
	bytes, _ := json.Marshal(m)
	return bytes
}
