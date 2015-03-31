package task

import (
	"errors"
	"github.com/xeniumd-china/flamingo/lang"
	"github.com/xeniumd-china/xeniumd-monitor/dao"
	"reflect"
	"time"
)

const (
	CONFIG_GROUP = "task.dispatch"
)

type TaskContainer struct {
	start bool

	dispatchRegistry TaskDispatchRegistry

	parameters map[string]interface{}
}

func (this *TaskContainer) Start() error {
	if this.dispatchRegistry == nil {
		return errors.New("dispatchRegistry is null")
	}
	if this.parameters == nil {
		return errors.New("parameters is null")
	}
	if !this.start {
		this.start = true

		//首次运行加载动态参数
		this.refreshParameters()

		//注册任务分发和打断函数
		this.dispatchRegistry.RegisterOnDispatch(func(executor TaskExecutor) *TaskResult {
			return this.execute(executor)
		})
		this.dispatchRegistry.RegisterOnCancel(func(executor TaskExecutor) {
			this.interrupt(executor)
		})

		//开启动态参数定时刷新
		go func() {
			timer := time.NewTicker(60 * time.Second)
			for this.start {
				<-timer.C
				go this.refreshParameters()
			}
		}()
	}
	return nil
}

func (this *TaskContainer) refreshParameters() {
	configs, err := dao.FindConfigsByGroup(CONFIG_GROUP)
	if err != nil {
		temp := make(map[string]interface{})
		for key, value := range this.parameters {
			temp[key] = value
		}
		for _, config := range configs {
			temp[config.Key] = config.Value
		}
		this.parameters = temp
	}
}

func (this *TaskContainer) execute(executor TaskExecutor) (result *TaskResult) {
	task := executor.GetTask()
	if this.start {
		context := NewContext(this.parameters)

		//将URL里面的参数注入到context中
		url, err := Parse(task.Url)
		if err != nil {
			result = NewTaskResult(ERROR, *task, errors.New("task url is unvaild"))
			return
		}

		for key, value := range url.Parameters {
			context.SetString(key, value)
		}

		//		//拦截panic
		//		defer func() {
		//			logger.Info("Task %s end", task)
		//			if message := recover(); message != nil {
		//
		//			}
		//		}()

		logger.Info("Task %s start", task)
		message, trace := lang.CatchTrace(func() {
			result = executor.Execute(context)
		})
		if message != nil {
			logger.Error(*message + "\n" + trace)
			result = NewTaskResult(RETRY, *task, errors.New(reflect.TypeOf(message).String()))
		}
		if result == nil {
			return NewTaskResult(RETRY, *task, errors.New("Unknown result"))
		}
		switch result.ResultCode {
		case RETRY:
			logger.Error(result.Error)
		case ERROR:
			logger.Error(result.Error)
		}
	} else {
		result = NewTaskResult(RETRY, *task, errors.New("Container is not start yet"))
	}
	return
}

func (this *TaskContainer) interrupt(executor TaskExecutor) {
	if executor != nil {
		executor.Interrupt()
	}
}

func NewTaskContainer(dispatchRegistry TaskDispatchRegistry, parameters map[string]interface{}) *TaskContainer {
	container := new(TaskContainer)
	container.parameters = parameters
	container.dispatchRegistry = dispatchRegistry
	return container
}
