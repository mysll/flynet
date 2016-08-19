package server

const (
	PRIORITY_LOWEST  = 100
	PRIORITY_LOW     = 200
	PRIORITY_NORMAL  = 300
	PRIORITY_HIGH    = 400
	PRIORITY_HIGHEST = 500
)

type CalleeInfo struct {
	callee   []Objecter
	priority []int
}

var (
	callees = make(map[string]*CalleeInfo)
)

func (c *CalleeInfo) Add(callee Objecter, priority int) {

	index := -1
	for k, v := range c.priority {
		if v < priority {
			index = k
			break
		}
	}

	if index == len(c.priority) || index == -1 {
		c.callee = append(c.callee, callee)
		c.priority = append(c.priority, priority)
		return
	}

	c.callee = append(c.callee, callee)
	copy(c.callee[index+1:], c.callee[index:])
	c.callee[index] = callee
	c.priority = append(c.priority, priority)
	copy(c.priority[index+1:], c.priority[index:])
	c.priority[index] = priority
}

//注册回调
func RegisterCallee(typ string, c Objecter) error {
	return RegisterCalleePriority(typ, c, PRIORITY_NORMAL)
}

func RegisterCalleePriority(typ string, c Objecter, priority int) error {
	if _, dup := callees[typ]; !dup {
		callee := &CalleeInfo{}
		callee.callee = make([]Objecter, 0, 32)
		callee.priority = make([]int, 0, 32)
		callees[typ] = callee
	}
	callees[typ].Add(c, priority)
	return nil
}

//获取回调
func GetCallee(typ string) []Objecter {
	if _, dup := callees[typ]; !dup {
		callees[typ] = &CalleeInfo{}
	}
	return callees[typ].callee
}
