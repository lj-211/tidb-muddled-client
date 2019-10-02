package db_coordinate

// -------------------------------------------------------------------------
// model
const (
	CiStatus_Regist    = iota // 批次节点刚注册
	CiStatus_InitOk           // 批次节点初始化成功(已压入任务完成)
	CiStatus_Completed        // 任务可以启动的状态
)

type BatchLock struct {
	ID      uint `gorm:"primary_key"`
	BatchId string
}

func (e *BatchLock) TableName() string {
	return "batch_lock"
}

type CoordinateInfo struct {
	ID          uint   `gorm:"primary_key"`
	BatchId     string // 批次id
	NodeId      string // 节点id
	TaskCnt     int    // 任务数量
	DoneTaskCnt int    // 已执行任务数量
	Status      int    // 状态
	Ttl         int64  // 心跳时间
}

func (e *CoordinateInfo) TableName() string {
	return "coordinate_info"
}

type CmdInfo struct {
	ID      uint   `gorm:"primary_key"`
	BatchId string // 批次id
	NodeId  string // 命令所属的节点id
	Sql     string // 命令
}

func (e *CmdInfo) TableName() string {
	return "cmd_info"
}

type SimpleCmdInfo struct {
	ID     uint   `gorm:"primary_key"`
	NodeId string // 命令所属的节点id
}

func (e *SimpleCmdInfo) TableName() string {
	return "cmd_info"
}

type CmdOrder struct {
	ID      uint   `gorm:"primary_key"`
	BatchId string // 批次id
	CmdId   uint   // 命令id
	NodeId  string // 节点id
	IsDone  int    // 是否完成
}

func (e *CmdOrder) TableName() string {
	return "cmd_order"
}
