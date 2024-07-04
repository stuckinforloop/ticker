package notifier

import (
	"github.com/spf13/cobra"
	taskexec "github.com/stuckinforloop/ticker/internal/task_exec"
	"github.com/stuckinforloop/ticker/worker"
)

// NotifierCmd represents the notifier command
var NotifierCmd = &cobra.Command{
	Use:   "notifier",
	Short: "starts notifier for ticker",
	Run: func(cmd *cobra.Command, args []string) {
		w := worker.New()
		taskExecDAO := taskexec.NewTaskExecDAO(w.DAO)
		w.Run(taskExecDAO.UpdateTaskStatusNotify)
	},
}

func init() {}
