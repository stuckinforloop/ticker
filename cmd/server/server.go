package server

import (
	"log"

	"github.com/spf13/cobra"
)

// ServerCmd represents the server command
var ServerCmd = &cobra.Command{
	Use:   "server",
	Short: "starts server for ticker",
	Run: func(cmd *cobra.Command, args []string) {
		if err := cmd.Help(); err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
}
