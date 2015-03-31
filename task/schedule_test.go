package task

import (
	"fmt"
	"github.com/xeniumd-china/xeniumd-monitor/dao"
	"github.com/xeniumd-china/xeniumd-monitor/leader"
	"testing"
)

func TestSchedule(t *testing.T) {
	err := dao.InitDataSource("eagle:eagle@tcp(192.168.15.64:3306)/eagle?charset=utf8")
	if err != nil {
		fmt.Println(err)
	}

	connection, _ := dao.GetDataSource()
	leaderSelector, _ := leader.NewDbLeaderSelector(connection, "task.dispatcher", "test1")
	leaderSelector.Start()

	scheduleTasks(leaderSelector, make(map[string]interface{}))
}
