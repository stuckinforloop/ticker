package executor

import (
	"log"

	"github.com/spf13/cobra"
)

// ExecutorCmd represents the executor command
var ExecutorCmd = &cobra.Command{
	Use:   "executor",
	Short: "starts executor for ticker",
	Run: func(cmd *cobra.Command, args []string) {
		if err := cmd.Help(); err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
}
