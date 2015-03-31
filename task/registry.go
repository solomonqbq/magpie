package task

import (
	"github.com/xeniumd-china/xeniumd-monitor/common"
	"github.com/xeniumd-china/xeniumd-monitor/dao"
	"github.com/xeniumd-china/xeniumd-monitor/dao/model"
	"github.com/xeniumd-china/xeniumd-monitor/leader"
	"sync"
	"time"
)

const (
	HEARTBEAT_TIMEOUT  = 30
	MAX_RETRY_INTERVAL = 10 * time.Minute
)

type DbTaskDispatchRegistry struct {
	Name string

	closed         bool
	runningTasks   *TaskExecuteStatus
	leaderSelector leader.LeaderSelector

	dispatch   Dispatch
	onDispatch OnDispatch
	onCancel   OnCancel

	KeepALiveInterval int
	DispatchInterval  int
	LoadTasksInterval int
}

func NewDbTaskDispatchRegistry(leaderSelector leader.LeaderSelector) *DbTaskDispatchRegistry {
	result := new(DbTaskDispatchRegistry)
	result.leaderSelector = leaderSelector
	result.KeepALiveInterval = 5
	result.DispatchInterval = 5
	result.LoadTasksInterval = 5
	return result
}

func (this *DbTaskDispatchRegistry) RegisterDispatch(dispatch Dispatch) {
	this.dispatch = dispatch
}

func (this *DbTaskDispatchRegistry) RegisterOnDispatch(onDispatch OnDispatch) {
	this.onDispatch = onDispatch
}

func (this *DbTaskDispatchRegistry) RegisterOnCancel(onCancel OnCancel) {
	this.onCancel = onCancel
}

func (this *DbTaskDispatchRegistry) Start(name string) error {
	this.runningTasks = NewTaskExecuteStatus()
	this.Name = name

	//节点心跳
	go func() {
		timer := time.NewTicker(time.Duration(this.KeepALiveInterval) * time.Second)
		this.heartbeat()
		for !this.closed {
			<-timer.C
			go this.heartbeat()
		}
	}()

	//每隔5秒进行任务分发
	go func() {
		timer := time.NewTicker(time.Duration(this.DispatchInterval) * time.Second)
		for !this.closed {
			<-timer.C
			go func() {
				if !this.leaderSelector.IsLeader() {
					return
				}
				if this.dispatch == nil {
					return
				}
				logger.Debug("Start dispatch tasks")
				dispatchContext := this.getDispatchContext()
				updatedTasks := this.dispatch(dispatchContext)
				if updatedTasks != nil {
					for _, updatedTask := range updatedTasks {
						dao.DispatchTask(updatedTask)
					}
				}
			}()
		}
	}()

	//每隔5秒获取分配给自己的任务
	go func() {
		timer := time.NewTicker(time.Duration(this.LoadTasksInterval) * time.Second)
		for !this.closed {
			<-timer.C
			go func() {
				if this.onDispatch == nil || this.onCancel == nil {
					return
				}
				logger.Debug("Get tasks")
				this.findMyTasks()
			}()
		}
	}()
	return nil
}

func (this *DbTaskDispatchRegistry) Stop() {
	this.closed = true
}

func (this *DbTaskDispatchRegistry) heartbeat() {
	rowAffected, err := dao.TaskNodeHeartbeat(this.Name, HEARTBEAT_TIMEOUT)
	if err != nil {
		logger.Error(err)
		return
	}
	if rowAffected == 0 {
		taskNode := &model.TaskNode{
			Name:  this.Name,
			Group: TASK_GROUP,
			Size:  20,
			Type:  0,
		}
		err = dao.InsertTaskNode(taskNode, HEARTBEAT_TIMEOUT)
		if err != nil {
			logger.Error(err)
			return
		}
	}
}

func (this *DbTaskDispatchRegistry) getDispatchContext() *DispatchContext {
	dispatchContext := new(DispatchContext)

	var err error
	//获取存活的任务节点
	dispatchContext.AvailableNodes, err = dao.FindTaskNodeByGroup(TASK_GROUP)
	if err != nil {
		logger.Error(err)
		return nil
	}

	availableNodeSize := len(dispatchContext.AvailableNodes)
	if availableNodeSize == 0 {
		logger.Warn("No available task node")
		return nil
	}

	//获取当前节点的工作任务列表
	normalNodeTasks := make([]*NodeTasks, 0)
	daemonNodeTasks := make([]*NodeTasks, 0)
	var runningTasks []*model.Task
	var daemonTasks []*model.Task
	var normalTasks []*model.Task
	for _, node := range dispatchContext.AvailableNodes {
		runningTasks, err = dao.FindTaskByOwner(node.Name)
		daemonTasks = make([]*model.Task, 0)
		normalTasks = make([]*model.Task, 0)
		if err != nil {
			logger.Error(err)
			return nil
		}
		for _, task := range runningTasks {
			if task.Daemon {
				daemonTasks = append(daemonTasks, task)
			} else {
				normalTasks = append(normalTasks, task)
			}
		}
		normalNodeTasks = append(normalNodeTasks, &NodeTasks{node.Name, node.Size, normalTasks})
		daemonNodeTasks = append(daemonNodeTasks, &NodeTasks{node.Name, node.Size, daemonTasks})
	}

	dispatchContext.NormalNodeTasks = normalNodeTasks
	dispatchContext.DaemonNodeTasks = daemonNodeTasks

	//分发重试的任务
	dispatchContext.RetryTasks, err = dao.FindRetryTasks()
	if err != nil {
		logger.Error(err)
		return nil
	}

	//分发死节点的任务
	dispatchContext.DeadTasks = make([]*model.Task, 0)
	deadNodes, err := dao.FindDeadTaskNode(TASK_GROUP)
	if err != nil {
		logger.Error(err)
		return nil
	}

	for _, deadNode := range deadNodes {
		deadTasks, err := dao.FindTaskByOwner(deadNode.Name)
		if err != nil {
			logger.Error(err)
			return nil
		}
		dispatchContext.DeadTasks = append(dispatchContext.DeadTasks, deadTasks[:]...)
	}

	//分发新生成的任务
	newTasks, err := dao.FindNewTasks()
	if err != nil {
		logger.Error(err)
		return nil
	}
	dispatchContext.NewTasks = newTasks

	newDaemonTasks, err := dao.FindNewDaemonTasks()
	if err != nil {
		logger.Error(err)
		return nil
	}
	dispatchContext.NewDaemonTasks = newDaemonTasks

	return dispatchContext
}

func (this *DbTaskDispatchRegistry) notifyResult(result *TaskResult) {
	executor := this.runningTasks.Get(result.Task.Id)
	if executor != nil {
		switch result.ResultCode {
		case RETRY:
			executor.GetTask().Status = TASK_FAIL
			executor.GetTask().Exception = result.Error.Error()
		case INTERRUPTED:
			executor.GetTask().Status = TASK_FAIL
		case UNSUPPORTED:
			executor.GetTask().Status = TASK_ERROR
			executor.GetTask().Exception = result.Task.Code + " is not supported"
		case ERROR:
			executor.GetTask().Status = TASK_ERROR
			executor.GetTask().Exception = result.Error.Error()
		case SUCCESS:
			executor.GetTask().Status = TASK_SUCCESS
			executor.GetTask().Exception = ""
		default:
		}
		exception := executor.GetTask().Exception
		if len(exception) > 512 {
			executor.GetTask().Exception = exception[:500] + "..."
		}
	}
}

func (this *DbTaskDispatchRegistry) findMyTasks() {
	logger.Debug("Node[%s] load tasks from database", this.Name)
	tasks, err := dao.FindTaskByOwner(this.Name)
	if err != nil {
		logger.Error(err)
		return
	}
	runningTasks := this.runningTasks.Clone().tasks
	for _, task := range tasks {
		_, exist := runningTasks[task.Id]
		if !exist {
			if !this.execute(task) {
				this.notifyResult(NewTaskResult(UNSUPPORTED, *task, nil))
			}
		} else {
			delete(runningTasks, task.Id)
		}
	}

	for _, executor := range runningTasks {
		this.onCancel(executor)
	}
	this.syncTaskStatus()
}

func (this *DbTaskDispatchRegistry) execute(task *model.Task) bool {
	factory := executorFactories[task.Code]
	if factory != nil {
		executor := this.runningTasks.Get(task.Id)
		if executor != nil {
			return true
		}

		task.Status = TASK_RUNNING
		dao.UpdateTask(task)
		go func() {
			executor = factory()
			executor.Initialize(task)
			this.runningTasks.Put(task.Id, executor)
			result := this.onDispatch(executor)
			this.notifyResult(result)
		}()
		return true
	} else {
		return false
	}
}

func (this *DbTaskDispatchRegistry) syncTaskStatus() {
	logger.Debug("Node[%s] sync task status", this.Name)
	var task *model.Task
	cleanTaskIds := make([]uint64, 0)
	for _, value := range this.runningTasks.tasks {
		task = value.GetTask()
		switch task.Status {
		case TASK_FAIL:
			if !task.Daemon {
				retryInterval := time.Duration(30*task.Retry) * time.Second
				if retryInterval > MAX_RETRY_INTERVAL {
					retryInterval = MAX_RETRY_INTERVAL
				}
				task.Retry++
				task.RetryTime = common.ParseTime(task.RetryTime).Add(retryInterval).Format(common.FORMAT_SECOND)
			}
			cleanTaskIds = append(cleanTaskIds, task.Id)
		case TASK_SUCCESS:
			cleanTaskIds = append(cleanTaskIds, task.Id)
		case TASK_ERROR:
			cleanTaskIds = append(cleanTaskIds, task.Id)
		default:
		}
		_, err := dao.UpdateTask(task)
		if err != nil {
			logger.Error(err)
		}

		for _, id := range cleanTaskIds {
			this.runningTasks.Remove(id)
		}
	}
}

type TaskExecuteStatus struct {
	tasks map[uint64]TaskExecutor
	lock  *sync.RWMutex
}

func (this *TaskExecuteStatus) Get(key uint64) TaskExecutor {
	this.lock.RLock()
	defer this.lock.RUnlock()
	return this.tasks[key]
}

func (this *TaskExecuteStatus) Put(key uint64, value TaskExecutor) TaskExecutor {
	this.lock.Lock()
	defer this.lock.Unlock()
	this.tasks[key] = value
	return value
}

func (this *TaskExecuteStatus) PutIfAbsent(key uint64, value TaskExecutor) TaskExecutor {
	this.lock.Lock()
	defer this.lock.Unlock()
	executor, exist := this.tasks[key]
	if exist {
		return executor
	} else {
		this.tasks[key] = value
		return value
	}
}

func (this *TaskExecuteStatus) Remove(key uint64) TaskExecutor {
	this.lock.Lock()
	defer this.lock.Unlock()
	executor, exist := this.tasks[key]
	if exist {
		delete(this.tasks, key)
		return executor
	} else {
		return nil
	}
}

func (this *TaskExecuteStatus) Clone() *TaskExecuteStatus {
	this.lock.Lock()
	defer this.lock.Unlock()
	status := new(TaskExecuteStatus)
	status.tasks = make(map[uint64]TaskExecutor)
	status.lock = new(sync.RWMutex)
	for key, value := range this.tasks {
		status.tasks[key] = value
	}
	return status
}

func NewTaskExecuteStatus() *TaskExecuteStatus {
	status := new(TaskExecuteStatus)
	status.tasks = make(map[uint64]TaskExecutor)
	status.lock = new(sync.RWMutex)
	return status

}
