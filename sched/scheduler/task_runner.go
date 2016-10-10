package scheduler

import (
	"github.com/scootdev/scoot/common/stats"
	"github.com/scootdev/scoot/saga"
	"github.com/scootdev/scoot/sched"
	"github.com/scootdev/scoot/sched/worker"
	"github.com/scootdev/scoot/workerapi"
)

// Run the task on the specified worker, and update the SagaLog appropriately.  Returns an error if one
// occurs while running a task or writing to the SagaLog.  This method blocks until all saga messages
// are logged and the task completes
func runTaskAndLog(saga *saga.Saga, worker worker.Worker, taskId string, task sched.TaskDefinition, stat stats.StatsReceiver) error {
	// Log StartTask Message to SagaLog
	err := saga.StartTask(taskId, nil)
	if err != nil {
		return err
	}
	// runtask on worker
	processStatus, err := worker.RunAndWait(task)
	if err != nil {
		stat.Counter("failedTaskRunCounter").Inc(1)
		return err
	}

	statusAsBytes, err := workerapi.SerializeProcessStatus(processStatus)
	if err != nil {
		stat.Counter("failedTaskSerializeCounter").Inc(1)
		return err
	}

	// Log EndTask Message to SagaLog
	err = saga.EndTask(taskId, statusAsBytes)
	if err != nil {
		stat.Counter("failedTaskSagaCounter").Inc(1)
	} else {
		stat.Counter("completedTaskCounter").Inc(1)
	}

	return err
}
