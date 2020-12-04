package cmd

import (
	"os"

	"github.com/sirupsen/logrus"
	"github.com/smf8/shalgham/cmd/client"
	"github.com/smf8/shalgham/cmd/server"
	"github.com/smf8/shalgham/config"
	"github.com/smf8/shalgham/log"

	"github.com/spf13/cobra"
)

const ExitFailure = 1

func Execute() {
	cfg := config.New()

	log.SetupLogger(cfg.Logger)

	var root = &cobra.Command{
		Use:   "Shalgham",
		Short: "Shalgham TCP chat",
	}

	server.Register(root, cfg)
	client.Register(root, cfg)
	//migrate.Register(root, cfg)

	if err := root.Execute(); err != nil {
		logrus.Errorf("failed to execute root command: %s", err.Error())
		os.Exit(ExitFailure)
	}
}
