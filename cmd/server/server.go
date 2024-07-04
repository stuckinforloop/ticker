package server

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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
	ServerCmd.Flags().IntP("port", "p", 9000, "port to listen for requests")
	viper.GetViper().BindPFlag("port", ServerCmd.Flags().Lookup("port"))
}
