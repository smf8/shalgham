package log

import (
	"io/ioutil"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/smf8/shalgham/config"
)

func SetupLogger(cfg config.Logger) {
	logLevel, err := logrus.ParseLevel(cfg.Level)
	if err != nil {
		logLevel = logrus.ErrorLevel
	}

	if !cfg.Enabled {
		logrus.SetOutput(ioutil.Discard)
	}

	logrus.SetLevel(logLevel)

	if logLevel == logrus.DebugLevel {
		logrus.SetReportCaller(true)
		logrus.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: time.RFC3339,
		})
	} else {
		logrus.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: time.RFC3339,
		})
	}
}
