package command

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/smf8/shalgham/common"
)

type FileMessage struct {
	Author         string `json:"author"`
	ConversationID int    `json:"conversation_id"`
	FileDigest     string `json:"file_digest"`
	FileName       string `json:"file_name"`
	File           []byte `json:"file"`
}

func CreateFileMessageCommand(author, filename string, cid int, file []byte) *FileMessage {
	f := &FileMessage{
		Author:         author,
		ConversationID: cid,
		File:           file,
		FileName:       filename,
	}
	digest := sha256.Sum256(file)
	f.FileDigest = fmt.Sprintf("%x", digest)

	return f
}

func CreateFileMessageFromMsgs(msg []common.Msg) (*FileMessage, error) {
	fileMsg := &FileMessage{}

	if err := json.Unmarshal(msg[0].Data, fileMsg); err != nil {
		return nil, fmt.Errorf("failed to parse data part to file message cmd: %s", err)
	}

	currentPart := 1

	bBuffer := new(bytes.Buffer)

	for i := range msg {
		for j := currentPart; j <= msg[i].NumberOfParts; j++ {
			if msg[i].SequenceNumber == currentPart {
				fileMsg := &FileMessage{}
				if err := json.Unmarshal(msg[i].Data, fileMsg); err != nil {
					return nil, fmt.Errorf("failed to parse data part %d to file message cmd: %s", currentPart, err)
				}

				bBuffer.Write(fileMsg.File)
				currentPart++
			}
		}
	}

	d := sha256.Sum256(bBuffer.Bytes())

	if fileMsg.FileDigest != fmt.Sprintf("%x", d) {
		return nil, fmt.Errorf("failed to create file message: file checksums are not equal")
	}

	fileMsg.File = bBuffer.Bytes()

	return fileMsg, nil
}

func (t *FileMessage) GetMessages() []common.Msg {

	filesize := len(t.File)
	numberOfChunks := (filesize / FileChunkSize) + 1

	chunks := make([][]byte, numberOfChunks)

	for byteIndex := 0; byteIndex < numberOfChunks; byteIndex++ {
		//chunks[byteIndex] = make([]byte, FileChunkSize)
		if (byteIndex+1)*FileChunkSize > len(t.File) {
			chunks[byteIndex] = t.File[(byteIndex)*FileChunkSize : len(t.File)-1]
		} else {
			chunks[byteIndex] = t.File[byteIndex*FileChunkSize : (byteIndex+1)*FileChunkSize]
		}
	}

	msgs := make([]common.Msg, numberOfChunks)
	for i := range chunks {
		msgs[i].NumberOfParts = numberOfChunks
		msgs[i].SequenceNumber = i + 1
		msgs[i].Type = TypeFileMessage
		t.File = chunks[i]

		data, err := json.Marshal(t)
		if err != nil {
			logrus.Errorf("failed to create file message: %s", err)

			return nil
		}

		msgs[i].Data = data
	}

	return msgs
}

func GetInfoFromMsg(msg common.Msg) (*FileMessage, error) {
	fileMsg := &FileMessage{}
	if err := json.Unmarshal(msg.Data, fileMsg); err != nil {
		return nil, fmt.Errorf("failed to parse data to file message cmd: %s", err)
	}

	return fileMsg, nil
}
