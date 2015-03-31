package db

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/xeniumd-china/magpie/core"
	"github.com/xeniumd-china/magpie/global"
	"log"
	"strconv"
	"strings"
	"time"
)

func NewDBBoard() *core.Board {
	b := core.NewBoard()

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
		id, err := InsertBoard(name, time.Duration(global.Properties.Int("board.timeout.interval", 10))*time.Second)
		if err != nil {
			return err
		}
		b.Id = strconv.Itoa(id)
		return nil
	}
	b.HeartBeat = func() {
		if b.Id != 0 {
			id, _ := strconv.Atoi(b.Id)
			UpdateBoardTimeout(int64(id), time.Duration(global.Properties.Int("board.timeout.interval", 10))*time.Second)
		}

	}

	b.Cleanup = func(group string) {
		//清理board
		DeleteTimeoutBoard()
		//清理member
		DeleteTimeoutMember()
	}

	b.LoadAllGroup = func() (groups []string, err error) {
		return QueryAllGroup()
	}

	b.LoadActiveMembers = func(group string) (mems []*core.MemID, err error) {
		mp_members, err := QueryActiveMembers(group, time.Duration(global.Properties.Int("member.timeout.interval", 10))*time.Second)
		if err != nil {
			return nil, err
		}
		mems = make([]*core.MemID, 0)
		for _, mp_mem := range mp_members {
			mem := new(core.MemID)
			mem.Id = strconv.FormatInt(mp_mem.Id, 10)
			mem.Group = mp_mem.Group
			mems = append(mems, mem)
		}
		return mems, nil
	}

	b.LoadTasks = func(group string) (tasks []*core.Task, err error) {
		mp_tasks, err := QueryNewAndFailedTasks()
		if err != nil {
			return nil, err
		}
		tasks := make([]*core.Task, 0)
		for _, mp_t := range mp_tasks {
			t := new(core.Task)
			t.ID = mp_t.Id
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
			t.Running_type = mp_t.Run_type
			t.Interval = mp_t.Interval * time.Second
			tasks = append(tasks, t)
		}
		return
	}
	b.SelectLeader = func(group string) bool {
		affact, err := UpdateBoardGroup(group, b.Id, time.Duration(global.Properties.Int("board.timeout.interval", 10))*time.Second)
		if err != nil {
			return false
		} else {
			return affact >= 1
		}
	}

	b.DispatchTasks = func(mems []*core.MemID, tasks []*core.Task) {
		if mems==nil||le
		//根据任务数均分
	}
	return b
}

func NewDBMember(b *core.Board, group string) *core.Member {
	m := core.NewMember(b, group)
	m.HeartBeat_interval = 3 * time.Second
	m.ReportTasksStaus = func() error {
		log.Printf("%s上报任务状态", m.Id)
		return nil
	}
	m.HeartBeat = func() {
		log.Printf("%s发送心跳", m.Id)
	}
	return m
}
