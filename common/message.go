package common

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
)

type Msg struct {
	Sender         string          `json:"sender"`
	Digest         string          `json:"digest"`
	Type           string          `json:"type"`
	SequenceNumber int             `json:"sequence_number"`
	NumberOfParts  int             `json:"number_of_parts"`
	Data           json.RawMessage `json:"data"`
}

type TextMessage struct {
}

func SuccessDataMessage() json.RawMessage {
	m := map[string]string{
		"data": "success",
	}
	j, _ := json.Marshal(m)

	return j
}

func ErrorMessageData(err error) json.RawMessage {
	m := map[string]string{
		"data": err.Error(),
	}
	j, _ := json.Marshal(m)

	return j
}

func (m *Msg) ValidateCheckSum() bool {
	digest := m.Digest
	return digest == m.CalculateChecksum()
}

func (m *Msg) CalculateChecksum() string {
	checksum := sha256.Sum256([]byte(fmt.Sprintf("%v", m)))
	result := fmt.Sprintf("%x", checksum)
	m.Digest = result

	return result
}

func (m *Msg) ToJSON() []byte {
	bytes, _ := json.Marshal(m)
	return bytes
}
