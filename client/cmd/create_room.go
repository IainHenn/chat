package cmd

import (
	"fmt"

	"github.com/IainHenn/chat/client/internal/api"
	"github.com/IainHenn/chat/client/internal/config"
	"github.com/spf13/cobra"
)

var createRoomCmd = &cobra.Command{
	Use:   "create-room [name]",
	Short: "Create a new chat room",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		server, err := config.ServerURL(serverFlag)
		if err != nil {
			return err
		}

		client := api.New(server)
		if err := client.CreateRoom(args[0], cfg.Username); err != nil {
			return err
		}

		fmt.Printf("Created room %q\n", args[0])
		return nil
	},
}

func init() {
	rootCmd.AddCommand(createRoomCmd)
}
