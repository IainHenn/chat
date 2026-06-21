package cmd

import (
	"github.com/IainHenn/chat/client/internal/api"
	"github.com/IainHenn/chat/client/internal/chat"
	"github.com/IainHenn/chat/client/internal/config"
	"github.com/spf13/cobra"
)

var joinRoomCmd = &cobra.Command{
	Use:   "join-room [name]",
	Short: "Join a room and start chatting over WebSocket",
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

		return chat.RunRoom(server, args[0], cfg.Username, api.New(server))
	},
}

func init() {
	rootCmd.AddCommand(joinRoomCmd)
}
