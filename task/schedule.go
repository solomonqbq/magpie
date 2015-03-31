package task

import (
	"github.com/xeniumd-china/xeniumd-monitor/common"
	"github.com/xeniumd-china/xeniumd-monitor/dao"
	"github.com/xeniumd-china/xeniumd-monitor/dao/model"
	"github.com/xeniumd-china/xeniumd-monitor/leader"
	"time"
)

const (
	UNKNOWN_TYPE = iota
	ONCE
	PERIOD
)

const (
	UNKNOWN_UNIT = iota
	HOUR
	DAY
	WEEK
)

var taskCreators = make(map[string]TaskCreator)

func calcNextTime(thisTime *time.Time, schedule *model.Schedule) *time.Time {
	workTime := *common.ParseTime(schedule.TimeAt)
	switch schedule.Unit {
	case HOUR:
		workTime = time.Date(thisTime.Year(), thisTime.Month(), thisTime.Day(), thisTime.Hour(), workTime.Minute(), workTime.Second(), 0, time.Local)
		workTime = workTime.Add(time.Hour)
	case DAY:
		workTime = time.Date(thisTime.Year(), thisTime.Month(), thisTime.Day(), workTime.Hour(), workTime.Minute(), workTime.Second(), 0, time.Local)
		workTime = workTime.Add(24 * time.Hour)
	case WEEK:
		workTime = time.Date(thisTime.Year(), thisTime.Month(), thisTime.Day(), workTime.Hour(), workTime.Minute(), workTime.Second(), 0, time.Local)
		workTime = workTime.Add(7 * 24 * time.Hour)
	default:
	}
	return &workTime
}

type TaskCreator interface {
	CreateTasks(context *TaskContext, schedule *model.Schedule) []*model.Task

	GetType() string
}

func scheduleTasks(leaderSelector leader.LeaderSelector, parameters map[string]interface{}) {
	logger.Debug("Schedule create tasks")
	if leaderSelector == nil {
		return
	}

	if !leaderSelector.IsLeader() {
		return
	}

	// 找出时间点已到，应该执行的调度任务
	schedules, err := dao.FindAllSchedules()
	if err != nil {
		logger.Error(err)
		return
	}

	// 对每一个应该执行的任务，生成task
	for _, schedule := range schedules {
		context := NewContext(parameters)
		url, err := Parse(schedule.Url)
		if err != nil {
			logger.Error(err)
			continue
		}

		for key, value := range url.Parameters {
			context.SetString(key, value)
		}
		taskCreator := taskCreators[url.Protocol]
		if taskCreator != nil {
			tasks := taskCreator.CreateTasks(context, schedule)
			err = dao.InsertTask(tasks)
			if err != nil {
				logger.Error(err)
				continue
			}
			// 更新schedule表的next_time时间
			if schedule.ScheduleType == ONCE {
				// 一次性的任务不删除，只是设置为无效
				schedule.Status = model.DISABLED
				_, err = dao.UpdateSchedule(schedule)
				if err != nil {
					logger.Error(err)
					continue
				}
			} else if schedule.ScheduleType == PERIOD {
				// 周期性的任务
				schedule.NextTime = calcNextTime(common.ParseTime(schedule.NextTime), schedule).Format(common.FORMAT_SECOND)
				_, err = dao.UpdateSchedule(schedule)
				if err != nil {
					logger.Error(err)
					continue
				}
			} else {
				logger.Error("schedule type error:id=%d,type=%s.", schedule.Id, schedule.ScheduleType)
			}
		} else {
			logger.Error("Can not find creator of %s", schedule.Url)
		}
	}
}

func StartSchedule(leaderSelector leader.LeaderSelector, parameters map[string]interface{}, interval int) {
	go func() {
		timer := time.NewTicker(time.Duration(interval) * time.Second)
		for {
			<-timer.C
			go scheduleTasks(leaderSelector, parameters)
		}
	}()
}

func RegistryCreator(name string, creator TaskCreator) {
	taskCreators[name] = creator
}
