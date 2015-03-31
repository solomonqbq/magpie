package db

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/dmotylev/goproperties"
	_ "github.com/go-sql-driver/mysql"
	"github.com/xeniumd-china/magpie/core"
	"github.com/xeniumd-china/magpie/db/model"
	"github.com/xeniumd-china/magpie/global"
	"time"
)

type DataSource_KEY string

var dataSource *sql.DB

const (
	DEFAULT_DS = "magpie"
)

const (
	QUERY_ACTIVE_MEMBERS         = "SELECT `id`, `group` FROM `mp_member` WHERE `group`=? and `time_out`>= now()"
	QUERY_ALL_GROUP              = "SELECT `group` FROM `mp_group`"
	QUERY_TASKS                  = "SELECT `id`, `name`, `group`, `mem_id`, `status`,`run_type`,`interval` FROM `mp_task` WHERE `group` = ? and status in (?)"
	DELETE_TIME_OUT_BOARD        = "DELETE FROM `mp_board` WHERE `time_out` < now()"
	DELETE_TIME_OUT_MEMBER       = "DELETE FROM `mp_member` WHERE id in (?)"
	QUERY_TIME_OUT_MEMBER        = "SELECT `id` from `mp_member` where `time_out` < now()"
	UPDATE_TASK_STATUS_BY_MEM_ID = "UPDATE `mp_task` set `status`= ? where mem_id = ?"
	UPDATE_BOARD_TIME_OUT        = "UPDATE `mp_board` SET `time_out`=DATE_ADD(now(),INTERVAL ? SECOND) WHERE id = ?"
	INSERT_BOARD                 = "INSERT INTO `mp_board`(`name`, `created_time`, `time_out`) VALUES (?,?,DATE_ADD(now(),INTERVAL ? SECOND))"
	UPDATE_BORAD_GROUP           = "UPDATE `mp_board_group` SET group = ? and board_id = ? and `time_out`= DATE_ADD(now(),INTERVAL ? SECOND)) WHERE (group=? and board_id=?) or (`time_out` < now())"
)

func InsertBoard(name string, time_out_interval time.Duration) (id int64, err error) {
	result, err := dataSource.Exec(INSERT_BOARD, global.NowStr(), time_out_interval.Second())
	if err != nil {
		return -1, err
	}
	id, err = result.LastInsertId()
	if err != nil {
		return -1, err
	}
	return
}

func UpdateBoardGroup(group string, board_id int64, time_out_interval time.Duartion) (affact int64, err error) {
	result, err := dataSource.Exec(UPDATE_BORAD_GROUP, group, board_id, time_out_interval.Second(), group, board_id)
	if err != nil {
		return 0, err
	}
	affact, err = result.RowsAffected()
	return
}

func UpdateBoardTimeout(id int64, interval time.Duration) error {
	_, err = dataSource.Exec(UPDATE_BOARD_TIME_OUT, interval.Second(), id)
	return
}

func DeleteTimeoutBoard() error {
	_, err := dataSource.Exec(DELETE_TIME_OUT_BOARD)
	return err
}

func DeleteTimeoutMember() error {
	rows, err := dataSource.Query(QUERY_TIME_OUT_MEMBER)
	if err != nil {
		return err
	}
	defer rows.Close()
	ids := make([]uint64, 0)
	for rows.Next() {
		var id uint64
		err = rows.Scan(&id)
		if err != nil {
			return err
		}
		ids = append(ids, id)
	}
	if len(ids) == 0 {
		return nil
	}
	//将已死组员的任务变为可分配
	for _, id := range ids {
		dataSource.Exec(UPDATE_TASK_STATUS_BY_MEM_ID, core.TASK_FAIL, id)
	}

	_, err = dataSource.Exec(DELETE_TIME_OUT_MEMBER, ids)
	return err
}

func InitAllDS(prop properties.Properties) {
	dsKeys := []string{DEFAULT_DS}
	var db_url_template = "%s:%s@tcp(%s:%s)/%s?charset=%s"
	for _, key := range dsKeys {
		account := prop.String(key+".db.account", "account")
		password := prop.String(key+".db.password", "password")
		ip := prop.String(key+".db.ip", "127.0.0.1")
		port := prop.String(key+".db.port", "3306")
		schema := prop.String(key+".db.schema", "")
		charset := prop.String(key+".db.charset", "utf8")
		maxIdle := int(prop.Int(key+".db.maxIdle", 25))
		maxOpen := int(prop.Int(key+".db.maxOpen", 50))
		db_url := fmt.Sprintf(db_url_template, account, password, ip, port, schema, charset)
		InitDS(DataSource_KEY(key), db_url, maxIdle, maxOpen)
	}
}

func InitDS(dsKey DataSource_KEY, url string, maxIdle int, maxOpen int) (err error) {
	db, err := sql.Open("mysql", url)
	if maxIdle > 0 {
		db.SetMaxIdleConns(maxIdle)
	} else {
		db.SetMaxIdleConns(25)
	}
	if maxOpen > 0 {
		db.SetMaxOpenConns(maxOpen)
	} else {
		db.SetMaxOpenConns(50)
	}

	if err != nil {
		return
	}
	err = db.Ping()

	if err != nil {
		return
	}

	switch dsKey {
	case DEFAULT_DS:
		dataSource = db
	default:
		return errors.New("unkown dsKsy:" + string(dsKey))
	}
	return nil
}

func queryTasksByStatus(group string, status []int) (tasks []*model.Mp_task, err error) {
	rows, err := dataSource.Query(QUERY_TASKS, group, status)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	tasks = make([]*model.Mp_task, 0)
	for rows.Next() {
		t := new(model.Mp_task)
		err = rows.Scan(&t.Id, &t.Name, &t.Group, &t.Mem_id, &t.Status, &t.Run_type, &t.Interval)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}
	return tasks, err
}

func QueryNewAndFailedTasks(group string) (tasks []*model.Mp_task, err error) {
	s := []int{core.TASK_NEW, core.TASK_FAIL}
	return queryTasksByStatus(group, s)
}

//查询所有组
func QueryAllGroup() (groups []string, err error) {
	rows, err := dataSource.Query(QUERY_ALL_GROUP)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	groups = make([]string, 0)
	for rows.Next() {
		var g string
		err = rows.Scan(&g)
		if err != nil {
			return nil, err
		}
		groups = append(groups, g)
	}
	return
}

//查询存活的组员
func QueryActiveMembers(group string) (mp_members []*model.Mp_member, err error) {
	result, err := dataSource.Query(QUERY_ACTIVE_MEMBERS, group)
	if err != nil {
		return nil, err
	}
	defer result.Close()
	mp_members = make([]*model.Mp_member, 0)
	for result.Next() {
		m := new(model.Mp_member)
		err = result.Scan(&m.Id, &m.Group)
		if err != nil {
			return nil, err
		}
		mp_members = append(mp_members, m)
	}
	return
}

//
////插入负载均衡实例
//func InsertInstance(instance *model.Instance) error {
//	stmt, err := dataSource.Prepare(INSERT_INSTANCE)
//	if err != nil {
//		return err
//	}
//	defer stmt.Close()
//	if instance.Id == "" {
//		return errors.New("负载均衡实例Id不能为空")
//	}
//	if instance.Name == "" {
//		instance.Name = instance.Id
//	}
//	if instance.User_id == "" {
//		return errors.New("用户ID不能为空！")
//	}
//	if instance.Vip == "" {
//		return errors.New("VIP不能为空！")
//	}
//
//	instance.Status = 1
//	if instance.Max_conns == 0 {
//		instance.Max_conns = 5000
//	}
//
//	t := time.Now()
//	instance.Created_time = t.Format(global.FORMAT_SECOND)
//	instance.Updated_time = instance.Created_time
//	instance.Partition_id = 1
//	result, err := stmt.Exec(instance.Id, instance.Name, instance.User_id, instance.Vip, instance.Type, instance.Status, instance.Max_conns, instance.Created_time, instance.Updated_time, instance.Partition_id)
//	if err != nil {
//		return err
//	}
//	affected, err := result.RowsAffected()
//	if err != nil {
//		return err
//	}
//	if affected != 1 {
//		return errors.New("未成功插入！")
//	}
//	InsertChangeLogByInstanceId(instance.Id)
//	return err
//}
