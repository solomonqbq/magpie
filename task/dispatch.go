package task

import (
	"errors"
	"github.com/xeniumd-china/xeniumd-monitor/dao/model"
)

const (
	TASK_GROUP = "Monitor"
)

type TaskDispatcher struct {
	dispatchRegistry TaskDispatchRegistry
}

func (this *TaskDispatcher) Start() error {
	if this.dispatchRegistry == nil {
		return errors.New("dispatchRegistry is null")
	}
	//注册任务分发方法
	this.dispatchRegistry.RegisterDispatch(this.dispatch)
	return nil
}

func (this *TaskDispatcher) Stop() {
	this.dispatchRegistry.Stop()
}

func (this *TaskDispatcher) dispatch(dispatchContext *DispatchContext) []*model.Task {
	if dispatchContext == nil {
		return nil
	}

	if logger.IsDebugEnable() {
		logger.Debug("Dispatch runs. Alive nodes: %s", dispatchContext.AvailableNodes)
		logger.Debug("Normal tasks:%s", dispatchContext.NormalNodeTasks)
		logger.Debug("Daemon tasks:%s", dispatchContext.DaemonNodeTasks)
	}

	normalTasks := make([]*model.Task, 0)
	daemonTasks := make([]*model.Task, 0)

	//获取失效节点中已分发的任务
	for _, deadTask := range dispatchContext.DeadTasks {
		if deadTask.Daemon {
			daemonTasks = append(daemonTasks, deadTask)
		} else {
			normalTasks = append(normalTasks, deadTask)
		}
	}

	//获取新任务
	normalTasks = append(normalTasks, dispatchContext.NewTasks[:]...)
	daemonTasks = append(daemonTasks, dispatchContext.NewDaemonTasks[:]...)

	//分发通常任务
	updatedTasks := dispatch(dispatchContext.NormalNodeTasks, normalTasks)

	//重分配后台任务
	updatedTasks = append(updatedTasks, dispatchDaemon(dispatchContext.DaemonNodeTasks, daemonTasks)[:]...)

	if logger.IsDebugEnable() {
		logger.Debug("Normal tasks after dispatched:%s", dispatchContext.NormalNodeTasks)
		logger.Debug("Daemon tasks after dispatched:%s", dispatchContext.DaemonNodeTasks)
	}

	return updatedTasks
}

func getMinSize(tasks []*NodeTasks) *NodeTasks {
	size := len(tasks)
	if size == 0 {
		return nil
	}
	var result *NodeTasks
	for i := 0; i < size; i++ {
		if result == nil {
			result = tasks[i]
		}
		if len(result.Tasks) >= len(tasks[i].Tasks) {
			result = tasks[i]
		}
	}
	return result
}

func getMaxSize(tasks []*NodeTasks) *NodeTasks {
	size := len(tasks)
	if size == 0 {
		return nil
	}
	var result *NodeTasks
	for i := 0; i < size; i++ {
		if result == nil {
			result = tasks[i]
		}
		if len(result.Tasks) <= len(tasks[i].Tasks) {
			result = tasks[i]
		}
	}
	return result
}

func dispatch(nodeTasks []*NodeTasks, tasks []*model.Task) []*model.Task {
	result := make([]*model.Task, 0)
	var nodeTask *NodeTasks
	for _, task := range tasks {
		//获取当前任务数最少的节点
		nodeTask = getMinSize(nodeTasks)
		if nodeTask == nil {
			break
		}
		//如果未达到最大执行上限，将任务分配给任务数最少的节点
		if len(nodeTask.Tasks) < nodeTask.Limit {
			task.Status = TASK_DISPATCHED
			task.Owner = nodeTask.NodeName
			result = append(result, task)
			nodeTask.Tasks = append(nodeTask.Tasks, task)
		} else {
			break
		}
	}

	return result
}

func dispatchDaemon(nodeTasks []*NodeTasks, tasks []*model.Task) []*model.Task {
	result := make([]*model.Task, 0)

	var nodeTask *NodeTasks
	for _, task := range tasks {
		nodeTask = getMinSize(nodeTasks)
		if nodeTask == nil {
			break
		}
		task.Status = TASK_DISPATCHED
		task.Owner = nodeTask.NodeName
		result = append(result, task)
		nodeTask.Tasks = append(nodeTask.Tasks, task)
	}

	minSizeNode := getMinSize(nodeTasks)
	maxSizeNode := getMaxSize(nodeTasks)
	minSize := len(minSizeNode.Tasks)
	maxSize := len(maxSizeNode.Tasks)
	var task *model.Task

	//将目前后台任务最多节点的一个任务分配给最少的，直到两个的差值小于等于1
	for maxSize-minSize > 1 {
		task = maxSizeNode.Tasks[0]
		task.Status = TASK_DISPATCHED
		task.Owner = minSizeNode.NodeName
		result = append(result, task)

		minSizeNode.Tasks = append(minSizeNode.Tasks, task)
		if maxSize > 1 {
			maxSizeNode.Tasks = maxSizeNode.Tasks[1:]
		} else {
			maxSizeNode.Tasks = make([]*model.Task, 0)
		}

		minSizeNode = getMinSize(nodeTasks)
		maxSizeNode = getMaxSize(nodeTasks)
		minSize = len(minSizeNode.Tasks)
		maxSize = len(maxSizeNode.Tasks)
	}

	return result
}

func NewTaskDispatcher(dispatchRegistry TaskDispatchRegistry) *TaskDispatcher {
	dispatcher := new(TaskDispatcher)
	dispatcher.dispatchRegistry = dispatchRegistry
	return dispatcher
}
