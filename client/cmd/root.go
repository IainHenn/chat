package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var serverFlag string

var rootCmd = &cobra.Command{
	Use:   "chat",
	Short: "CLI client for the chat server",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&serverFlag, "server", "", "Chat server URL (default from config or http://localhost:8080)")
}
