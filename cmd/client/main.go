package client

import (
	"fmt"

	"github.com/smf8/shalgham/client"
	"github.com/smf8/shalgham/cmd/client/ui"
	"github.com/smf8/shalgham/config"
	"github.com/spf13/cobra"
)

func main(cfg config.Config) {
	conn, client := client.Connect(cfg.Server.Address)
	//cfg.Logger.Enabled = false

	//log.SetupLogger(cfg.Logger)
	ui.ShowUI(client)
	defer conn.Close()

	//client.Login(nil, nil)
	//time.Sleep(1 * time.Second)
	//client.JoinConversation(nil, nil)
	//time.Sleep(40 * time.Second)
}

// Register client command.
func Register(root *cobra.Command, cfg config.Config) {
	root.AddCommand(
		&cobra.Command{
			Use:   "client",
			Short: "Shalgham TCP chat client with a mighty TUI",
			RunE: func(cmd *cobra.Command, args []string) error {
				if len(args) != 1 {
					return fmt.Errorf("usage: %s <address:port>", cmd.ValidArgs[0])
				}

				main(cfg)

				return nil
			},
		},
	)
}
