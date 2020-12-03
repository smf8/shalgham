package model

import (
	"errors"
	"time"

	"github.com/jinzhu/gorm"
)

var (
	ErrMessageNotFound      = errors.New("message not found")
	ErrParticipantNotFound  = errors.New("participant not found")
	ErrConversationNotFound = errors.New("conversation not found")
)

type (
	Participants struct {
		ID             int `json:"id"`
		ConversationID int `json:"conversation_id"`
		UserID         int `json:"user_id"`
	}

	Conversations struct {
		ID           int            `json:"id"`
		Name         string         `json:"name"`
		Participants []Participants `json:"participants"`
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
	FindConversations(userID int) ([]*Conversations, error)
	FindMessages(cid int) ([]Message, error)
	FindParticipants(cid int) ([]Participants, error)
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
			return nil, ErrMessageNotFound
		}

		return nil, err
	}

	return msgs, nil
}

func (s SQLChatRepo) FindConversation(name string) (*Conversations, error) {
	conversation := &Conversations{}

	if err := s.DB.Where("name = ?", name).Find(conversation).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, ErrConversationNotFound
		}

		return nil, err
	}

	return conversation, nil
}

func (s SQLChatRepo) FindParticipants(cid int) ([]Participants, error) {
	participants := make([]Participants, 0)

	if err := s.DB.Where("conversation_id = ?", cid).Find(&participants).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, ErrParticipantNotFound
		}

		return nil, err
	}

	return participants, nil
}

// fuck the new version of gorm !
func (s SQLChatRepo) FindConversations(userID int) ([]*Conversations, error) {
	conversations := make([]*Conversations, 0)

	if err := s.DB.Raw(`SELECT * FROM conversations as c
INNER JOIN (SELECT conversation_id FROM participants WHERE user_id = ?)
as p ON p.conversation_id = c.id`, userID).Find(&conversations).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, ErrConversationNotFound
		}

		return nil, err
	}

	return conversations, nil
}
