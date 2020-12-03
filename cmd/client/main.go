package client

import (
	"fmt"
	"time"

	"github.com/smf8/shalgham/client"
	"github.com/spf13/cobra"
)

func main(addr string) {
	//ui.ShowUI()
	conn, client := client.Connect(addr)

	defer conn.Close()

	client.Login(nil, nil)
	time.Sleep(40 * time.Second)
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
