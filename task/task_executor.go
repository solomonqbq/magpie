package task

type TaskExecutor interface {
	Initialize(task *model.Task) error

	Execute(context *TaskContext) *TaskResult

	Interrupt()

	GetType() string

	GetTask() *model.Task
}
