package cmd

import (
	"fmt"

	"github.com/IainHenn/chat/client/internal/api"
	"github.com/IainHenn/chat/client/internal/config"
	"github.com/spf13/cobra"
)

var leaveRoomCmd = &cobra.Command{
	Use:   "leave-room [name]",
	Short: "Leave a room without an active chat session",
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

		if err := api.New(server).LeaveRoom(args[0], activeUsername(cfg.Username)); err != nil {
			return err
		}

		fmt.Printf("Left room %q\n", args[0])
		return nil
	},
}

func init() {
	rootCmd.AddCommand(leaveRoomCmd)
}
