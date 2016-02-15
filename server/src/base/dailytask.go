package base

type DailyTask struct {
	TaskBase
}

func NewDailyTask() *DailyTask {
	dt := &DailyTask{}
	return dt
}
