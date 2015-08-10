package core

import "sync"

type Task_Executor struct {
	//当前任务
	Task        *Task
	Init        func() error   //初始化
	Execute     func() *Result //运行
	Stop        func() error   //停止
	GetTaskName func() string  //获取支持的任务名
}

type Result struct {
	Task        *Task
	Result_code int32 //任务状态
	Error       error
}

var lock = new(sync.RWMutex)

//第一层key是group,第二层key是taskName,value是executor
var executor_map map[string]map[string]*Task_Executor = make(map[string]map[string]*Task_Executor)

func Registry(group, task_name string, task_executor *Task_Executor) {
	lock.Lock()
	defer lock.Unlock()
	if _, ok := executor_map[group]; !ok {
		executor_map[group] = make(map[string]*Task_Executor)
	}
	executor_map[group][task_name] = task_executor
}

func GetTaskExecutor(group, task_name string) *Task_Executor {
	lock.RLock()
	defer lock.RUnlock()
	if _, ok := executor_map[group]; !ok {
		return nil
	} else {
		return executor_map[group][task_name]
	}
}
