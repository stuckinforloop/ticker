package scheduler

import (
	"log"

	"github.com/spf13/cobra"
)

// SchedulerCmd represents the scheduler command
var SchedulerCmd = &cobra.Command{
	Use:   "scheduler",
	Short: "starts scheduler for ticker",
	Run: func(cmd *cobra.Command, args []string) {
		if err := cmd.Help(); err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
}
