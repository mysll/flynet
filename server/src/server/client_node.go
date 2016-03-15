package server

import (
	"fmt"
	"io"
	"libs/log"
	"libs/rpc"
	"sync"
	"time"
)

//每个客户端的连接
type ClientNode struct {
	Session      int64
	DelayDestroy int
	Connected    bool
	rwc          io.ReadWriteCloser
	Sendbuf      chan *Message
	Addr         string
	quit         bool
	MailBox      rpc.Mailbox
	lastping     time.Time
}

func NewClientNode() *ClientNode {
	cn := &ClientNode{}
	cn.Sendbuf = make(chan *Message, 32)

	return cn
}

func (c *ClientNode) Ping() {
	c.lastping = time.Now()
}

//向客户端发包，存进一个消息队列，异步发送
func (c *ClientNode) SendMessage(message *Message) error {
	if c.quit {
		return fmt.Errorf("client is quit")
	}
	message.Dup()
	select {
	case c.Sendbuf <- message:
		return nil
	default:
		message.Free()
		return fmt.Errorf("send buffer full")
	}
}

//向客户端发包，存进一个消息队列，异步发送
func (c *ClientNode) Send(data []byte) error {
	if c.quit {
		return fmt.Errorf("client is quit")
	}
	message := NewMessage(len(data))
	message.Body = append(message.Body, data...)
	select {
	case c.Sendbuf <- message:
		return nil
	default:
		message.Free() //满了，直接扔掉
		return fmt.Errorf("send buffer full")
	}
}

//延迟删除计数
func (c *ClientNode) DelCount(count int) {
	c.DelayDestroy = count
}

//主循环
func (c *ClientNode) Run() {
	c.quit = false
	go c.innerSend()
}

//向客户端发送数据
func (c *ClientNode) innerSend() {
sendloop:
	for {
		select {
		case message := <-c.Sendbuf:
			if c.rwc != nil {
				_, err := c.rwc.Write(message.Body)
				log.LogMessage("send message to client:", c.Addr, " size:", len(message.Body))
				message.Free()
				if err != nil {
					break sendloop
				}
			} else {
				if message != nil {
					message.Free()
					break sendloop
				}
			}
		default:
			if c.quit {
				break sendloop
			}

			time.Sleep(time.Millisecond)
		}
	}

	//丢弃所有消息
	for {
		select {
		case m := <-c.Sendbuf:
			m.Free()
		default:
			log.LogMessage("client node quit loop")
			return
		}
	}

}

//关闭连接
func (c *ClientNode) Close() {
	if c.rwc != nil {
		c.rwc.Close()
	}
	c.quit = true
	c.Connected = false
}

//客户端连接池
type ClientList struct {
	l         sync.RWMutex
	freeNodes *ClientNode
	clients   map[int64]*ClientNode
	serial    int64
}

//关闭所有连接
func (cl *ClientList) CloseAll() {
	cl.l.Lock()
	defer cl.l.Unlock()
	for _, node := range cl.clients {
		node.Close()
	}
}

//增加一个客户端连接
func (cl *ClientList) Add(c io.ReadWriteCloser, addr string) int64 {
	cl.l.Lock()
	defer cl.l.Unlock()
	cl.serial++
	cn := NewClientNode()
	cn.Session = cl.serial
	cn.Connected = true
	cn.Addr = addr
	cn.rwc = c
	cn.DelayDestroy = -1
	cn.lastping = time.Now()
	cl.clients[cl.serial] = cn
	return cl.serial
}

//交换两个socket,用于顶号处理
func (cl *ClientList) Switch(src, dest int64) bool {
	cl.l.Lock()
	defer cl.l.Unlock()
	var srcnode *ClientNode
	var destnode *ClientNode
	var ok bool
	if srcnode, ok = cl.clients[src]; !ok {
		return false
	}

	if destnode, ok = cl.clients[dest]; !ok {
		return false
	}

	srcnode.Session, destnode.Session = dest, src
	srcnode.MailBox, destnode.MailBox = destnode.MailBox, srcnode.MailBox
	cl.clients[src], cl.clients[dest] = destnode, srcnode

	return true
}

//移除一个客户端连接
func (cl *ClientList) Remove(session int64) {
	cl.l.Lock()
	defer cl.l.Unlock()
	if cn, ok := cl.clients[session]; ok {
		delete(cl.clients, session)
		log.LogDebug("remove client node, ", cn.Addr)
		cn.Close()
	}
}

//检查需要删除的连接（延迟删除）
func (cl *ClientList) Check() {
	cl.l.Lock()
	defer cl.l.Unlock()
	for _, v := range cl.clients {
		if v.DelayDestroy > 0 {
			v.DelayDestroy--
			if v.DelayDestroy == 0 {
				v.Close()
			}
			continue
		}

		if time.Now().Sub(v.lastping).Minutes() > 5.0 { //5分钟超时
			v.Close()
		}
	}
}

//通过session获取连接
func (cl *ClientList) FindNode(session int64) *ClientNode {
	cl.l.RLock()
	defer cl.l.RUnlock()
	if cn, ok := cl.clients[session]; ok {
		return cn
	}

	return nil
}

//创建新的连接池
func NewClientList() *ClientList {
	cl := &ClientList{}
	cl.clients = make(map[int64]*ClientNode)
	return cl
}
