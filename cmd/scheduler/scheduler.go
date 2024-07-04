package scheduler

import (
	"github.com/spf13/cobra"
	taskexec "github.com/stuckinforloop/ticker/internal/task_exec"
	"github.com/stuckinforloop/ticker/worker"
)

// SchedulerCmd represents the scheduler command
var SchedulerCmd = &cobra.Command{
	Use:   "scheduler",
	Short: "starts scheduler for ticker",
	Run: func(cmd *cobra.Command, args []string) {
		w := worker.New()
		taskExecDAO := taskexec.NewTaskExecDAO(w.DAO)
		w.Run(taskExecDAO.ScheduleTasks)
	},
}

func init() {
}
