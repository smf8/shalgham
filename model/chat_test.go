package model_test

import (
	"testing"

	"github.com/smf8/shalgham/config"
	"github.com/smf8/shalgham/model"
	"github.com/smf8/shalgham/postgres"
	"github.com/stretchr/testify/suite"
)

type CityResolutionRepoSuite struct {
	suite.Suite
	repo     model.SQLChatRepo
	userRepo model.SQLUserRepo
}

func (suite *CityResolutionRepoSuite) SetupSuite() {
	cfg := config.New()

	postgresDB, err := postgres.Create(cfg.Postgres)
	suite.NoError(err)

	suite.repo = model.SQLChatRepo{DB: postgresDB}
	suite.userRepo = model.SQLUserRepo{DB: postgresDB}
}

//func (suite *CityResolutionRepoSuite) SetupTest() {
//  err := suite.repo.DB.Exec(`truncate table mess`).Error
//  suite.NoError(err)
//}
//
//func (suite *CityResolutionRepoSuite) TearDownTest() {
//  err := suite.repo.DB.Exec(`truncate table city_resolutions`).Error
//  suite.NoError(err)
//}

//nolint:funlen
func (suite *CityResolutionRepoSuite) TestChat() {
	user1 := model.User{
		Username: "user1",
		Password: "user1",
	}
	user2 := model.User{
		Username: "user2",
		Password: "user2",
	}

	suite.NoError(suite.userRepo.Save(user1))
	suite.NoError(suite.userRepo.Save(user2))

	conversation := model.Conversations{Name: "test"}
	suite.NoError(suite.repo.SaveConversation(conversation))

	c, err := suite.repo.FindConversation("test")
	suite.NoError(err)

	u1, err := suite.userRepo.FindUser("user1")
	suite.NoError(err)

	u2, err := suite.userRepo.FindUser("user2")
	suite.NoError(err)

	msg1 := model.Message{
		FromID:         u1.ID,
		ConversationID: c.ID,
		Body:           "Hello there buddy",
	}
	msg2 := model.Message{
		FromID:         u2.ID,
		ConversationID: c.ID,
		Body:           "hello back to you",
	}

	suite.NoError(suite.repo.SaveMessage(msg1))
	suite.NoError(suite.repo.SaveMessage(msg2))

	msgs, err := suite.repo.FindMessages(c.ID)
	suite.NoError(err)

	suite.Len(msgs, 2)
}

func TestCityResolutionRepoSuite(t *testing.T) {
	suite.Run(t, new(CityResolutionRepoSuite))
}
