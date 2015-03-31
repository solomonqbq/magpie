package model


type Mp_board_group struct{
	Id	int64	//自增主键
	Group	string	//组名
	Board_id	string	//当前组的leader看板
	Time_out	string	//到期有效时间
	Created_time	string	//创建时间
	Updated_time	string	//最后更新时间
}


type Mp_group struct{
	Id	int32	//自增主键
	Group	string	//组名
	Created_time	string	//创建时间
	Updated_time	string	//更新时间
}


type Mp_member struct{
	Id	int64	//自增ID
	Ip	string	//节点所在IP
	Group	string	//所在组
	Time_out	string	//到期有效时间，这个时间超过当前系统时间即为该member无效
	Created_time	string	//创建时间
	Updated_time	string	//最后更新时间
}


type Mp_task struct{
	Id	int64	//自增ID
	Name	string	//任务名
	Context	string	//任务上下文参数
	Group	string	//组名
	Mem_id	int64	//所有者的组员ID
	Retry	int32	//重试次数
	Run_type	int32	//任务类型 0:周期任务 1:一次性任务
	Interval	int32	//运行间隔，只有run_type是周期任务时才有效
	Exception	string	//最后一次错误
	Created_time	string	//创建时间
	Updated_tIme	string	//最后更新时间
	Status	int32	//0:新任务 1:已分配 2:运行中 3::失败 4:成功 5:错误
}


type Mp_board struct{
	Id	string	//主键
	Time_out	string	//到期有效时间
	Created_time	string	//创建时间
	Updated_time	string	//最后更新时间
}


