package core

import (
	"log"
	"math/rand"
	"strconv"
	"time"
)

func NewMockBoard() *Board {
	b := NewBoard()
	b.Init = func() error {
		log.Printf("init")
		return nil
	}
	b.HeartBeat = func() {
		log.Println("heart beat")
	}

	b.LoadAllGroup = func() (groups []string, err error) {
		groups = make([]string, 0)
		groups = append(groups, "test_group")
		return
	}
	b.LoadActiveMembers = func(group string) (mems []*MemID, err error) {
		m := new(MemID)
		m.Id = "1"
		m.Group = "test_group"
		mems = append(mems, m)
		return mems, nil
	}
	b.LoadTasks = func(group string) (tasks []*Task, err error) {
		tasks = make([]*Task, 0)
		c := rand.Intn(10)
		for i := 0; i <= c; i++ {
			t := NewTask(strconv.Itoa(i), "task_name_"+strconv.Itoa(i), "test_group")
			tasks = append(tasks, t)
		}
		log.Printf("memMockBoard LoadTasks finished")
		return
	}
	b.SelectLeader = func(group string) bool {
		return true
	}
	b.DispatchTasks = func(mems []*MemID, tasks []*Task) {
		log.Printf("开始DispatchTasks")
		for _, m := range mems {
			log.Printf("组员%s", m.Id)
		}
		for _, t := range tasks {
			log.Printf("组员%s", t.Name)
		}
		log.Printf("DispatchTasks完毕")
		//		b.NewTasks = make(map[string][]*Task, 0)
	}
	b.Cleanup = func(group string) {
		log.Println("nothing to do")
		return
	}
	return b
}

func NewMockMember(b *Board, group string) *Member {
	m := NewMember(b, group)
	m.HeartBeat_interval = 3 * time.Second
	m.ReportTasksStaus = func() error {
		log.Printf("%s上报任务状态", m.Id)
		return nil
	}
	m.HeartBeat = func() {
		log.Printf("%s发送心跳", m.Id)
	}
	m.Regist = func() (id string, err error) {
		return "1", nil
	}
	m.Start()
	return m
}
