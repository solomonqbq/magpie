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

//任务看板
type Worker struct {
	running bool
	Id      string //ID

	Init              func() error                                  //初始化方法
	SelectLeader      func(group string) bool                       //选举组长
	DispatchTasks     func(mems []*MemID, tasks []*Task)            //分配任务
	LoadAllGroup      func() (groups []string, err error)           //加载所有组
	LoadActiveMembers func(group string) (mems []*MemID, err error) //加载最新活动组员
	LoadTasks         func(group string) (tasks []*Task, err error) //加载最新任务
	Cleanup           func(group string)                            //清理
	HeartBeat         func()                                        //心跳
}

type MemID struct {
	Id    string //Id
	Group string //所属组
}

func (m *MemID) String() string {
	return "{id=" + m.Id + ",group=" + m.Group + "}"
}

//组员
type Member struct {
	MemID
	Board              *Board                           //看板
	Capacity           int32                            //可承担任务数
	Regist             func() (memId string, err error) //注册
	ReportTasksStaus   func() error                     //上报任务状态
	InteruptTask       func(taskId string) bool         //中断任务
	HeartBeat          func()                           //心跳
	HeartBeat_interval time.Duration                    //心跳间隔
	running            bool                             //运行状态
}

func NewBoard() *Board {
	b := new(Board)
	b.running = false
	return b
}

func (b *Board) Start() error {
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

	return nil
}

//务选举组长并分派任务
func (b *Board) selectLeaderAndDispatch() {
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
				mems, _ := b.LoadActiveMembers(group)
				for _, m := range mems {
					log.Printf("active member Id:%s group:%s", m.Id, m.Group)
				}

				tasks, _ := b.LoadTasks(group)
				for _, t := range tasks {
					log.Printf("active task Id:%s", t.ID)
				}

				b.DispatchTasks(mems, tasks)
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

func (b *Board) IsStarted() bool {
	return b.running
}

func NewMember(b *Board, group string) *Member {
	m := new(Member)
	m.Board = b
	m.Group = group
	m.running = false
	return m
}

func (m *Member) Start() error {
	if m.Id == "" {
		m.Id, _ = m.Regist()
	}
	m.running = true
	go func() {
		for m.running {
			m.HeartBeat()
			time.Sleep(m.HeartBeat_interval)
		}
	}()
	return nil
}

func (m *Member) Stop() error {
	m.running = false
	return nil
}

func (m *Member) IsAlive() bool {
	return m.running
}
