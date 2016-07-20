package server

import (
	"container/list"
	"fmt"
	"server/data/datatype"
)

const (
	MAX_POOL_FREE = 512
)

//对象池
type Pool struct {
	pool map[string]*list.List
}

//创建一个对象，如果池中有空闲的，则激活使用，否则新建一个对象
func (p *Pool) Create(typ string) datatype.Entityer {
	if l, exist := p.pool[typ]; exist {
		if e := l.Front(); e != nil {
			ele := e.Value.(datatype.Entityer)
			l.Remove(e)
			return ele
		}
	} else {
		p.pool[typ] = list.New()
	}

	if e := datatype.Create(typ); e != nil {
		return e
	}

	return nil
}

//释放一个具体对象，如果池中有空间，则回收，如果空闲的超过一定数量，则不回收，直接删除
func (p *Pool) Free(e datatype.Entityer) {
	if l, exist := p.pool[e.ObjTypeName()]; exist {
		if l.Len() <= MAX_POOL_FREE {
			e.Reset()
			l.PushBack(e)
		}
	} else {
		l := list.New()
		e.Reset()
		l.PushBack(e)
		p.pool[e.ObjTypeName()] = l
	}
}

//释放对象，如果有子对象一并释放
func (p *Pool) FreeObj(e datatype.Entityer) {
	chds := e.GetChilds()
	for _, ch := range chds {
		if ch != nil {
			p.FreeObj(ch)
		}
	}
	e.ClearChilds()
	p.Free(e)
}

//输出调试信息
func (p *Pool) DebugInfo() {
	fmt.Println("pool info:")
	for k, v := range p.pool {
		fmt.Println("pool:", k, ",", v.Len())
	}
	fmt.Println("pool info end")
}

//新建一个对象池
func NewEntityPool() *Pool {
	p := &Pool{}
	p.pool = make(map[string]*list.List)
	return p
}
