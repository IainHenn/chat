package cmd

import (
	"fmt"

	"github.com/IainHenn/chat/client/internal/api"
	"github.com/IainHenn/chat/client/internal/config"
	"github.com/spf13/cobra"
)

var deleteRoomCmd = &cobra.Command{
	Use:   "delete-room [name]",
	Short: "Delete a room you host",
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

		if err := api.New(server).DeleteRoom(args[0], activeUsername(cfg.Username)); err != nil {
			return err
		}

		fmt.Printf("Deleted room %q\n", args[0])
		return nil
	},
}

func init() {
	rootCmd.AddCommand(deleteRoomCmd)
}
