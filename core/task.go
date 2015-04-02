package core

import "time"

const (
	TASK_NEW        = iota //新建任务
	TASK_DISPATCHED        //已分配
	TASK_RUNNING           //运行中
	TASK_FAIL              //失败
	TASK_SUCCESS           //成功
	TASK_ERROR             //错误
)

//任务
type Task struct {
	ID           string                              //任务ID
	Name         string                              //任务名
	Group        string                              //任务所属组
	Context      map[string]interface{}              //任务运行所需环境参数
	Running_type RUNNING_TYPE                        //运行类型，周期或一次
	Interval     time.Duration                       //运行间隔，只有Running_type是周期类型时才有效
	Status       int32                               //任务状态
	UpdateStatus func(status int32, err error) error //修改任务状态
}

func (t *Task) String() string {
	return "{taskId=" + t.ID + "}"
}

func NewTask(id string, name string, group string) *Task {
	t := new(Task)
	t.ID = id
	t.Name = name
	t.Group = group
	return t
}
