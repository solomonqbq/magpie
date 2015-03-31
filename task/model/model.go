package model

type BasicModel struct {
	Id         uint64 `json:"id"`
	CreateTime string `json:"createTime"`
	UpdateTime string `json:"-"`
	Status     int8   `json:"-"`
}

type Config struct {
	Id       uint64
	Group    string
	Key      string
	Value    string
	Password bool
	Status   int8
}

type Task struct {
	// 类型
	Code string
	//是否为后台任务
	Daemon bool
	// 优先级
	Priority uint8
	//权重
	Weight uint8
	// 所有者
	Owner string
	// 参数
	Url string
	// 重试次数
	Retry uint16
	// 重试时间
	RetryTime string
	// 异常
	Exception string
}

func (this *Task) String() string {
	return fmt.Sprintf("[%d %s]", this.Id, this.Url)
}

type Schedule struct {
	BasicModel
	Group        string
	ScheduleType uint8
	Unit         uint8
	Period       uint32
	Url          string
	StartDate    string
	EndDate      string
	Cron         string
	TimeAt       string
	NextTime     string
}
type TaskNode struct {
	Id     uint64
	Ip     string
	Port   uint16
	Name   string
	Size   int
	Type   int8
	Group  string
	Status int8
}

func (this *TaskNode) String() string {
	return this.Name
}

type Chart struct {
	BasicModel
	ServerId    string         `json:"serverId"`
	ContainerId string         `json:"containerId"`
	Name        string         `json:"name"`
	Code        string         `json:"code"`
	Lines       []*MetricsType `json:"lines"`
}

type AlarmConfig struct {
	BasicModel
	Pin       string          `json:"pin"`
	Name      string          `json:"name"`
	Detail    string          `json:"detail"`
	Servers   []*Server       `json:"servers,omitempty"`
	Rules     []*AlarmRule    `json:"rules,omitempty"`
	Contracts []*ContractUser `json:"contracts,omitempty"`
}

type Server struct {
	Id          uint64
	AlarmConfig *AlarmConfig `json:"config,omitempty"`
	ServerId    string       `json:"serverId"`
	ContainerId string       `json:"containerId"`
}

type AlarmRule struct {
	BasicModel
	AlarmConfig *AlarmConfig `json:"config,omitempty"`
	ConfigId    uint64       `json:"-"`
	MetricsCode string       `json:"code"`
	Period      string       `json:"period"`
	Level       uint8        `json:"-"`
	Threshold   float64      `json:"threshold"`
	Interval    uint32       `json:"-"`
	DetectTimes uint16       `json:"-"`
}

type ContractUser struct {
	BasicModel
	Pin       string `json:"-"`
	Name      string `json:"name"`
	Key       string `json:"-"`
	Telephone string `json:"telephone"`
	Email     string `json:"email"`
}

type AlarmEvent struct {
	Id          uint64  `json:"id"`
	MetricsCode string  `json:"code"`
	ServerId    string  `json:"serverId"`
	ContainerId string  `json:"containerId"`
	ThreshHold  float32 `json:"threshold"`
	Value       float32 `json:"value"`
	AlarmTime   string  `json:"alarmTime"`
}

type ClientMetrics struct {
	ServerId    string               `json:"host_id"`
	ContainerId string               `json:"container_id"`
	StartTime   int64                `json:"start_time"`
	EndTime     int64                `json:"end_time"`
	Data        map[string][]float64 `json:"data"`
}

type AlarmRuleServersView struct {
	AlarmRule
	Server
}
