package core

import (
	"log"
	"math/rand"
	"strconv"
)

func NewMockBoard() *Worker {
	b := NewWorker()

	b.Init = func() error {
		log.Printf("init")
		b.Id = "qbq"
		return nil
	}

	b.SelectLeader = func(group string) bool {
		return true
	}

	b.DispatchTasks = func(workerIds []string, tasks []*Task) error {
		log.Printf("开始DispatchTasks")
		for _, id := range workerIds {
			log.Printf("组员%s", id)
		}
		for _, t := range tasks {
			log.Printf("任务%s", t.Name)
		}
		log.Printf("DispatchTasks完毕")
		return nil
	}

	b.LoadAllGroup = func() (groups []string, err error) {
		groups = make([]string, 0)
		groups = append(groups, "test_group")
		return
	}

	b.LoadActiveWorkers = func(group string) (ids []string, err error) {
		ids = make([]string, 0)
		ids = append(ids, "1")
		return
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

	b.HeartBeat = func() error {
		log.Println("heart beat")
		return nil
	}

	b.Cleanup = func(group string) {
		log.Println("clean up done")
		return
	}
	b.TakeTasks = func() (tasks []*Task, err error) {
		log.Printf("%s的worker领取任务...", b.Id)
		return make([]*Task, 0), nil
	}
	return b
}
