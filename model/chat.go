package model

import (
	"time"

	"github.com/jinzhu/gorm"
)

type (
	Participants struct {
		ID             int `json:"id"`
		ConversationID int `json:"conversation_id"`
		UserID         int `json:"user_id"`
	}

	Conversations struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}

	Message struct {
		ID             int       `json:"id"`
		ConversationID int       `json:"conversation_id"`
		FromID         int       `json:"from_id"`
		Body           string    `json:"body"`
		CreatedAT      time.Time `json:"created_at"`
	}
)

type ChatRepo interface {
	SaveParticipant(p Participants) error
	SaveMessage(msg Message) error
	SaveConversation(c Conversations) error
	FindConversation(name string) (*Conversations, error)
	FindMessages(cid int) ([]Message, error)
}

type SQLChatRepo struct {
	DB *gorm.DB
}

func (s SQLChatRepo) SaveParticipant(p Participants) error {
	return s.DB.Create(&p).Error
}

func (s SQLChatRepo) SaveMessage(msg Message) error {
	return s.DB.Create(&msg).Error
}

func (s SQLChatRepo) SaveConversation(c Conversations) error {
	return s.DB.Create(&c).Error
}

func (s SQLChatRepo) FindMessages(cid int) ([]Message, error) {
	msgs := make([]Message, 0)

	if err := s.DB.Where("conversation_id = ?", cid).Find(&msgs).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, ErrUserNotFound
		}

		return nil, err
	}

	return msgs, nil
}

func (s SQLChatRepo) FindConversation(name string) (*Conversations, error) {
	conversation := &Conversations{}

	if err := s.DB.Where("name = ?", name).Find(conversation).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, ErrUserNotFound
		}

		return nil, err
	}

	return conversation, nil
}
