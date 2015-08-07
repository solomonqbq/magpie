package core

import ()

type Worker struct {
	running           bool
	Id                string                                             //ID
	Group             string                                             //组
	Executors         map[string]*Task_Executor                          //key是taskId
	Init              func() error                                       //初始化方法
	SelectLeader      func(group string) bool                            //选举组长
	DispatchTasks     func(workerIds []string, tasks []*Task) error      //分配任务
	LoadAllGroup      func() (groups []string, err error)                //加载所有组
	LoadActiveWorkers func(group string) (workerIds []string, err error) //加载最新活动worker
	LoadTasks         func(group string) (tasks []*Task, err error)      //加载最新任务
	Cleanup           func(group string)                                 //清理
	HeartBeat         func() error                                       //心跳
	TakeTasks         func() (tasks []*Task, err error)                  //领任务
}

func NewWorker(group string) *Worker {
	b := new(Worker)
	b.running = false
	b.Executors = make(map[string]*Task_Executor, 0)
	b.Group = group
	return b
}
