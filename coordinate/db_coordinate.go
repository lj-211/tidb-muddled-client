package coordinate

import (
	"context"
	"fmt"
	"log"
	"sort"
	"time"

	"github.com/jinzhu/gorm"
	cdb "github.com/lj-211/common/db"
	"github.com/pkg/errors"

	"github.com/lj-211/tidb-muddled-client/algorithm"
	"github.com/lj-211/tidb-muddled-client/common"
)

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

const ttlTimeDuration time.Duration = time.Second * 10
const nodeExpireTime int64 = 60 // 单位秒

// -------------------------------------------------------------------------
// coordinater
// NOTE:
//	1. db协调依赖于mysql,不可在tidb中实现,因为初始化任务序列依赖于锁抢占
//		而tidb是在提交时，才检测锁冲突
//	2. 因为脚本可能会被调度，所以调度本身不考虑设计任务超时退出的逻辑
type DbCoordinater struct {
	Db       *gorm.DB
	Id       string
	BatchId  string
	Partners []string
	Ctx      context.Context
	Cancel   context.CancelFunc
	Done     chan TaskRst
	InitOk   chan bool
	Proc     TaskProcesser
}

// 创建Db协调者
func NewDbCoordinate(id, bid string, partners []string, db *gorm.DB) (*DbCoordinater, error) {
	if db == nil {
		return nil, common.NilInputErr
	}
	if len(partners) == 0 {
		return nil, common.ParamInvalidErr
	}

	return &DbCoordinater{
		Db:       db,
		BatchId:  bid,
		Partners: partners,
		Id:       id,
	}, nil
}

// 启动协调器
func (this *DbCoordinater) Start(ctx context.Context, proc TaskProcesser) error {
	this.Proc = proc
	return this.doStart(this.BatchId, this.Partners)
}

// 向Db协调器注册任务
func (this *DbCoordinater) PushTask(ctx context.Context, data interface{}, done bool) error {
	if data == nil {
		return common.NilInputErr
	}
	cdata, ok := data.(CmdInfo)
	if !ok {
		return errors.New("推送数据异常")
	}
	switch {
	case done == true:
		// set ci status to InitOk
		err := this.Db.Model(&CoordinateInfo{}).Where("batch_id = ? and node_id = ?", cdata.BatchId, this.Id).
			Update("status", CiStatus_InitOk).Error
		if err != nil {
			return errors.Wrap(err, "设置初始化完成失败")
		}
		log.Println("节点", this.Id, "加载完毕", cdata.BatchId, this.Id)
	case done == false:
		trans := func(db *gorm.DB) error {
			// add task
			err := db.Model(&cdata).Create(&cdata).Error
			if err != nil {
				return errors.Wrap(err, "压入任务失败")
			}

			err = db.Model(&CoordinateInfo{}).
				Where("batch_id = ? and node_id = ?", this.BatchId, this.Id).
				Update("task_cnt", gorm.Expr("task_cnt + 1")).Error
			if err != nil {
				return errors.Wrap(err, "更新任务数量失败")
			}

			return nil
		}
		if err := cdb.DoTrans(this.Db, trans); err != nil {
			return err
		}
		log.Println("push sql: ", cdata.Sql)
	}

	return nil
}

// 监听协调器接收调度并且执行任务
func (this *DbCoordinater) Watch(ctx context.Context) error {
WAIT_OK_LOOP:
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-this.InitOk:
			break WAIT_OK_LOOP
		}
	}

	log.Println("开始消费任务")

	tk := time.NewTicker(time.Millisecond * 10)
	defer tk.Stop()
	// start watch task
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-tk.C:
			err := this.DoTask(ctx)
			if err != nil && !gorm.IsRecordNotFoundError(errors.Cause(err)) {
				log.Println("ERROR: 执行任务失败 ", err.Error())
			}
		}
	}

	log.Println("INFO: 任务监听协程退出")

	return nil
}

// 执行任务逻辑
func (this *DbCoordinater) DoTask(ctx context.Context) error {
	co := &CmdOrder{}
	err := this.Db.Model(co).Where("batch_id = ? and is_done = 0", this.BatchId).
		Order("id asc").Limit(1).Find(co).Error
	if err != nil {
		return errors.Wrap(err, "query task fail")
	}

	if co.NodeId != this.Id {
		return nil
	}

	trans := func(db *gorm.DB) error {
		ci := &CmdInfo{}
		err := db.Model(ci).Where("id = ?", co.CmdId).Find(ci).Error
		if err != nil {
			return errors.Wrap(err, "查询任务失败")
		}
		sql := ci.Sql
		// TODO
		// 这里可能存在数据不一致的问题，待优化
		err = this.Proc(sql)
		if err != nil {
			return errors.Wrap(err, "执行任务失败")
		}
		err = db.Model(co).Where("id = ?", co.ID).Update("is_done", 1).Error
		if err != nil {
			return errors.Wrap(err, "提交任务失败")
		}

		err = db.Model(&CoordinateInfo{}).Where("batch_id = ? and node_id = ?", this.BatchId, this.Id).
			Update("done_task_cnt", gorm.Expr("done_task_cnt + 1")).Error
		if err != nil {
			return errors.Wrap(err, "更新已完成任务失败")
		}

		return nil
	}

	if err := cdb.DoTrans(this.Db, trans); err != nil {
		return err
	}

	return nil
}

// 阻塞获取完成状态
func (this *DbCoordinater) BlockCheckDone(ctx context.Context) TaskRst {
	rt := <-this.Done
	return rt
}

// 检查所有任务是否完成的状态
func (this *DbCoordinater) checkDone() (TaskRst, error) {
	doneState := DoneState_Unknown
	msg := ""

	// 伙伴节点的任务全部都完成
	allIds := this.getAllIds()

	cis := make([]*CoordinateInfo, 0)
	err := this.Db.Model(&CoordinateInfo{}).
		Where("batch_id = ? and node_id in (?)", this.BatchId, allIds).Find(&cis).Error
	if err != nil {
		return TaskRst{}, errors.Wrap(err, "查询是否完成失败")
	}
	allTask := 0
	doneTask := 0
	allAlive := true
	now := time.Now().Unix()
	for i := 0; i < len(cis); i++ {
		v := cis[i]
		allTask += v.TaskCnt
		doneTask += v.DoneTaskCnt
		if (now - v.Ttl) > nodeExpireTime {
			allAlive = false
		}
	}

	switch {
	case allTask == doneTask:
		doneState = DoneState_OK
	case allTask < doneTask:
		if !allAlive {
			doneState = DoneState_OverTime
			msg = "任务节点已超时, 任务退出"
		} else {
			doneState = DoneState_Doing
			msg = fmt.Sprintf("任务进度: %d / %d", doneTask, allTask)
		}
	default:
		doneState = DoneState_Unknown
		msg = "任务异常，请检查任务列表"
	}

	return TaskRst{
		DoneState: doneState,
		Msg:       msg,
	}, nil
}

// 等待初始化完成
func (this *DbCoordinater) WatchInitOk(ctx context.Context) {
	tk := time.NewTicker(time.Second * 2)
	defer tk.Stop()

	allIds := this.getAllIds()

	for {
		select {
		case <-ctx.Done():
			return
		case <-tk.C:
			// 1. do query if all node status
			cis := make([]*CoordinateInfo, 0)
			err := this.Db.Model(&CoordinateInfo{}).
				Where("batch_id = ? and node_id in (?)", this.BatchId, allIds).Find(&cis).Error
			if err != nil {
				break
			}

			if len(cis) != len(allIds) {
				break
			}

			// 1. 检查是否是全InitOk状态，是的话尝试去初始化任务序列表
			// 2. 因为1里面做了double check，所以可能已经初始化完成，跳出进行下一次判断，期待进入allCom的逻辑
			allOk := true
			for i := 0; i < len(cis); i++ {
				v := cis[i]
				if v.Status != CiStatus_InitOk {
					allOk = false
					break
				}
			}
			if allOk {
				err = this.genTaskOrder(ctx)
				if err != nil {
					log.Println("gen: ", err.Error())
				} else {
					log.Println("初始化任务次序成功")
				}
				break
			}

			allCom := true
			for i := 0; i < len(cis); i++ {
				v := cis[i]
				if v.Status != CiStatus_Completed {
					allCom = false
					break
				}
			}
			if allCom {
				log.Println("初始化成功")
				this.InitOk <- true
				return
			}
		}
	}
}

// 批次锁，用于多个client抢占执行初始化任务序列
func (this *DbCoordinater) lockBatch(db *gorm.DB, bid string) error {
	bi := &BatchLock{}
	return db.Set("gorm:query_option", "FOR UPDATE").Model(bi).
		Where("batch_id = ?", bid).Find(bi).Error
}

// 生成全排列任务序列
func (this *DbCoordinater) genTaskOrder(ctx context.Context) error {
	// TODO
	//	这里的查询以及下面的查询可能存在过多数据查询和大事务的问题，
	//	可以根据业务情况优化
	allIds := this.getAllIds()
	sort.Strings(allIds)

	numList := make([][]int, 0)
	idxMap := make(map[int]*CmdInfo)
	for i := 0; i < len(allIds); i++ {
		cmds := make([]*CmdInfo, 0)
		err := this.Db.Model(&CmdInfo{}).Order("id asc").
			Where("batch_id = ? and node_id = ?", this.BatchId, allIds[i]).Find(&cmds).Error
		if err != nil {
			return errors.Wrap(err, "查询任务列表失败")
		}

		nums := make([]int, 0)
		for _, v := range cmds {
			nums = append(nums, int(v.ID))
			idxMap[int(v.ID)] = v
		}

		numList = append(numList, nums)
	}

	orders := algorithm.FullListPermutation(numList)

	trans := func(db *gorm.DB) error {
		var err error

		err = this.lockBatch(db, this.BatchId)
		if err != nil {
			return errors.Wrap(err, "lock fail")
		}

		// double check
		ci := &CoordinateInfo{}
		err = db.Model(ci).
			Where("batch_id = ? and node_id = ?", this.BatchId, allIds[0]).Find(ci).Error
		if err != nil {
			return errors.Wrap(err, "double check status fail")
		}
		if ci.Status != CiStatus_InitOk {
			log.Println("状态异常，放弃生成序列数据")
			return nil
		}

	OUT_LOOP:
		for _, v := range orders {
			for _, e := range v {
				cmd, _ := idxMap[e]
				co := &CmdOrder{
					BatchId: this.BatchId,
					CmdId:   cmd.ID,
					NodeId:  cmd.NodeId,
				}

				err = db.Model(co).Create(co).Error
				if err != nil {
					break OUT_LOOP
				}
			}
		}

		err = this.Db.Model(&CoordinateInfo{}).Where("batch_id = ? and node_id in (?)", this.BatchId, allIds).
			Update(map[string]interface{}{
				"status":   CiStatus_Completed,
				"task_cnt": gorm.Expr("task_cnt * ?", len(orders))}).Error

		return err
	}

	if err := cdb.DoTrans(this.Db, trans); err != nil {
		return errors.Wrap(err, "执行生成任务失败")
	}

	return nil
}

// db协调启动逻辑
func (this *DbCoordinater) doStart(bid string, partners []string) error {
	this.Ctx, this.Cancel = context.WithCancel(context.Background())
	this.Done = make(chan TaskRst)
	this.InitOk = make(chan bool)

	trans := func(db *gorm.DB) error {
		ci := &CoordinateInfo{
			BatchId:     bid,
			NodeId:      this.Id,
			TaskCnt:     0,
			DoneTaskCnt: 0,
			Status:      CiStatus_Regist,
		}
		err := this.Db.Model(ci).Create(ci).Error
		if err != nil {
			return errors.Wrap(err, "launch coordinate fail")
		}

		err = this.Db.Exec(fmt.Sprintf("insert ignore into batch_lock (batch_id) values (%s)", bid)).Error
		if err != nil {
			return errors.Wrap(err, "create batch lock fail")
		}

		return nil
	}
	if err := cdb.DoTrans(this.Db, trans); err != nil {
		return err
	}

	// check task all push ok then start watcher
	wioCtx, _ := context.WithCancel(this.Ctx)
	go this.WatchInitOk(wioCtx)
	// task watcher
	wtCtx, _ := context.WithCancel(this.Ctx)
	go this.Watch(wtCtx)
	// watch done
	wdCtx, _ := context.WithCancel(this.Ctx)
	go this.watchDone(wdCtx)
	// ttl
	ttlCtx, _ := context.WithCancel(this.Ctx)
	go this.ttl(ttlCtx)

	return nil
}

// 内部用函数获取批次所有节点名称
func (this *DbCoordinater) getAllIds() []string {
	allIds := make([]string, 0)
	allIds = append(allIds, this.Id)
	allIds = append(allIds, this.Partners...)
	return allIds
}

// 实现检查任务是否完成的监听
func (this *DbCoordinater) watchDone(ctx context.Context) {
	tk := time.NewTicker(time.Second)
	defer tk.Stop()

OUT_LOOP:
	for {
		select {
		case <-tk.C:
			ds, err := this.checkDone()
			if err != nil {
				log.Println("ERROR: 检查完成状态失败: ", err.Error(), " 正在重新检查...")
			} else {
				st := ds.DoneState
				switch {
				case st == DoneState_OK || st == DoneState_OverTime || st == DoneState_ErrOccur:
					this.Cancel()
					this.Done <- ds
					break OUT_LOOP
				default:
					log.Println("任务状态: ", DoneStateToStr(ds.DoneState), ds.Msg)
				}
			}
		}
	}
}

// 节点保活
func (this *DbCoordinater) ttl(ctx context.Context) {
	// TODO 更新ci信息中的ttl字段
	updateTtl := func(t time.Time) {
		err := this.Db.Model(&CoordinateInfo{}).Where("batch_id = ? and node_id = ?",
			this.BatchId, this.Id).Update("ttl", t.Unix()).Error
		if err != nil {
			log.Println("ERROR: 心跳失败，稍后尝试下次心跳")
		}
	}

	tk := time.NewTicker(ttlTimeDuration)
	defer tk.Stop()

OUT_LOOP:
	for {
		select {
		case <-ctx.Done():
			break OUT_LOOP
		case t := <-tk.C:
			updateTtl(t)
		}
	}

	log.Println("INFO: 心跳退出")
}
