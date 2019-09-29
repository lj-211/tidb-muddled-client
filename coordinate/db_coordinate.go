package coordinate

import (
	"context"
	"log"
	"time"

	"github.com/jinzhu/gorm"
	cdb "github.com/lj-211/common/db"
	"github.com/pkg/errors"

	"github.com/lj-211/tidb-muddled-client/common"
)

const TtlTime time.Duration = time.Second * 5
const ExpireTime time.Duration = TtlTime * 2

// -------------------------------------------------------------------------
// model
// 初始化过程
//	1. insert ignore into
//	2. 如果存在，则检查是否存活，如果存活则异常退出
//	3. 如果插入成功，则开始保活
const (
	CiStatus_Regist = iota
	CiStatus_InitOk
	CiStatus_Completed
)

type CoordinateInfo struct {
	gorm.Model
	BatchId     string // 批次id
	NodeId      string // 节点id
	TaskCnt     int    // 任务数量
	DoneTaskCnt int    // 已执行任务数量
	Status      int    // 状态
}

func (e *CoordinateInfo) TableName() string {
	return "coordinate_info"
}

type CmdInfo struct {
	gorm.Model
	BatchId string // 批次id
	NodeId  string // 命令所属的节点id
	Sql     string // 命令
}

func (e *CmdInfo) TableName() string {
	return "cmd_info"
}

type CmdOrder struct {
	gorm.Model
	BatchId string // 批次id
	CmdId   uint   // 命令id
	NodeId  string // 节点id
	IsDone  int    // 是否完成
}

func (e *CmdOrder) TableName() string {
	return "cmd_order"
}

// -------------------------------------------------------------------------
// coordinater
type DbCoordinater struct {
	Db       *gorm.DB
	Id       string
	BatchId  string
	Partners []string
	Ctx      context.Context
	Cancel   context.CancelFunc
	Done     chan bool
	InitOk   chan bool
	Proc     TaskProcesser
}

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

func (this *DbCoordinater) Start(ctx context.Context, proc TaskProcesser) error {
	this.Proc = proc
	return this.DoStart(this.BatchId, this.Partners)
}

func (this *DbCoordinater) DoStart(bid string, partners []string) error {
	this.Ctx, this.Cancel = context.WithCancel(context.Background())
	this.Done = make(chan bool)
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

		/*
			cnt := 0
			err := this.Db.Model(ci).Where("node_id in ?", partners).Count(&cnt).Error
			if err != nil {
				return errors.Wrap(err, "查询是否启动完成错误")
			}
		*/
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
	go func() {
		tk := time.NewTicker(time.Second)
		defer tk.Stop()

		for {
			select {
			case <-tk.C:
				done, err := this.CheckDone()
				if err != nil {
					log.Println("ERROR: 检查完成状态失败: ", err.Error())
				} else {
					if done {
						this.Cancel()
						this.Done <- true
						return
					}
				}
			}
		}
	}()

	return nil
}

// NOTE
//	这里有个特殊逻辑，如果done == true的话，只处理推送任务完成逻辑，忽略data
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
	}

	return nil
}

func (this *DbCoordinater) Watch(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-this.InitOk:
			break
		}
	}

	tk := time.NewTicker(time.Second)
	defer tk.Stop()
	// start watch task
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-tk.C:
			err := this.DoTask(ctx)
			if err != nil {
				log.Println("ERROR: 执行任务失败")
			}
		}
	}

	return nil
}

func (this *DbCoordinater) DoTask(ctx context.Context) error {
	co := &CmdOrder{}
	err := this.Db.Model(co).Where("batch_id = ? and is_done = 0").
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
		err = this.Proc(db, sql)
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

func (this *DbCoordinater) IsDone(ctx context.Context) bool {
	v := <-this.Done
	return v
}

func (this *DbCoordinater) CheckDone() (bool, error) {
	done := false

	// 伙伴节点的任务全部都完成
	allIds := make([]string, 0)
	allIds = append(allIds, this.Id)
	allIds = append(allIds, this.Partners...)

	cis := make([]*CoordinateInfo, 0)
	err := this.Db.Model(&CoordinateInfo{}).
		Where("batch_id = ? and node_id in ?", this.BatchId, allIds).Find(&cis).Error
	if err != nil {
		return done, errors.Wrap(err, "查询是否完成失败")
	}
	allTask := 0
	doneTask := 0
	allInitOk := true
	for i := 0; i < len(cis); i++ {
		v := cis[i]
		if v.Status != CiStatus_InitOk {
			allInitOk = false
			break
		}
		allTask += v.TaskCnt
		doneTask += v.DoneTaskCnt
	}

	if allInitOk && allTask == doneTask {
		done = true
	}

	return done, nil
}

func (this *DbCoordinater) genTaskOrder(ctx context.Context) error {
	// TODO
	//	这里的查询以及下面的查询可能存在过多数据查询和大事务的问题，
	//	可以根据业务情况优化
	cmds := make([]*CmdInfo, 0)
	err := this.Db.Model(&CmdInfo{}).Where("batch_id = ?", this.BatchId).Find(&cmds).Error
	if err != nil {
		return errors.Wrap(err, "查询任务列表失败")
	}

	nums := make([]int, len(cmds))
	for i := 0; i < len(nums); i++ {
		nums[i] = i
	}
	orders := common.FullPermutation(nums)

	trans := func(db *gorm.DB) error {
		var err error

	OUT_LOOP:
		for _, v := range orders {
			for _, e := range v {
				co := &CmdOrder{
					BatchId: this.BatchId,
					CmdId:   cmds[e].ID,
					NodeId:  cmds[e].NodeId,
				}

				err = db.Model(co).Create(co).Error
				if err != nil {
					break OUT_LOOP
				}
			}
		}

		return err
	}

	if err := cdb.DoTrans(this.Db, trans); err != nil {
		return errors.Wrap(err, "执行生成任务失败")
	}

	return nil
}

func (this *DbCoordinater) WatchInitOk(ctx context.Context) {
	tk := time.NewTicker(time.Second * 2)
	defer tk.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-tk.C:
			// 1. do query if all node is OK
			allIds := make([]string, 0)
			allIds = append(allIds, this.Id)
			allIds = append(allIds, this.Partners...)
			cis := make([]*CoordinateInfo, 0)
			err := this.Db.Model(&CoordinateInfo{}).
				Where("batch_id = ? and node_id in ?", this.BatchId, allIds).Find(&cis).Error
			if err != nil {
				break
			}

			allOk := true
			for i := 0; i < len(cis); i++ {
				v := cis[i]
				if v.Status != CiStatus_InitOk {
					allOk = false
					break
				}
			}
			if !allOk {
				break
			}

			err = this.genTaskOrder(ctx)
			if err != nil {
				break
			}

			this.InitOk <- true

			return
		}
	}
}
