package client

import (
	"fmt"

	"github.com/smf8/shalgham/client"
	"github.com/smf8/shalgham/cmd/client/ui"
	"github.com/spf13/cobra"
)

func main(addr string) {
	ui.ShowUI()

	client.Connect(addr)
}

// Register client command.
func Register(root *cobra.Command) {
	root.AddCommand(
		&cobra.Command{
			Use:   "client",
			Short: "Shalgham TCP chat client with a mighty TUI",
			RunE: func(cmd *cobra.Command, args []string) error {
				if len(args) != 1 {
					return fmt.Errorf("usage: %s <address:port>", cmd.ValidArgs[0])
				}

				main(args[0])

				return nil
			},
		},
	)
}
