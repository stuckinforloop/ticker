package server

import (
	"github.com/spf13/cobra"
	"github.com/stuckinforloop/ticker/server"
)

// ServerCmd represents the server command
var ServerCmd = &cobra.Command{
	Use:   "server",
	Short: "starts server for ticker",
	Run: func(cmd *cobra.Command, args []string) {
		srv := server.New()
		srv.Start()
	},
}

func init() {
}
