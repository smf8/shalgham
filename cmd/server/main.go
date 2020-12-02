package server

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"
	"github.com/smf8/shalgham/config"
	"github.com/smf8/shalgham/server"
	"github.com/spf13/cobra"
)

//nolint:funlen
func main(cfg config.Config) {
	sv := server.StartServer()
	go sv.Listen(cfg.Server)
	go sv.HandleClients()

	logrus.Infof("started server in %s\n", cfg.Server.Address)
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	s := <-sig

	logrus.Infof("signal %s received", s)
}

func Register(root *cobra.Command, cfg config.Config) {
	root.AddCommand(
		&cobra.Command{
			Use:   "server",
			Short: "Shalgham Server instance",
			Run: func(cmd *cobra.Command, args []string) {
				main(cfg)
			},
		},
	)
}
