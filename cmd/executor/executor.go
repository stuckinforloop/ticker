package executor

import (
	"github.com/spf13/cobra"
	taskexec "github.com/stuckinforloop/ticker/internal/task_exec"
	"github.com/stuckinforloop/ticker/worker"
)

// ExecutorCmd represents the executor command
var ExecutorCmd = &cobra.Command{
	Use:   "executor",
	Short: "starts executor for ticker",
	Run: func(cmd *cobra.Command, args []string) {
		w := worker.New()
		taskExecDAO := taskexec.NewTaskExecDAO(w.DAO)
		w.Run(taskExecDAO.ExecuteTasks)
	},
}

func init() {
}
