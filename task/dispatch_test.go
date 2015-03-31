package task

import (
	"fmt"
	"github.com/xeniumd-china/xeniumd-monitor/dao"
	"github.com/xeniumd-china/xeniumd-monitor/leader"
	"testing"
	"time"
)

func TestDispatch(t *testing.T) {
	err := dao.InitDataSource("root:@tcp(127.0.0.1:3306)/eagle?charset=utf8")
	if err != nil {
		fmt.Println(err)
	}
	connection, _ := dao.GetDataSource()
	leaderSelector, _ := leader.NewDbLeaderSelector(connection, "task.dispatcher", "test_node1")
	leaderSelector.Start()

	registry := NewDbTaskDispatchRegistry(leaderSelector)
	registry.Start("test_node1")

	dispatcher := NewTaskDispatcher(registry)
	dispatcher.Start()
	for {
		time.Sleep(time.Second)
	}
}
