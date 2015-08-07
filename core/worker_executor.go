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
	//时间间隔
	BOARD_REFRESH_INTERVAL = 5 * time.Second
)

var (
	log = l.GetLogger("magpie")
)

type Worker_Executor struct {
	worker  *Worker
	running bool //运行状态
}

func NewWorkerExecutor(worker *Worker) *Worker_Executor {
	w_e := new(Worker_Executor)
	w_e.worker = worker
	w_e.running = false
	return w_e
}

//定时选举并分配任务
func (this *Worker_Executor) do_select_leader_and_dispatch() {
	for this.running {
		groups, err := this.worker.LoadAllGroup()
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
			go func(g string) {
				leader := this.worker.SelectLeader(g)
				if leader {
					go this.worker.Cleanup(g)
					//成为组长
					ids, _ := this.worker.LoadActiveWorkers(g)
					for _, id := range ids {
						log.Debug("active member Id:%s group:%s", id, g)
					}

					tasks, _ := this.worker.LoadTasks(g)
					for _, t := range tasks {
						log.Debug("active task Id:%s", t.ID)
					}

					err := this.worker.DispatchTasks(ids, tasks)
					if err != nil {
						log.Error(err)
					}
				}
				signal <- 1
			}(group)
		}
		//等待所有的组都完成
		for counter > 0 {
			<-signal
			counter = counter - 1
		}
		time.Sleep(BOARD_REFRESH_INTERVAL)
	}
}

func (this *Worker_Executor) do_heart_beat() {
	max_err_count := int(global.Properties.Int("magpie.worker.heartbeat.max.error.count", 3))
	//连续失败次数
	err_count := 0
	for this.running {
		err := this.worker.HeartBeat()
		if err != nil {
			err_count++
		} else {
			err_count = 0
		}

		if err_count >= max_err_count {
			//启动失败处理机制
			for tid, te := range this.worker.Executors {
				go Try(func() {
					te.Stop()
					te.Task.UpdateStatus(TASK_FAIL, errors.New("任务节点已死，放弃任务"+tid+"..."))
				})
			}
		}
		time.Sleep(BOARD_REFRESH_INTERVAL)
	}
}

func (this *Worker_Executor) do_occupy_tasks() {
	for this.running {
		tasks, err := this.worker.TakeTasks()
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
						v, ok := this.worker.Executors[t.ID]
						if ok {
							//已有任务，需要杀掉重新执行
							v.Stop()
						}
						this.worker.Executors[t.ID] = te
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
							log.Debug("更新任务%s的状态为%d", r.Task.ID, r.Result_code)
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
}

//启动
func (this *Worker_Executor) Start() error {
	this.running = true
	err := this.worker.Init()
	if err != nil {
		this.running = false
		return err
	}

	//选举并分配任务
	go this.do_select_leader_and_dispatch()

	//定时心跳
	go this.do_heart_beat()

	//定时获取任务
	go this.do_occupy_tasks()

	return nil
}

func (w_e *Worker_Executor) IsStarted() bool {
	return w_e.running
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
