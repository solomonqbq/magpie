package core

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

//key是taskName,value是executor
var executor_map map[string]*Task_Executor = make(map[string]*Task_Executor)

func Registry(task_name string, task_executor *Task_Executor) {
	executor_map[task_name] = task_executor
}

func GetTaskExecutor(task_name string) *Task_Executor {
	v, ok := executor_map[task_name]
	if !ok {
		return nil
	}
	return v
}
