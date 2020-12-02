package postgres

import (
	"fmt"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/sirupsen/logrus"
	"github.com/smf8/shalgham/config"

	_ "github.com/jinzhu/gorm/dialects/postgres" // postgres driver
)

const (
	healthCheckInterval = 1
	maxAttempts         = 60
)

func Create(postgres config.Postgres) (*gorm.DB, error) {
	url := fmt.Sprintf(
		"host=%s port=%d user=%s dbname=%s password=%s connect_timeout=%d sslmode=disable",
		postgres.Host, postgres.Port, postgres.Username, postgres.DBName, postgres.Password,
		10,
	)

	postgresDB, err := gorm.Open("postgres", url)
	if err != nil {
		return nil, err
	}

	return postgresDB, nil
}

func WithRetry(fn func(postgres config.Postgres) (*gorm.DB, error), postgres config.Postgres) *gorm.DB {
	for i := 0; i < maxAttempts; i++ {
		db, err := fn(postgres)
		if err == nil {
			return db
		}

		logrus.Errorf("postgres: cannot connect to postgres. waiting 1 second. error is: %s", err.Error())

		time.Sleep(healthCheckInterval * time.Second)
	}

	panic(fmt.Sprintf("postgres: could not connect to postgres after %d attempts", maxAttempts))
}
