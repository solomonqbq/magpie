package task

import (
	"fmt"
	"github.com/xeniumd-china/xeniumd-monitor/dao"
	"github.com/xeniumd-china/xeniumd-monitor/dao/model"
	"github.com/xeniumd-china/xeniumd-monitor/leader"
	"testing"
	"time"
)

func TestDbDispatch(t *testing.T) {
	err := dao.InitDataSource(fmt.Sprintf("root:@tcp(127.0.0.1:3306)/eagle?charset=utf8"))
	if err != nil {
		fmt.Println(err)
	}

	registryExecutorFactory()
	parameters := make(map[string]interface{})
	parameters["test"] = "sdfsdf"

	connection, _ := dao.GetDataSource()
	leaderSelector, _ := leader.NewDbLeaderSelector(connection, "task.dispatcher", "test_node1")
	leaderSelector.Start()

	registry := NewDbTaskDispatchRegistry(leaderSelector)
	registry.Start("test_node1")

	container := NewTaskContainer(registry, parameters)
	err = container.Start()
	if err != nil {
		fmt.Println(err)
		return
	}
	for {
		time.Sleep(5 * time.Second)
	}
}

func registryExecutorFactory() {
	Registry("test", func() TaskExecutor {
		executor := new(TestTaskExecutor)
		executor.code = "test"
		return executor
	})
}

type TestTaskExecutor struct {
	code string
	task *model.Task
}

func (this *TestTaskExecutor) Initialize(task *model.Task) error {
	this.task = task
	fmt.Println(fmt.Sprintf("task:%v", task))
	return nil
}

func (this *TestTaskExecutor) Execute(context *TaskContext) *TaskResult {
	fmt.Println(context.GetString("test", "default"))
	return nil
}

func (this *TestTaskExecutor) Interrupt() {

}

func (this *TestTaskExecutor) GetType() string {
	return this.code
}

func (this *TestTaskExecutor) GetTask() *model.Task {
	return this.task
}
