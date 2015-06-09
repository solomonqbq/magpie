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
	"strconv"
	"time"
)

type DataSource_KEY string

var dataSource *sql.DB

const (
	DEFAULT_DS = "magpie"
)

const (
	DELETE_TIME_OUT_WORKER_BY_ID = "DELETE FROM `mp_worker` WHERE `id` = ?"

	INSERT_WORKER = "INSERT INTO `mp_worker`(`name`, `created_time`, `time_out`,`group`) VALUES (?,?,DATE_ADD(now(),INTERVAL ? SECOND),?)"

	QUERY_ACTIVE_WORKERS   = "SELECT `id` FROM `mp_worker` WHERE `time_out`>= now() and `group`=?"
	QUERY_ALL_GROUP        = "SELECT `group` FROM `mp_group`"
	QUERY_TASKS            = "SELECT `id`, `name`, `group`, `worker_id`, `status`,`run_type`,`interval`,`context` FROM `mp_task` WHERE `group` = '%s' and `status` in (%s)"
	QUERY_NEW_FAILED_TASKS = "SELECT `id`, `name`, `group`, `worker_id`, `status`,`run_type`,`interval`,`context` FROM `mp_task` WHERE `group` = '%s' and `status` in (%s) and DATE_ADD(`updated_time`,INTERVAL `interval` SECOND) <=now()"
	QUERY_TIME_OUT_WORKER  = "SELECT `id` FROM `mp_worker` WHERE `time_out` < now()"
	QUERY_DISPATCHED_TASKS = "SELECT `id`, `name`, `group`, `worker_id`, `status`,`run_type`,`interval`,`context` FROM `mp_task` WHERE `status`=1 and `worker_id`=?"
	QUERY_ACTIVE_TASKS     = "SELECT count(`id`),`worker_id` from `mp_task` where `status` = 1 or `status` = 2 group by `worker_id`"

	UPDATE_WORKER_GROUP            = "UPDATE `mp_worker_group` SET `worker_id` = ? , `time_out`= DATE_ADD(now(),INTERVAL ? SECOND) WHERE (`group`=? and `worker_id`=?) or (`time_out` < now())"
	UPDATE_WORKER_TIME_OUT         = "UPDATE `mp_worker` SET `time_out`=DATE_ADD(now(),INTERVAL ? SECOND) WHERE id = ?"
	UPDATE_TASK_OWNER              = "UPDATE `mp_task` set `status` = 1,`worker_id`=%d where `id` in (%s)"
	UPDATE_TASK_STATUS_BY_WORKER   = "UPDATE `mp_task` set `status` = ? where `worker_id` = ? and (`status`<>4 and `status`<>5)"
	UPDATE_TASK_STATUS_BY_ID       = "UPDATE `mp_task` set `status` = ? where `id` = ?"
	UPDATE_TASK_STATUS_ERROR_BY_ID = "UPDATE `mp_task` set `status` = ?,`exception` = ? where `id` = ?"
)

func UpdateTaskStatusErr(id int64, status int, exp error) error {
	var err error
	if exp == nil {
		_, err = dataSource.Exec(UPDATE_TASK_STATUS_BY_ID, status, id)
	} else {
		_, err = dataSource.Exec(UPDATE_TASK_STATUS_ERROR_BY_ID, status, exp.Error(), id)
	}
	return err
}

func UpdateTaskStatus(id int64, status int) error {
	_, err := dataSource.Exec(UPDATE_TASK_STATUS_BY_ID, status, id)
	return err
}

func QueryDispatchedTasksByWorker(worker_id string) (tasks []*model.Mp_task, err error) {
	rows, err := dataSource.Query(QUERY_DISPATCHED_TASKS, worker_id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		t := new(model.Mp_task)
		err = rows.Scan(&t.Id, &t.Name, &t.Group, &t.Worker_id, &t.Status, &t.Run_type, &t.Interval, &t.Context)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}
	return
}

func DispathTask(wts []*WorkerTask) error {
	tx, err := dataSource.Begin()
	if err != nil {
		return err
	}
	for _, wt := range wts {
		if len(wt.new_task_id) == 0 {
			continue
		}
		var task_id_param string = ""
		for i, tid := range wt.new_task_id {
			if i != 0 {
				task_id_param = task_id_param + ","
			}
			task_id_param = task_id_param + strconv.FormatInt(tid, 10)
		}
		//		fmt.Println(fmt.Sprintf(UPDATE_TASK_OWNER, wt.id, task_id_param))
		result, err := tx.Exec(fmt.Sprintf(UPDATE_TASK_OWNER, wt.id, task_id_param))
		if err != nil {
			tx.Rollback()
			return err
		}
		affected, err := result.RowsAffected()
		if err != nil {
			return err
		}
		if affected != int64(len(wt.new_task_id)) {
			tx.Rollback()
			return errors.New(fmt.Sprintf("任务分配出错！期待分配给worker_id：%d的任务数是%d,实际分配了%d条，回滚...", wt.id, len(wt.new_task_id), affected))
		}
	}
	tx.Commit()
	return nil
}

func QueryActiveTasks() (workerIds []int64, taskCount []int64, err error) {
	rows, err := dataSource.Query(QUERY_ACTIVE_TASKS)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()
	workerIds = make([]int64, 0)
	taskCount = make([]int64, 0)
	for rows.Next() {
		var count int64
		var workerId int64
		err = rows.Scan(&count, &workerId)
		if err != nil {
			return nil, nil, err
		}
		workerIds = append(workerIds, workerId)
		taskCount = append(taskCount, count)
	}
	return

}

func InsertWorker(name string, time_out_interval time.Duration, group string) (id int64, err error) {
	result, err := dataSource.Exec(INSERT_WORKER, name, global.NowStr(), time_out_interval/time.Second, group)
	if err != nil {
		return -1, err
	}
	id, err = result.LastInsertId()
	if err != nil {
		return -1, err
	}
	return
}

func UpdateWorkerGroup(group string, worker_id int64, time_out_interval time.Duration) (affact int64, err error) {
	result, err := dataSource.Exec(UPDATE_WORKER_GROUP, worker_id, time_out_interval/time.Second, group, worker_id)
	if err != nil {
		return 0, err
	}
	affact, err = result.RowsAffected()
	return
}

func UpdateWorkerTimeout(id int64, interval time.Duration) (affected int64, err error) {
	result, err := dataSource.Exec(UPDATE_WORKER_TIME_OUT, interval/time.Second, id)
	if err != nil {
		return 0, err
	}
	affected, err = result.RowsAffected()
	return
}

func QueryTimeoutWorker() (workersId []int64, err error) {
	workersId = make([]int64, 0)
	rows, err := dataSource.Query(QUERY_TIME_OUT_WORKER)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var id int64
		rows.Scan(&id)
		if err != nil {
			return
		}
		workersId = append(workersId, id)
	}
	if len(workersId) == 0 {
		return
	}
	return
}

func DeleteTimeoutWorker(workersId []int64) (affected int64, err error) {
	tx, err := dataSource.Begin()
	if err != nil {
		return 0, err
	}

	for _, wid := range workersId {
		_, err := tx.Exec(UPDATE_TASK_STATUS_BY_WORKER, core.TASK_FAIL, wid)
		if err != nil {
			tx.Rollback()
			return 0, err
		}
		result, err := tx.Exec(DELETE_TIME_OUT_WORKER_BY_ID, wid)
		if err != nil {
			tx.Rollback()
			return 0, err
		}
		a, err := result.RowsAffected()
		if err != nil {
			tx.Rollback()
			return 0, err
		}
		affected = affected + a
	}

	tx.Commit()
	return
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
	var status_param string = ""
	for i, s := range status {
		if i != 0 {
			status_param = status_param + ","
		}
		status_param = status_param + strconv.Itoa(s)
	}
	rows, err := dataSource.Query(fmt.Sprintf(QUERY_NEW_FAILED_TASKS, group, status_param))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	tasks = make([]*model.Mp_task, 0)
	for rows.Next() {
		t := new(model.Mp_task)
		err = rows.Scan(&t.Id, &t.Name, &t.Group, &t.Worker_id, &t.Status, &t.Run_type, &t.Interval, &t.Context)
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
func QueryActiveWorkers(group string) (workerIds []string, err error) {
	rows, err := dataSource.Query(QUERY_ACTIVE_WORKERS, group)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	workerIds = make([]string, 0)
	for rows.Next() {
		var Id string
		err = rows.Scan(&Id)
		if err != nil {
			return nil, err
		}
		workerIds = append(workerIds, Id)
	}
	return
}
