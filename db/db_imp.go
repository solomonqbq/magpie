package db

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/xeniumd-china/flamingo/log"
	"github.com/xeniumd-china/magpie/core"
	"github.com/xeniumd-china/magpie/db/model"
	"github.com/xeniumd-china/magpie/global"
	"strconv"
	"strings"
	"time"
)

type WorkerTask struct {
	id             int64
	old_task_count int64
	new_task_id    []int64
}

func register(name string, group string) (id int64, err error) {
	log.Info("准备注册worker")
	id, err = InsertWorker(name, time.Duration(global.Properties.Int("magpie.worker.timeout.interval", 10))*time.Second, group)
	if err != nil {
		log.Error(err)
		return
	}
	log.Info("注册worker成功,ID:%d", id)
	return
}

func getName() string {
	//注册并获取ID
	useFirtIP := global.Properties.Bool("magpie.firstIP", true)
	var name string
	if useFirtIP {
		name = global.GetFirstLocalIP()
	} else {
		name = global.GetLocalIP()
	}
	return name
}

func Init(worker *core.Worker) func() error {
	return func() error {
		//构造DB连接
		InitAllDS(global.Properties)

		//注册并获取ID
		name := getName()
		id, err := register(name, worker.Group)
		if err != nil {
			log.Error(err)
			return err
		}
		worker.Id = strconv.Itoa(int(id))
		return nil
	}
}

func HeartBeat(w *core.Worker) func() error {
	return func() error {
		if w.Id != "" {
			id, _ := strconv.Atoi(w.Id)
			affected, err := UpdateWorkerTimeout(int64(id), time.Duration(global.Properties.Int("woker.timeout.interval", 10))*time.Second)
			if err == nil {
				if affected == 0 {
					log.Info("当前任务已被清除，需重新注册...")
					//重新注册
					id, err := register(getName(), w.Group)
					if err != nil {
						log.Error("注册失败:%s", err)
						return err
					}
					w.Id = strconv.Itoa(int(id))
				}
				log.Debug("%s完成心跳", w.Id)
			} else {
				log.Debug("%s心跳失败", w.Id)
				return err
			}
		}
		return nil
	}
}

func Cleanup(w *core.Worker) func(group string) {
	return func(group string) {
		//清理超时worker
		workersId, err := QueryTimeoutWorker()
		if workersId == nil || len(workersId) == 0 {
			log.Debug("清理完成..")
			return
		}
		//清理worker
		count, err := DeleteTimeoutWorker(workersId)
		if err != nil {
			log.Error("清除超时worker出错！%s", err)
		}
		if count != 0 {
			log.Info("清除%d个超时worker", count)
		}
	}
}

func LoadTasks() func(group string) (tasks []*core.Task, err error) {
	return func(group string) (tasks []*core.Task, err error) {
		mp_tasks, err := QueryNewAndFailedTasks(group)
		if err != nil {
			return nil, err
		}
		tasks = copyTasks(mp_tasks)
		log.Debug("组%s loadTasks:%d", group, len(tasks))
		return
	}
}

func SelectLeader(w *core.Worker) func(group string) bool {
	return func(group string) bool {
		id, err := strconv.Atoi(w.Id)
		if err != nil {
			log.Error("id不是个数字:%s", w.Id)
			return false
		}
		log.Debug("worker:%d准备选举组%s的组长", id, group)
		affect, err := UpdateWorkerGroup(group, int64(id), time.Duration(global.Properties.Int("woker.timeout.interval", 10))*time.Second)
		if err != nil {
			log.Error(err)
			return false
		} else {
			log.Debug("组%s的选举结果%s,", group, strconv.FormatBool(affect >= 1))
			return affect >= 1
		}
	}
}

func DispatchTasks(w *core.Worker) func(workerIds []string, tasks []*core.Task) error {
	return func(workerIds []string, tasks []*core.Task) error {
		if tasks == nil || len(tasks) == 0 {
			return nil
		}
		if workerIds == nil || len(workerIds) == 0 {
			log.Warn("无可用worker")
			return nil
		}

		//查询worker已分配的任务
		ids, taskCount, err := QueryActiveTasks()
		if err != nil {
			log.Error(err)
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
}

func TakeTasks(w *core.Worker) func() (tasks []*core.Task, err error) {
	return func() (tasks []*core.Task, err error) {
		mp_tasks, err := QueryDispatchedTasksByWorker(w.Id)
		if err != nil {
			log.Error("领取任务出错！%s", err)
		}
		tasks = copyTasks(mp_tasks)
		return tasks, nil
	}
}

func NewDBWorker(group string) *core.Worker {
	w := core.NewWorker(group)

	//初始化方法
	w.Init = Init(w)

	//心跳实现
	w.HeartBeat = HeartBeat(w)

	//清理
	w.Cleanup = Cleanup(w)

	//查组
	w.LoadAllGroup = func() (groups []string, err error) {
		return QueryAllGroup()
	}

	//加载所有存活的用户
	w.LoadActiveWorkers = func(group string) (ids []string, err error) {
		ids, err = QueryActiveWorkers(group)
		if err != nil {
			return nil, err
		}
		return
	}

	//加载所有任务
	w.LoadTasks = LoadTasks()

	//竞选组长
	w.SelectLeader = SelectLeader(w)

	//分配任务
	w.DispatchTasks = DispatchTasks(w)

	//领任务
	w.TakeTasks = TakeTasks(w)

	return w
}

func LeastTaskWorker(wts []*WorkerTask) *WorkerTask {
	var tmp *WorkerTask = nil
	for _, wt := range wts {
		if tmp == nil {
			tmp = wt
		}
		if (tmp.old_task_count + int64(len(tmp.new_task_id))) > (wt.old_task_count + int64(len(wt.new_task_id))) {
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
		//逗号分割
		if mp_t.Context != "" {
			params := strings.Split(mp_t.Context, ",")
			for _, p := range params {
				str := strings.SplitN(p, "=", 2)
				if len(str) == 2 {
					t.Context[str[0]] = str[1]
				}
			}
		}
		t.Status = mp_t.Status
		t.Running_type = core.RUNNING_TYPE(mp_t.Run_type)
		t.Interval = time.Duration(mp_t.Interval) * time.Second
		t.UpdateStatus = func(status int32, err error) error {
			tid, _ := strconv.ParseInt(t.ID, 0, 64)
			return UpdateTaskStatusErr(tid, int(status), err)
		}

		tasks = append(tasks, t)
	}
	return tasks
}
