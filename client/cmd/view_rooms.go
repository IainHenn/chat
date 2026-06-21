package cmd

import (
	"fmt"

	"github.com/IainHenn/chat/client/internal/api"
	"github.com/IainHenn/chat/client/internal/config"
	"github.com/spf13/cobra"
)

var viewRoomsCmd = &cobra.Command{
	Use:   "view-rooms",
	Short: "List all rooms",
	RunE: func(cmd *cobra.Command, args []string) error {
		if _, err := config.Load(); err != nil {
			return err
		}

		server, err := config.ServerURL(serverFlag)
		if err != nil {
			return err
		}

		rooms, err := api.New(server).ViewRooms()
		if err != nil {
			return err
		}

		if len(rooms) == 0 {
			fmt.Println("No rooms.")
			return nil
		}

		for _, room := range rooms {
			fmt.Printf("- %s (created %s)\n", room.Name, room.CreatedAt)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(viewRoomsCmd)
}
