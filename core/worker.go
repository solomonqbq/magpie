package core

import (
	"errors"
	l "github.com/xeniumd-china/flamingo/log"
	"github.com/xeniumd-china/magpie/global"
	"runtime/debug"
	"time"
)

type RUNNING_TYPE int //0:周期任务 1:一次性任务

const (
	//看板定时刷新频度
	BOARD_REFRESH_INTERVAL = 5 * time.Second
)

var (
	log = l.GetLogger("magpie")
)

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
		max_err_count := int(global.Properties.Int("worker.heartbeat.max.error.count", 3))
		//连续失败次数
		err_count := 0
		for b.running {
			err := b.HeartBeat()
			if err != nil {
				err_count++
			} else {
				err_count = 0
			}

			if err_count >= max_err_count {
				//启动失败处理机制
				for tid, te := range b.Executors {
					go Try(func() {
						te.Stop()
						te.Task.UpdateStatus(TASK_FAIL, errors.New("任务节点已死，放弃任务"+tid+"..."))
					})
				}
			}
			time.Sleep(BOARD_REFRESH_INTERVAL)
		}
	}()

	//定时获取任务
	go func() {
		for b.running {
			tasks, err := b.TakeTasks()
			if err != nil {
				log.Error("领取任务出错！%s", err)
				continue
			} else {
				//分配任务
				for _, t := range tasks {
					te := GetTaskExecutor(t.Name)
					if te == nil {
						log.Warn("不支持的任务:%s", t.Name)
						t.UpdateStatus(TASK_FAIL, errors.New("任务节点不支持名为"+t.Name+"的任务！"))
						continue
					} else {
						te.Task = t
						Try(func() {
							//启动新的te
							v, ok := b.Executors[t.ID]
							if ok {
								//已有任务，需要杀掉重新执行
								v.Stop()
							}
							b.Executors[t.ID] = te
							if err != nil {
								t.UpdateStatus(TASK_FAIL, err)
							} else {
								err := te.Init()
								if err != nil {
									te.Task.UpdateStatus(TASK_FAIL, errors.New("task executor初始化失败！"))
									return
								}
								r := te.Execute()
								err = te.Task.UpdateStatus(r.Result_code, r.Error)
								if err != nil {
									log.Error("任务%s状态更新失败,错误:%s", te.Task.ID, err)
								}
							}
						})
					}
				}
			}
			time.Sleep(BOARD_REFRESH_INTERVAL)
		}
	}()

	return nil
}

//选举组长并分派任务
func (b *Worker) selectLeaderAndDispatch() {
	groups, err := b.LoadAllGroup()
	if err != nil {
		log.Error("获取分组信息出错！%s", err)
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
					log.Debug("active member Id:%s group:%s", id, group)
				}

				tasks, _ := b.LoadTasks(group)
				for _, t := range tasks {
					log.Debug("active task Id:%s", t.ID)
				}

				err := b.DispatchTasks(ids, tasks)
				if err != nil {
					log.Error(err)
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

func Try(fun func()) {
	defer func() {
		if err := recover(); err != nil {
			debug.PrintStack()
			log.Error(err)
		}
	}()
	fun()
}
