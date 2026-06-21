package cmd

import (
	"fmt"

	"github.com/IainHenn/chat/client/internal/config"
	"github.com/spf13/cobra"
)

var signupCmd = &cobra.Command{
	Use:   "signup [username]",
	Short: "Create a local username for this machine",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		server := serverFlag
		if server == "" {
			server = "http://localhost:8080"
		}

		if err := config.Save(args[0], server); err != nil {
			return err
		}

		fmt.Printf("Signed up as %q (server: %s)\n", args[0], server)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(signupCmd)
}
