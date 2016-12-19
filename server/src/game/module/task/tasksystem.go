package task

import (
	"fmt"
	"logicdata/entity"
	"server"
	"server/data/datatype"
	"server/data/helper"
	"server/libs/log"
	"strings"
	"time"
)

const (
	TASK_INFO   = "conf_taskinfo.csv"
	TASK_COLL   = "conf_task_collect.csv"
	TASK_HUNT   = "conf_task_hunt.csv"
	TASK_EVENT  = "conf_task_event.csv"
	TASK_PROP   = "conf_task_prop.csv"
	TASK_REWARD = "conf_task_reward.csv"
)

type TaskSystem struct {
	server.Callee
	rules       []ITaskRule //任务规则
	taskinfos   map[string]*Task
	propcods    map[string]*PropCondition
	huntcods    map[string]*HuntCondition
	collcods    map[string]*CollCondition
	eventcods   map[string]*EventConditon
	rewardinfos map[string]*RewardInfo
}

func NewTaskSystem() *TaskSystem {
	ts := &TaskSystem{}
	ts.rules = []ITaskRule{
		NewTaskBaseRule(),
		NewDailyTask(),
	}

	ts.taskinfos = make(map[string]*Task)
	ts.propcods = make(map[string]*PropCondition)
	ts.huntcods = make(map[string]*HuntCondition)
	ts.collcods = make(map[string]*CollCondition)
	ts.eventcods = make(map[string]*EventConditon)
	ts.rewardinfos = make(map[string]*RewardInfo)
	return ts
}

func (ts *TaskSystem) LoadTaskInfo() bool {
	ids := helper.GetConfigIds(TASK_INFO)
	for _, id := range ids {
		task := &Task{}
		err := helper.LoadStructByFile(TASK_INFO, id, task)
		if err != nil {
			log.LogError(err)
			continue
		}

		ts.taskinfos[id] = task
	}
	log.LogMessage("load ", TASK_INFO, " ok, total:", len(ts.taskinfos))

	ids = helper.GetConfigIds(TASK_PROP)
	for _, id := range ids {
		pc := &PropCondition{}
		err := helper.LoadStructByFile(TASK_PROP, id, pc)
		if err != nil {
			log.LogError(err)
			continue
		}
		ts.propcods[id] = pc
	}
	log.LogMessage("load ", TASK_PROP, " ok, total:", len(ts.propcods))

	ids = helper.GetConfigIds(TASK_COLL)
	for _, id := range ids {
		col := &CollCondition{}
		err := helper.LoadStructByFile(TASK_COLL, id, col)
		if err != nil {
			log.LogError(err)
			continue
		}
		ts.collcods[id] = col
	}
	log.LogMessage("load ", TASK_COLL, " ok, total:", len(ts.collcods))

	ids = helper.GetConfigIds(TASK_HUNT)
	for _, id := range ids {
		hunt := &HuntCondition{}
		err := helper.LoadStructByFile(TASK_HUNT, id, hunt)
		if err != nil {
			log.LogError(err)
			continue
		}
		ts.huntcods[id] = hunt
	}
	log.LogMessage("load ", TASK_HUNT, " ok, total:", len(ts.huntcods))

	ids = helper.GetConfigIds(TASK_EVENT)
	for _, id := range ids {
		event := &EventConditon{}
		err := helper.LoadStructByFile(TASK_EVENT, id, event)
		if err != nil {
			log.LogError(err)
			continue
		}
		ts.eventcods[id] = event
	}
	log.LogMessage("load ", TASK_EVENT, " ok, total:", len(ts.eventcods))

	ids = helper.GetConfigIds(TASK_REWARD)
	for _, id := range ids {
		reward := &RewardInfo{}
		err := helper.LoadStructByFile(TASK_REWARD, id, reward)
		if err != nil {
			log.LogError(err)
			continue
		}
		ts.rewardinfos[id] = reward
	}
	log.LogMessage("load ", TASK_REWARD, " ok, total:", len(ts.rewardinfos))
	return true
}

func (ts *TaskSystem) GetTask(taskid string) *Task {
	if task, exist := ts.taskinfos[taskid]; exist {
		return task
	}
	return nil
}

func (ts *TaskSystem) GetReward(id string) *RewardInfo {
	if reward, exist := ts.rewardinfos[id]; exist {
		return reward
	}
	return nil
}

func (ts *TaskSystem) GetHuntinfo(id string) *HuntCondition {
	if hunt, exist := ts.huntcods[id]; exist {
		return hunt
	}
	return nil
}

func (ts *TaskSystem) GetCollection(id string) *CollCondition {
	if coll, exist := ts.collcods[id]; exist {
		return coll
	}
	return nil
}

func (ts *TaskSystem) GetPropInfo(id string) *PropCondition {
	if prop, exist := ts.propcods[id]; exist {
		return prop
	}
	return nil
}

func (ts *TaskSystem) GetEventInfo(id string) *EventConditon {
	if event, exist := ts.eventcods[id]; exist {
		return event
	}
	return nil
}

func (ts *TaskSystem) OnCommand(self datatype.Entity, sender datatype.Entity, msgid int, msg interface{}) int {
	if !entity.IsPlayer(self) {
		return 1
	}

	//player := self.(*entity.Player)
	switch msgid {

	}
	return 1
}

//玩家属性变化
func (ts *TaskSystem) OnPropertyChange(self datatype.Entity, prop string, old interface{}) int {
	if !entity.IsPlayer(self) {
		return 1
	}
	player := self.(*entity.Player)

	switch prop {
	}

	val, _ := player.Get(prop)

	row := player.TaskPropRecord_r.FindProperty(prop)
	for row != -1 {
		taskid, _, needval, _ := player.TaskPropRecord_r.GetRow(row)
		res, err := datatype.CompareNumber(needval, val)
		if err != nil {
			log.LogError(err)
		}

		if res <= 0 { //已经满条件，删除这条记录
			player.TaskPropRecord_r.Del(row)
			task := ts.GetTask(taskid)
			if task != nil {
				ts.TestComplete(player, task)
			}
			row--
		}

		row = player.TaskPropRecord_r.FindNextProperty(prop, row)
	}
	return 1
}

//容器里增加物品
func (ts *TaskSystem) OnAfterAdd(self datatype.Entity, sender datatype.Entity, index int) int {
	if !entity.IsContainer(self) {
		return 1
	}
	return 1
}

//每分钟更新一次
func (ts *TaskSystem) OnUpdate(player *entity.Player) {
	rows := player.TaskTimeLimit_r.GetRows()
	now := time.Now()
	for i := rows - 1; i >= 0; i-- {
		if now.Sub(time.Unix(player.TaskTimeLimit_r.Rows[i].EndTime, 0)).Seconds() >= 0 { //超时
			ts.DeleteTask(player, player.TaskTimeLimit_r.Rows[i].ID)
		}
	}
}

//rpc
func (ts *TaskSystem) Submit(player *entity.Player, taskid string) error {
	task := ts.GetTask(taskid)
	if task == nil {
		return fmt.Errorf("task not found")
	}

	ts.TryCompleteTask(player, task)
	return nil
}

func (ts *TaskSystem) UpdateRecordCompare(player *entity.Player, typ int32, key string, val int32) {
	if key == "" {
		return
	}

	row := player.TaskRecord_r.FindKey(key)
	for row != -1 {
		taskid, rtyp, _, oldval, amount, _, _ := player.TaskRecord_r.GetRow(row)
		if rtyp == typ { //有重复的
			if val >= amount {
				player.TaskRecord_r.Del(row)
				ts.TestComplete(player, ts.GetTask(taskid)) //检测是否完成
				row--
			} else if oldval < val {
				player.TaskRecord_r.SetCurrentAmount(row, val) //更新最大值
			}
		}

		row = player.TaskRecord_r.FindNextKey(key, row)
	}

	if typ == TYPE_EVENT && strings.HasPrefix(key, "G_") { //全局事件
		row = player.TaskGlobalRecord_r.FindKey(key)
		if row != -1 {
			oldval, _ := player.TaskGlobalRecord_r.GetCurrentAmount(row)
			if oldval < val {
				player.TaskGlobalRecord_r.SetCurrentAmount(row, val) //更新最大值
			}
			return
		}

		player.TaskGlobalRecord_r.AddRowValue(-1, typ, key, val)
	}
}

func (ts *TaskSystem) UpdateRecord(player *entity.Player, typ int32, key string, val int32) {
	if key == "" {
		return
	}

	if val <= 0 {
		return
	}

	row := player.TaskRecord_r.FindKey(key)
	for row != -1 {
		taskid, rtyp, _, oldval, amount, _, _ := player.TaskRecord_r.GetRow(row)
		if rtyp == typ { //有重复的
			if oldval+val >= amount {
				player.TaskRecord_r.Del(row)
				ts.TestComplete(player, ts.GetTask(taskid)) //检测是否完成
				row--
			} else {
				player.TaskRecord_r.SetCurrentAmount(row, oldval+val)
			}
		}

		row = player.TaskRecord_r.FindNextKey(key, row)
	}

	if typ == TYPE_EVENT && strings.HasPrefix(key, "G_") { //全局事件
		row = player.TaskGlobalRecord_r.FindKey(key)
		if row != -1 {
			oldval, _ := player.TaskGlobalRecord_r.GetCurrentAmount(row)
			player.TaskGlobalRecord_r.SetCurrentAmount(row, oldval+val)
			return
		}

		player.TaskGlobalRecord_r.AddRowValue(-1, typ, key, val)
	}
}

func (ts *TaskSystem) ClearGlobalRecord(player *entity.Player, key string) {
	row := player.TaskGlobalRecord_r.FindKey(key)
	if row != -1 {
		player.TaskGlobalRecord_r.Del(row)
	}
}

//删除某个ID的任务相关信息
func (ts *TaskSystem) DeleteTask(player *entity.Player, task_id string) {
	row := player.TaskAccepted_r.FindID(task_id)
	for row != -1 {
		player.TaskAccepted_r.Del(row)
		row = player.TaskAccepted_r.FindNextID(task_id, row-1)
	}

	row = player.TaskRecord_r.FindID(task_id)
	for row != -1 {
		player.TaskRecord_r.Del(row)
		row = player.TaskRecord_r.FindNextID(task_id, row-1)
	}

	row = player.TaskCanAccept_r.FindID(task_id)
	for row != -1 {
		player.TaskCanAccept_r.Del(row)
		row = player.TaskCanAccept_r.FindNextID(task_id, row-1)
	}

	row = player.TaskTimeLimit_r.FindID(task_id)
	for row != -1 {
		player.TaskTimeLimit_r.Del(row)
		row = player.TaskTimeLimit_r.FindNextID(task_id, row-1)
	}

	row = player.TaskPropRecord_r.FindID(task_id)
	for row != -1 {
		player.TaskPropRecord_r.Del(row)
		row = player.TaskPropRecord_r.FindNextID(task_id, row-1)
	}
}

func (ts *TaskSystem) NewDay(player *entity.Player) {
	rows := player.TaskAccepted_r.GetRows()
	deltask := make([]string, 0, 10)
	for r := 0; r < rows; r++ {
		if task, exist := ts.taskinfos[player.TaskAccepted_r.Rows[r].ID]; exist {
			if task.BaseType != TASK_TYPE_DIALY {
				continue
			}
			deltask = append(deltask, player.TaskAccepted_r.Rows[r].ID)
		}
	}

	ts.ClearGlobalRecord(player, "G_Active") //清除每日活跃
	for _, id := range deltask {
		ts.DeleteTask(player, id)
	}

	ts.CheckCanAcceptTask(player)
}

func (ts *TaskSystem) CheckTaskInfo(player *entity.Player) {
	rows := player.TaskAccepted_r.GetRows()
	deltask := make([]string, 0, 10)
	for r := 0; r < rows; r++ {
		if task, exist := ts.taskinfos[player.TaskAccepted_r.Rows[r].ID]; exist {
			ts.TestComplete(player, task)
			continue
		}
		deltask = append(deltask, player.TaskAccepted_r.Rows[r].ID)
	}

	for _, id := range deltask {
		//任务已经不存在
		ts.DeleteTask(player, id)
	}

	//检测可以接的任务
	ts.CheckCanAcceptTask(player)
}

func (ts *TaskSystem) CheckCanAcceptTask(player *entity.Player) {
	for _, task := range ts.taskinfos {
		if task.AutoAccept == 1 { //自动接
			ts.TryAcceptTask(player, task)
		}
	}
}

func (ts *TaskSystem) TestComplete(player *entity.Player, task *Task) bool {
	for _, r := range ts.rules {
		if !r.OnTestTaskComplete(player, task) {
			return false
		}
	}
	return true
}

//尝试承接任务
func (ts *TaskSystem) TryAcceptTask(player *entity.Player, task *Task) bool {
	//执行所有条件判断
	for _, r := range ts.rules {
		if !r.OnTestTaskAccept(player, task) {
			return false
		}
	}
	//执行所有任务承接前处理
	for _, r := range ts.rules {
		if !r.OnPreTaskAccept(player, task) {
			return false
		}
	}
	//执行所有任务承接处理
	for _, r := range ts.rules {
		if !r.OnTaskAccept(player, task) {
			return false
		}
	}
	//执行所有任务承接后处理
	for _, r := range ts.rules {
		if !r.OnAffterTaskAccept(player, task) {
			return false
		}
	}
	return true
}

//尝试完成任务
func (ts *TaskSystem) TryCompleteTask(player *entity.Player, task *Task) bool {
	//执行所有任务完成条件判断
	for _, r := range ts.rules {
		if !r.OnTestTaskComplete(player, task) {
			return false
		}
	}
	//执行所有任务完成前处理
	for _, r := range ts.rules {
		if !r.OnPreTaskComplete(player, task) {
			return false
		}
	}
	//执行所有任务完成处理
	for _, r := range ts.rules {
		if !r.OnTaskComplete(player, task) {
			return false
		}
	}
	//执行所有任务完成后处理
	for _, r := range ts.rules {
		if !r.OnAffterTaskComplete(player, task) {
			return false
		}
	}
	return true
}
