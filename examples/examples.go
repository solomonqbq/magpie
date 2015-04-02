package main

import (
	"errors"
	"github.com/xeniumd-china/magpie/core"
	"github.com/xeniumd-china/magpie/db"
	"github.com/xeniumd-china/magpie/global"
	"math/rand"
	"time"
)

type MockTaskExecutor struct {
}

func NewMockTaskExecutor(taskName string) *core.Task_Executor {
	te := new(core.Task_Executor)
	var tn = taskName
	te.Init = func() error {
		return nil
	}
	te.Execute = func() *core.Result {
		i := rand.Intn(100)
		r := new(core.Result)
		r.Task = te.Task
		if i <= 49 {
			r.Result_code = core.TASK_SUCCESS
		} else {
			r.Result_code = core.TASK_FAIL
			r.Error = errors.New("mock error")
		}
		time.Sleep(3 * time.Second)
		return r
	}
	te.Stop = func() error {
		return nil
	}
	te.GetTaskName = func() string {
		return tn
	}
	return te
}

func main() {
	confs := []string{"magpie.properties", "src/github.com/xeniumd-china/magpie/magpie.properties"}
	conf := global.FindConf(confs)
	global.Load(conf)
	tns := []string{"deliver_conf", "test", "test2", "test3", "test4", "test5"}
	for _, tn := range tns {
		core.Registry(tn, NewMockTaskExecutor(tn))
	}

	w := db.NewDBWorker()
	w.Start()
	time.Sleep(1000 * time.Second)

}
