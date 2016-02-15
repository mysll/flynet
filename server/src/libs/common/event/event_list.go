package event

type Event struct {
	Typ  string
	Args map[string]interface{}
}

type EventList struct {
	priorQueue chan *Event //高优先级队列
	eventQueue chan *Event //普通队列
	eventCache chan *Event //缓存
}

func (e *EventList) Push(t string, args map[string]interface{}, priority bool) {
	var event *Event
	select {
	case event = <-e.eventCache:
	default:
		event = new(Event)
	}

	event.Typ = t
	event.Args = args
	if priority {
		e.priorQueue <- event
		return
	}
	e.eventQueue <- event
}

func (e *EventList) Pop() *Event {
	var event *Event
	select {
	case event = <-e.priorQueue:
	case event = <-e.eventQueue:
	default:
	}
	return event
}

func (e *EventList) getEvent() *Event {
	var event *Event
	select {
	case event = <-e.eventCache:
	default:
		event = &Event{}
	}
	return event
}

func (e *EventList) FreeEvent(event *Event) {
	if event == nil {
		return
	}
	for k, _ := range event.Args {
		delete(event.Args, k)
	}
	select {
	case e.eventCache <- event:
	default:
	}
}

func NewEventList() *EventList {
	el := &EventList{}
	el.eventCache = make(chan *Event, 512)
	el.priorQueue = make(chan *Event, 32)
	el.eventQueue = make(chan *Event, 512)
	return el
}
