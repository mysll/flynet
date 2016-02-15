package base

import (
	"data/datatype"
	"data/entity"
	"libs/log"
	"strings"
)

const (
	TASK_TYPE_NORMAL      = 1 + iota //普通任务
	TASK_TYPE_DIALY                  //每日任务
	TASK_TYPE_ACHIEVEMENT            //成就任务
)
const (
	TYPE_PROP = 1 + iota
	TYPE_COL
	TYPE_HUNT
	TYPE_EVENT
)

const (
	TASK_FLAG_NORMAL = iota
	TASK_FLAG_CANSUBMIT
	TASK_FLAG_COMPLETE
)

//任务信息
type Task struct {
	ID              string //任务编号
	Level           int32  //任务限制等级
	BaseType        int32  //任务集分类基础类型
	SubType         int32  //任务集分类子类型
	PreTaskId       string //前提任务
	NextTaskId      string //后续任务
	TimeLimitType   int8   //任务计时类型(0:无计时 1:普通计时 ...)
	TimeLimitLength int32  //任务计时长度限制(0:无限制 >0:时间长度,时间到后产生事件)(分钟)
	CantGiveup      int8   //是否不可放弃
	AwardID         string //奖励ID
	PropConditionID string //属性条件
	HuntID          string //猎杀条件
	CollectID       string //收集条件
	EventID         string //事件条件
	AutoAccept      int8   //是否自动接受
}

type PropCondition struct {
	Prop1, PropValue1 string
	Prop2, PropValue2 string
	Prop3, PropValue3 string
	Prop4, PropValue4 string
	Prop5, PropValue5 string
	Prop6, PropValue6 string
	Prop7, PropValue7 string
	Prop8, PropValue8 string
	Prop9, PropValue9 string
}

type HuntCondition struct {
	MonsterID1  string
	MonsterNum1 int32
	MonsterID2  string
	MonsterNum2 int32
	MonsterID3  string
	MonsterNum3 int32
	MonsterID4  string
	MonsterNum4 int32
	MonsterID5  string
	MonsterNum5 int32
	MonsterID6  string
	MonsterNum6 int32
	MonsterID7  string
	MonsterNum7 int32
	MonsterID8  string
	MonsterNum8 int32
	MonsterID9  string
	MonsterNum9 int32
}

type CollCondition struct {
	ItemID1    string
	ItemCount1 int32
	ItemID2    string
	ItemCount2 int32
	ItemID3    string
	ItemCount3 int32
	ItemID4    string
	ItemCount4 int32
	ItemID5    string
	ItemCount5 int32
	ItemID6    string
	ItemCount6 int32
	ItemID7    string
	ItemCount7 int32
	ItemID8    string
	ItemCount8 int32
	ItemID9    string
	ItemCount9 int32
}

type EventConditon struct {
	Event1 string
	Count1 int32
	Event2 string
	Count2 int32
	Event3 string
	Count3 int32
	Event4 string
	Count4 int32
	Event5 string
	Count5 int32
	Event6 string
	Count6 int32
	Event7 string
	Count7 int32
	Event8 string
	Count8 int32
	Event9 string
	Count9 int32
}

type RewardInfo struct {
	Reward1 string
	Num1    int32
	Reward2 string
	Num2    int32
	Reward3 string
	Num3    int32
	Reward4 string
	Num4    int32
}

type TaskBase struct {
}

type ITaskRule interface {
	OnTestTaskAccept(player *entity.Player, task *Task) bool
	OnPreTaskAccept(player *entity.Player, task *Task) bool
	OnTaskAccept(player *entity.Player, task *Task) bool
	OnAffterTaskAccept(player *entity.Player, task *Task) bool
	OnTestTaskComplete(player *entity.Player, task *Task) bool
	OnPreTaskComplete(player *entity.Player, task *Task) bool
	OnTaskComplete(player *entity.Player, task *Task) bool
	OnAffterTaskComplete(player *entity.Player, task *Task) bool
	OnTaskFail(player *entity.Player, task *Task) bool
	OnTaskSuccess(player *entity.Player, task *Task) bool
	OnGiveupTask(player *entity.Player, task *Task) bool
	OnFinishTaskTargets(player *entity.Player, task *Task) bool
	OnTaskTimeOut(player *entity.Player, task *Task) bool
}

/**
任务承接条件判断
task_id: 任务编号
false:失败 true:成功
*/
func (tb *TaskBase) OnTestTaskAccept(player *entity.Player, task *Task) bool {
	return true
}

/**
承接任务之前
task_id: 任务编号
false:失败 true:成功
*/
func (tb *TaskBase) OnPreTaskAccept(player *entity.Player, task *Task) bool {
	return true
}

/**
承接任务
task_id: 任务编号
false:失败 true:成功
*/
func (tb *TaskBase) OnTaskAccept(player *entity.Player, task *Task) bool {
	return true
}

/**
承接任务之后
task_id: 任务编号
false:失败 true:成功
*/
func (tb *TaskBase) OnAffterTaskAccept(player *entity.Player, task *Task) bool {
	return true
}

/**
任务完成条件判断
task_id: 任务编号
false:失败 true:成功
*/
func (tb *TaskBase) OnTestTaskComplete(player *entity.Player, task *Task) bool {
	return true
}

/**
完成任务之前
task_id: 任务编号
false:失败 true:成功
*/
func (tb *TaskBase) OnPreTaskComplete(player *entity.Player, task *Task) bool {
	return true
}

/**
完成任务
task_id: 任务编号
false:失败 true:成功
*/
func (tb *TaskBase) OnTaskComplete(player *entity.Player, task *Task) bool {
	return true
}

/**
完成任务之后
task_id: 任务编号
false:失败 true:成功
*/
func (tb *TaskBase) OnAffterTaskComplete(player *entity.Player, task *Task) bool {
	return true
}

/**
任务失败
task_id: 任务编号
false:失败 true:成功
*/
func (tb *TaskBase) OnTaskFail(player *entity.Player, task *Task) bool {
	return true
}

/**
任务成功
task_id: 任务编号
false:失败 true:成功
*/
func (tb *TaskBase) OnTaskSuccess(player *entity.Player, task *Task) bool {
	return true
}

/**
任务放弃
task_id: 任务编号
false:失败 true:成功
*/
func (tb *TaskBase) OnGiveupTask(player *entity.Player, task *Task) bool {
	return true
}

/**
完成任务目标处理接口
player:玩家对象 task_id:任务id
false:失败 true:成功
*/
func (tb *TaskBase) OnFinishTaskTargets(player *entity.Player, task *Task) bool {
	return true
}

/**
任务时间到达
player:玩家对象 task_id:任务id
false:失败 true:成功
*/
func (tb *TaskBase) OnTaskTimeOut(player *entity.Player, task *Task) bool {
	return true
}

func (tb *TaskBase) AddTaskRecord(player *entity.Player, taskid string, typ int32, key string, val int32) {
	if key == "" {
		return
	}
	row := player.TaskRecord_r.FindID(taskid)
	for row != -1 {
		_, rtyp, rkey, _, _, _, _ := player.TaskRecord_r.GetRow(row)
		if rtyp == typ && rkey == key { //有重复的
			return
		}
		row = player.TaskRecord_r.FindNextID(taskid, row)
	}

	if typ == TYPE_EVENT && strings.HasPrefix(key, "G_") { //全局事件
		row = player.TaskGlobalRecord_r.FindKey(key)
		if row != -1 {
			cur, _ := player.TaskGlobalRecord_r.GetCurrentAmount(row)
			player.TaskRecord_r.AddRowValue(-1, taskid, typ, key, cur, val, 0)
			return
		}
	}
	player.TaskRecord_r.AddRowValue(-1, taskid, typ, key, 0, val, 0)
}

func (tb *TaskBase) AddPropRecord(player *entity.Player, taskid string, property string, needval string) bool {
	if property == "" {
		return true
	}

	val, err := player.Get(property) //没有这个属性
	if err != nil {
		log.LogError("task need property not found, ", property)
		return true
	}

	res, err := datatype.CompareNumber(needval, val)
	if err != nil {
		log.LogError(err)
		return true
	}

	row := player.TaskPropRecord_r.FindID(taskid) //表中是否已经有记录
	for row != -1 {
		_, prop, _, _ := player.TaskPropRecord_r.GetRow(row)
		if prop == property {
			if res <= 0 { //已经满条件，删除这条记录
				player.TaskPropRecord_r.Del(row)
				return true
			}
			return false
		}
		row = player.TaskRecord_r.FindNextID(taskid, row)
	}

	if res <= 0 {
		return true
	}

	player.TaskPropRecord_r.AddRowValue(-1, taskid, property, needval)
	return false
}

func (tb *TaskBase) CheckComplete(player *entity.Player, task *Task) bool {

	taskid := task.ID

	taskrow := player.TaskAccepted_r.FindID(taskid)
	if taskrow == -1 {
		return false
	}

	flag, _ := player.TaskAccepted_r.GetFlag(taskrow)
	if flag == TASK_FLAG_COMPLETE { //已经完成
		return false
	}

	if flag == TASK_FLAG_CANSUBMIT {
		return true
	}

	row := player.TaskPropRecord_r.FindID(taskid)
	if row != -1 {
		return false
	}

	row = player.TaskRecord_r.FindID(taskid)
	for row != -1 {
		_, _, _, cur, need, _, _ := player.TaskRecord_r.GetRow(row)
		if cur < need { //有未完成的
			return false
		}
		//删除任务记录
		player.TaskRecord_r.Del(row)
		row = player.TaskRecord_r.FindNextID(taskid, row-1)
	}

	//设置完成标志
	player.TaskAccepted_r.SetFlag(taskrow, TASK_FLAG_CANSUBMIT)
	return true
}
