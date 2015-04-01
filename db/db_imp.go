package db

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/xeniumd-china/magpie/core"
	"github.com/xeniumd-china/magpie/db/model"
	"github.com/xeniumd-china/magpie/global"
	"log"
	"strconv"
	"strings"
	"time"
)

type WorkerTask struct {
	id             int64
	old_task_count int64
	new_task_id    []int64
}

func NewDBWorker() *core.Worker {
	b := core.NewWorker()

	b.Init = func() error {
		//构造DB连接
		InitAllDS(global.Properties)

		//注册并获取ID
		useFirtIP := global.Properties.Bool("magpie.firstIP", true)
		var name string
		if useFirtIP {
			name = global.GetFirstLocalIP()
		} else {
			name = global.GetLocalIP()
		}
		log.Printf("准备注册worker")
		id, err := InsertWorker(name, time.Duration(global.Properties.Int("woker.timeout.interval", 10))*time.Second)
		if err != nil {
			log.Println(err)
			return err
		}
		log.Printf("注册worker成功,ID:%d", id)
		b.Id = strconv.Itoa(int(id))
		return nil
	}

	b.HeartBeat = func() {
		if b.Id != "" {
			id, _ := strconv.Atoi(b.Id)
			err := UpdateWorkerTimeout(int64(id), time.Duration(global.Properties.Int("woker.timeout.interval", 10))*time.Second)
			if err == nil {
				log.Printf("%s完成心跳", b.Id)
			} else {
				log.Printf("%s心跳失败", b.Id)
			}
		}

	}

	b.Cleanup = func(group string) {
		workersId, err := QueryTimeoutWorker()
		if workersId == nil || len(workersId) == 0 {
			log.Println("清理完成..")
			return
		}
		//清理worker
		count, err := DeleteTimeoutWorker(workersId)
		if err != nil {
			log.Printf("清除超时worker出错！%s", err)
		}
		if count != 0 {
			log.Printf("清除%d个超时worker", count)
		}
	}

	b.LoadAllGroup = func() (groups []string, err error) {
		return QueryAllGroup()
	}

	b.LoadActiveWorkers = func(group string) (ids []string, err error) {
		ids, err = QueryActiveWorkers()
		if err != nil {
			return nil, err
		}
		return
	}

	b.LoadTasks = func(group string) (tasks []*core.Task, err error) {
		mp_tasks, err := QueryNewAndFailedTasks(group)
		if err != nil {
			return nil, err
		}
		log.Printf("loadTasks bc:%d", len(tasks))
		tasks = copyTasks(mp_tasks)
		log.Printf("loadTasks:%d", len(tasks))
		return
	}
	b.SelectLeader = func(group string) bool {
		id, err := strconv.Atoi(b.Id)
		if err != nil {
			log.Printf("id不是个数字:%s", b.Id)
			return false
		}
		log.Printf("worker:%d准备选举组%s的组长", id, group)
		affect, err := UpdateWorkerGroup(group, int64(id), time.Duration(global.Properties.Int("woker.timeout.interval", 10))*time.Second)
		if err != nil {
			log.Println(err)
			return false
		} else {
			log.Printf("选举结果%s,", strconv.FormatBool(affect >= 1))
			return affect >= 1
		}
	}

	b.DispatchTasks = func(workerIds []string, tasks []*core.Task) error {
		if tasks == nil || len(tasks) == 0 {
			log.Println("无任务分配...")
			return nil
		}
		if workerIds == nil || len(workerIds) == 0 {
			log.Println("无可用worker")
			return nil
		}

		//查询worker已分配的任务
		ids, taskCount, err := QueryActiveTasks()
		if err != nil {
			log.Println(err)
		}
		//均分任务
		wtm := make(map[int64]*WorkerTask, 0) //key是workerID
		for _, wid := range workerIds {
			id, _ := strconv.Atoi(wid)
			wt := newWorkerTask(int64(id))
			wtm[int64(id)] = wt
		}
		for index, id := range ids {
			wt := newWorkerTask(int64(id))
			wt.old_task_count = taskCount[index]
			wtm[id] = wt
		}
		wts := make([]*WorkerTask, 0)
		for _, wt := range wtm {
			wts = append(wts, wt)
		}

		for _, task := range tasks {
			ltw := LeastTaskWorker(wts)
			taskId, _ := strconv.Atoi(task.ID)
			ltw.new_task_id = append(ltw.new_task_id, int64(taskId))
		}

		return DispathTask(wts)
	}

	b.TakeTasks = func() {
		mp_tasks, err := QueryDispatchedTasksByWorker(b.Id)
		if err != nil {
			log.Printf("领取任务出错！%s", err)
		}
		//TODO
		tasks := copyTasks(mp_tasks)
		for _, t := range tasks {
			log.Printf("分配任务%d", t.ID)
		}
		for _, t := range tasks {
			id, _ := strconv.Atoi(t.ID)
			err := UpdateTaskStatus(id, core.TASK_RUNNING)
			if err != nil {
				log.Printf("更新任务状态出错！")
			}
		}
	}
	return b
}

func LeastTaskWorker(wts []*WorkerTask) *WorkerTask {
	var tmp *WorkerTask = nil
	for _, wt := range wts {
		if tmp == nil {
			tmp = wt
		}
		if (tmp.old_task_count + int64(len(tmp.new_task_id))) < (wt.old_task_count + int64(len(wt.new_task_id))) {
			tmp = wt
		}
	}
	return tmp
}

func newWorkerTask(id int64) *WorkerTask {
	wt := new(WorkerTask)
	wt.id = id
	wt.new_task_id = make([]int64, 0)
	wt.old_task_count = 0
	return wt
}

func copyTasks(mp_tasks []*model.Mp_task) []*core.Task {
	tasks := make([]*core.Task, 0)
	for _, mp_t := range mp_tasks {
		t := new(core.Task)
		t.ID = strconv.Itoa(int(mp_t.Id))
		t.Group = mp_t.Group
		t.Name = mp_t.Name
		t.Context = make(map[string]interface{}, 0)
		//分号分割
		if mp_t.Context != "" {
			params := strings.Split(mp_t.Context, ",")
			for _, p := range params {
				str := strings.SplitN(p, "=", 2)
				t.Context[str[0]] = str[1]
			}
		}
		t.Status = mp_t.Status
		t.Running_type = core.RUNNING_TYPE(mp_t.Run_type)
		t.Interval = time.Duration(mp_t.Interval) * time.Second
		tasks = append(tasks, t)
	}
	return tasks
}
