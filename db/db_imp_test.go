package db

import (
	"github.com/dmotylev/goproperties"
	"github.com/xeniumd-china/magpie/core"
	"github.com/xeniumd-china/magpie/global"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func initGroup() {
	current_path, _ := os.Getwd()
	properties_file := filepath.Join(current_path, "../magpie.properties")
	global.Properties, _ = properties.Load(properties_file)
	InitAllDS(global.Properties)
	dataSource.Exec("truncate table `mp_group`")
	dataSource.Exec("truncate table `mp_worker_greoup`")
	dataSource.Exec("INSERT INTO `mp_group`(`group`, `created_time`, `updated_time`) VALUES ('load_balance',now(),now())")
	dataSource.Exec("INSERT INTO `mp_worker_group`(`group`, `worker_id`, `time_out`, `created_time`) VALUES ('load_balance',-1,now(),now())")
}

func TestDBBoard(t *testing.T) {
	initGroup()
	b := core.NewWorkerExecutor(NewDBWorker("load_balance"))

	b.Start()
	//	b.Init()
	//	groups, err := b.LoadAllGroup()
	//	fmt.Println(groups)
	//	mems, err := b.LoadActiveMembers("load_balance")
	//	if err != nil {
	//		fmt.Println(err)
	//	}
	//	fmt.Println(mems)
	//
	//	mems, err := b.LoadActiveMembers("load_balance")
	//	if err != nil {
	//		fmt.Println(err)
	//	}
	//	fmt.Println(mems)

	//	b.Cleanup("load_balance")
	time.Sleep(1000 * time.Second)
}
