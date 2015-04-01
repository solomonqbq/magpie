package core

import (
	"log"
	"time"
)

type RUNNING_TYPE int //0:周期任务 1:一次性任务

const (
	//看板定时刷新频度
	BOARD_REFRESH_INTERVAL = 5 * time.Second
)

type Worker struct {
	running bool
	Id      string //ID

	Init              func() error                                       //初始化方法
	SelectLeader      func(group string) bool                            //选举组长
	DispatchTasks     func(workerIds []string, tasks []*Task) error      //分配任务
	LoadAllGroup      func() (groups []string, err error)                //加载所有组
	LoadActiveWorkers func(group string) (workerIds []string, err error) //加载最新活动worker
	LoadTasks         func(group string) (tasks []*Task, err error)      //加载最新任务
	Cleanup           func(group string)                                 //清理
	HeartBeat         func()                                             //心跳
	TakeTasks         func()                                             //领任务
}

func NewWorker() *Worker {
	b := new(Worker)
	b.running = false
	return b
}

func (b *Worker) Start() error {
	b.running = true
	err := b.Init()
	if err != nil {
		return err
	}

	//定时选举并分配任务
	go func() {
		for b.running {
			b.selectLeaderAndDispatch()
			time.Sleep(BOARD_REFRESH_INTERVAL)
		}
	}()

	//定时心跳
	go func() {
		for b.running {
			b.HeartBeat()
			time.Sleep(BOARD_REFRESH_INTERVAL)
		}
	}()

	//定时获取任务
	go func() {
		for b.running {
			b.TakeTasks()
			time.Sleep(BOARD_REFRESH_INTERVAL)
		}
	}()

	return nil
}

//务选举组长并分派任务
func (b *Worker) selectLeaderAndDispatch() {
	groups, err := b.LoadAllGroup()
	if err != nil {
		log.Printf("获取分组信息出错！%s", err)
		return
	}
	//分组抢锁
	counter := len(groups)
	if counter == 0 {
		return
	}
	signal := make(chan int, 1)
	for _, group := range groups {
		go func() {
			leader := b.SelectLeader(group)
			if leader {
				go b.Cleanup(group)
				//成为组长
				ids, _ := b.LoadActiveWorkers(group)
				for _, id := range ids {
					log.Printf("active member Id:%s group:%s", id, group)
				}

				tasks, _ := b.LoadTasks(group)
				for _, t := range tasks {
					log.Printf("active task Id:%s", t.ID)
				}

				err := b.DispatchTasks(ids, tasks)
				if err != nil {
					log.Println(err)
				}
			}
			signal <- 1
		}()
	}
	//等待所有的组都完成
	for counter > 0 {
		<-signal
		counter = counter - 1
	}
}

func (b *Worker) IsStarted() bool {
	return b.running
}
