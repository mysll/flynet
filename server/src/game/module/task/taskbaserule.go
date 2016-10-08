package task

import (
	"logicdata/entity"
	"server/libs/log"
)

type TaskBaseRule struct {
	TaskBase
}

func NewTaskBaseRule() *TaskBaseRule {
	dt := &TaskBaseRule{}
	return dt
}

/**
任务承接条件判断
task_id: 任务编号
false:失败 true:成功
*/
func (tb *TaskBaseRule) OnTestTaskAccept(player *entity.Player, task *Task) bool {
	//是否已经接了
	if player.TaskAccepted_r.FindID(task.ID) != -1 {
		return false
	}

	//前置任务是否已经完成
	if task.PreTaskId != "" && task.PreTaskId != "0" {
		row := player.TaskAccepted_r.FindID(task.PreTaskId)
		if row == -1 {
			return false
		}

		if flag, _ := player.TaskAccepted_r.GetFlag(row); flag != TASK_FLAG_COMPLETE {
			return false
		}
	}

	//等级是否满足条件
	if task.Level > int32(player.GetLevel()) {
		return false
	}

	return true
}

/**
承接任务
task_id: 任务编号
false:失败 true:成功
*/
func (tb *TaskBaseRule) OnTaskAccept(player *entity.Player, task *Task) bool {
	player.TaskAccepted_r.AddRowValue(-1, task.ID, 0)

	if task.PropConditionID != "" && task.PropConditionID != "0" {
		pinfo := TaskInst.GetPropInfo(task.PropConditionID)
		if pinfo != nil {
			tb.AddPropRecord(player, task.ID, pinfo.Prop1, pinfo.PropValue1)
			tb.AddPropRecord(player, task.ID, pinfo.Prop2, pinfo.PropValue2)
			tb.AddPropRecord(player, task.ID, pinfo.Prop3, pinfo.PropValue3)
			tb.AddPropRecord(player, task.ID, pinfo.Prop4, pinfo.PropValue4)
			tb.AddPropRecord(player, task.ID, pinfo.Prop5, pinfo.PropValue5)
			tb.AddPropRecord(player, task.ID, pinfo.Prop6, pinfo.PropValue6)
			tb.AddPropRecord(player, task.ID, pinfo.Prop7, pinfo.PropValue7)
			tb.AddPropRecord(player, task.ID, pinfo.Prop8, pinfo.PropValue8)
			tb.AddPropRecord(player, task.ID, pinfo.Prop9, pinfo.PropValue9)
		} else {
			log.LogError("prop info(", task.PropConditionID, ") not found")
		}
	}

	if task.HuntID != "" && task.HuntID != "0" {
		hinfo := TaskInst.GetHuntinfo(task.HuntID)
		if hinfo != nil {
			tb.AddTaskRecord(player, task.ID, TYPE_HUNT, hinfo.MonsterID1, hinfo.MonsterNum1)
			tb.AddTaskRecord(player, task.ID, TYPE_HUNT, hinfo.MonsterID2, hinfo.MonsterNum2)
			tb.AddTaskRecord(player, task.ID, TYPE_HUNT, hinfo.MonsterID3, hinfo.MonsterNum3)
			tb.AddTaskRecord(player, task.ID, TYPE_HUNT, hinfo.MonsterID4, hinfo.MonsterNum4)
			tb.AddTaskRecord(player, task.ID, TYPE_HUNT, hinfo.MonsterID5, hinfo.MonsterNum5)
			tb.AddTaskRecord(player, task.ID, TYPE_HUNT, hinfo.MonsterID6, hinfo.MonsterNum6)
			tb.AddTaskRecord(player, task.ID, TYPE_HUNT, hinfo.MonsterID7, hinfo.MonsterNum7)
			tb.AddTaskRecord(player, task.ID, TYPE_HUNT, hinfo.MonsterID8, hinfo.MonsterNum8)
			tb.AddTaskRecord(player, task.ID, TYPE_HUNT, hinfo.MonsterID9, hinfo.MonsterNum9)
		} else {
			log.LogError("hunt info(", task.HuntID, ") not found")
		}
	}

	if task.CollectID != "" && task.CollectID != "0" {
		cinfo := TaskInst.GetCollection(task.CollectID)
		if cinfo != nil {
			tb.AddTaskRecord(player, task.ID, TYPE_COL, cinfo.ItemID1, cinfo.ItemCount1)
			tb.AddTaskRecord(player, task.ID, TYPE_COL, cinfo.ItemID2, cinfo.ItemCount2)
			tb.AddTaskRecord(player, task.ID, TYPE_COL, cinfo.ItemID3, cinfo.ItemCount3)
			tb.AddTaskRecord(player, task.ID, TYPE_COL, cinfo.ItemID4, cinfo.ItemCount4)
			tb.AddTaskRecord(player, task.ID, TYPE_COL, cinfo.ItemID5, cinfo.ItemCount5)
			tb.AddTaskRecord(player, task.ID, TYPE_COL, cinfo.ItemID6, cinfo.ItemCount6)
			tb.AddTaskRecord(player, task.ID, TYPE_COL, cinfo.ItemID7, cinfo.ItemCount7)
			tb.AddTaskRecord(player, task.ID, TYPE_COL, cinfo.ItemID8, cinfo.ItemCount8)
			tb.AddTaskRecord(player, task.ID, TYPE_COL, cinfo.ItemID9, cinfo.ItemCount9)
		} else {
			log.LogError("collection info (", task.CollectID, ")not found")
		}
	}

	if task.EventID != "" && task.EventID != "0" {
		einfo := TaskInst.GetEventInfo(task.EventID)
		if einfo != nil {
			tb.AddTaskRecord(player, task.ID, TYPE_EVENT, einfo.Event1, einfo.Count1)
			tb.AddTaskRecord(player, task.ID, TYPE_EVENT, einfo.Event2, einfo.Count2)
			tb.AddTaskRecord(player, task.ID, TYPE_EVENT, einfo.Event3, einfo.Count3)
			tb.AddTaskRecord(player, task.ID, TYPE_EVENT, einfo.Event4, einfo.Count4)
			tb.AddTaskRecord(player, task.ID, TYPE_EVENT, einfo.Event5, einfo.Count5)
			tb.AddTaskRecord(player, task.ID, TYPE_EVENT, einfo.Event6, einfo.Count6)
			tb.AddTaskRecord(player, task.ID, TYPE_EVENT, einfo.Event7, einfo.Count7)
			tb.AddTaskRecord(player, task.ID, TYPE_EVENT, einfo.Event8, einfo.Count8)
			tb.AddTaskRecord(player, task.ID, TYPE_EVENT, einfo.Event9, einfo.Count9)
		} else {
			log.LogError("event info(", task.EventID, ") not found")
		}
	}

	return true
}

/**
承接任务之后
task_id: 任务编号
false:失败 true:成功
*/
func (tb *TaskBaseRule) OnAffterTaskAccept(player *entity.Player, task *Task) bool {
	TaskInst.TestComplete(player, task)
	return true
}

/**
任务完成条件判断
task_id: 任务编号
false:失败 true:成功
*/
func (tb *TaskBaseRule) OnTestTaskComplete(player *entity.Player, task *Task) bool {
	return tb.CheckComplete(player, task) //检查是否已经有完成的任务了
}

func (tb *TaskBaseRule) GetReward(player *entity.Player, reward string, num int32) {
	switch reward {
	}
}

func (tb *TaskBaseRule) OnTaskComplete(player *entity.Player, task *Task) bool {

	row := player.TaskAccepted_r.FindID(task.ID)
	if row == -1 {
		return false
	}

	raward := TaskInst.GetReward(task.AwardID) //获取奖励
	if raward != nil {
		if raward.Reward1 != "" {
			tb.GetReward(player, raward.Reward1, raward.Num1)
		}
		if raward.Reward2 != "" {
			tb.GetReward(player, raward.Reward2, raward.Num2)
		}

		if raward.Reward3 != "" {
			tb.GetReward(player, raward.Reward3, raward.Num3)
		}

		if raward.Reward4 != "" {
			tb.GetReward(player, raward.Reward4, raward.Num4)
		}
	}

	player.TaskAccepted_r.SetFlag(row, TASK_FLAG_COMPLETE)

	return true
}

func (tb *TaskBaseRule) OnAffterTaskComplete(player *entity.Player, task *Task) bool {
	nextid := task.NextTaskId
	nexttask := TaskInst.GetTask(nextid)
	if nexttask == nil {
		return true
	}

	if nexttask.AutoAccept == 1 {
		TaskInst.TryAcceptTask(player, nexttask)
	}
	return true
}
