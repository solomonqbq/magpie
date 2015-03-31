package task

import (
	"fmt"
	"github.com/xeniumd-china/flamingo/log"
	"github.com/xeniumd-china/xeniumd-monitor/dao/model"
)

const (
	RETRY = iota
	INTERRUPTED
	ERROR
	UNSUPPORTED
	SUCCESS
)

const (
	TASK_NEW = iota
	TASK_DISPATCHED
	TASK_RUNNING
	TASK_FAIL
	TASK_SUCCESS
	TASK_ERROR
)

var (
	logger            = log.GetLogger("task")
	executorFactories = make(map[string]ExecutorFactory)
)

func Registry(code string, factory ExecutorFactory) {
	executorFactories[code] = factory
}

type TaskResult struct {
	ResultCode uint8
	Task       model.Task
	Error      error
}

func NewTaskResult(resultCode uint8, task model.Task, error error) *TaskResult {
	result := new(TaskResult)
	result.ResultCode = resultCode
	result.Task = task
	result.Error = error

	return result
}

type TaskContext struct {
	parameters map[string]interface{}
}

func NewContext(parameters map[string]interface{}) *TaskContext {
	context := new(TaskContext)
	context.parameters = make(map[string]interface{})

	for key, value := range parameters {
		context.parameters[key] = value
	}
	return context
}

func (this *TaskContext) SetString(name string, value string) {
	this.parameters[name] = value
}

func (this *TaskContext) GetString(name string, defaultValue string) string {
	switch value := this.parameters[name].(type) {
	case string:
		return value
	default:
		return defaultValue
	}
}

func (this *TaskContext) GetObject(name string) interface{} {
	result, exist := this.parameters[name]
	if exist {
		return result
	} else {
		return nil
	}
}

type ExecutorFactory func() TaskExecutor

type TaskDispatchRegistry interface {
	Start(name string) error

	Stop()

	RegisterDispatch(dispatch Dispatch)

	RegisterOnDispatch(onDispatch OnDispatch)

	RegisterOnCancel(onCancel OnCancel)
}

type Dispatch func(*DispatchContext) []*model.Task

type OnDispatch func(TaskExecutor) *TaskResult

type OnCancel func(TaskExecutor)

type DispatchContext struct {
	AvailableNodes []*model.TaskNode

	NormalNodeTasks []*NodeTasks
	DaemonNodeTasks []*NodeTasks

	RetryTasks     []*model.Task
	DeadTasks      []*model.Task
	NewTasks       []*model.Task
	NewDaemonTasks []*model.Task
}

type NodeTasks struct {
	NodeName string
	Limit    int
	Tasks    []*model.Task
}

func (this *NodeTasks) String() string {
	return fmt.Sprintf("{%s: %d/%d}", this.NodeName, len(this.Tasks), this.Limit)
}
