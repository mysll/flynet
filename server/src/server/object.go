package server

import (
	. "data/datatype"
)

type Objecter interface {
	// 名称：OnCreate
	// 描述：对象创建
	// 参数说明：
	//		self:对象
	//		sender:父对象
	OnCreate(self Entityer, sender Entityer) int
	// 名称：OnCreateRole
	// 描述：创建角色
	// 参数说明：
	//		self:角色
	//		args n: 自定义，客户端上传的创建参数
	OnCreateRole(self Entityer, args interface{}) int
	// 名称：OnDestroy
	// 描述：对象销毁
	// 参数说明：
	//		self:对象
	//		sender:父对象
	OnDestroy(self Entityer, sender Entityer) int
	// 名称：OnEntry
	// 描述：进入容器
	// 参数说明：
	//		self:对象
	//		sender:父对象
	OnEntry(self Entityer, sender Entityer) int
	// 名称：OnLeave
	// 描述：离开容器
	// 参数说明：
	//		self:对象
	//		sender:父开者
	OnLeave(self Entityer, sender Entityer) int
	// 名称：OnTestAdd
	// 描述：测试容器能否加入
	// 参数说明：
	//		self:对象
	//		sender:加入者
	//		index: 整型,index,欲加入的位置
	OnTestAdd(self Entityer, sender Entityer, index int) int
	// 名称：OnAdd
	// 描述：容器中加入
	// 参数说明：
	//		self:对象
	//		sender:加入者
	//		index: 整型,index,加入的位置
	// 返回值说明：
	//		ret(位操作):1:当前没有处理(默认：kernel来处理)
	//		 	 		2:叠加
	//					4:不能加入
	//					8:中断回调
	//		destindex:ret返回2时，提供叠加的位置
	OnAdd(self Entityer, sender Entityer, index int) (ret int, destindex int)
	// 名称：OnAfterAdd
	// 描述：容器中加入之后
	// 参数说明：
	//		self:对象
	//		sender:加入者
	//		index: 整型,index,加入的位置
	OnAfterAdd(self Entityer, sender Entityer, index int) int
	// 名称：OnBeforeRemove
	// 描述：容器中移走之前
	// 参数说明：
	//		self:对象
	//		sender:移走者
	OnBeforeRemove(self Entityer, sender Entityer) int
	// 名称：OnRemove
	// 描述：容器中移走
	// 参数说明：
	//		self:对象
	//		sender:移走者
	OnRemove(self Entityer, sender Entityer, index int) int
	// 名称：OnLoad
	// 描述：数据载入完成
	// 参数说明：
	//		self:对象
	//		typ: 整型, type, 数据读取的类型
	OnLoad(self Entityer, typ int) int
	// 名称：OnStore
	// 描述：数据保存之前
	// 参数说明：
	//		self:对象
	//		typ: 整型, type, 数据保存的类型
	OnStore(self Entityer, typ int) int
	// 名称：OnDisconnect
	// 描述：客户端断线
	// 参数说明：
	//		self:对象
	OnDisconnect(self Entityer) int
	// 名称：OnEntryScene
	// 描述：玩家进入场景之前（SMSG_ENTRY_SCENE消息之前）
	// 参数说明：
	//		self:对象
	OnEntryScene(self Entityer) int

	// 名称：OnEnterScene
	// 描述：玩家进入场景之后（已收到场景模型和主角信息）
	// 参数说明：
	//		self:对象
	OnEnterScene(self Entityer) int

	// 名称：OnCommand
	// 描述：接收命令
	// 参数说明：
	//		self:对象
	//		sender:发送者
	//      msgid:消息id
	//		msg: 消息内容
	OnCommand(self Entityer, sender Entityer, msgid int, msg interface{}) int

	// 名称：OnUse
	// 描述：使用道具
	// 参数说明：
	//		self:道具
	//		sender:使用者
	OnUse(self Entityer, sender Entityer) int

	// 名称：OnUseTo
	// 描述：使用道具到其他物体
	// 参数说明：
	//		self:道具
	//		sender:使用者
	//		target, 被使用的其他物体
	OnUseTo(self Entityer, sender Entityer, target Entityer) int

	// 名称：OnEquip
	// 描述：装备道具
	// 参数说明：
	//		self:道具
	//		sender:使用者
	OnEquip(self Entityer, sender Entityer, idx int) int

	// 名称：OnEquip
	// 描述：卸下道具
	// 参数说明：
	//		self:道具
	//		sender:使用者
	OnUnEquip(self Entityer, sender Entityer, idx int) int

	// 名称：OnPropertyChange
	// 描述：属性变动
	// 参数说明：
	//		self:entity
	//		prop:属性名
	//    	old:原始值
	OnPropertyChange(self Entityer, prop string, old interface{}) int

	// 名称：OnTimer
	// 描述：定时器回调
	// 参数说明：
	//		self:entity
	//		args:参数
	//    	count:定时器剩余次数
	OnTimer(self Entityer, count int32, args interface{}) int

	// 名称：OnReady
	// 描述：客户端就绪
	// 参数说明：
	//		self:对象
	//		first:  是否是进入游戏后的第一次收到客户端就绪的消息
	OnReady(self Entityer, first bool) int
}

type Callee struct {
}

func (c *Callee) OnCreate(self Entityer, sender Entityer) int {
	return 1
}
func (c *Callee) OnCreateRole(self Entityer, args interface{}) int {
	return 1
}
func (c *Callee) OnDestroy(self Entityer, sender Entityer) int {
	return 1
}
func (c *Callee) OnEntry(self Entityer, sender Entityer) int {
	return 1
}
func (c *Callee) OnLeave(self Entityer, sender Entityer) int {
	return 1
}
func (c *Callee) OnTestAdd(self Entityer, sender Entityer, index int) int {
	return 1
}
func (c *Callee) OnAdd(self Entityer, sender Entityer, index int) (int, int) {
	return 1, -1
}
func (c *Callee) OnAfterAdd(self Entityer, sender Entityer, index int) int {
	return 1
}
func (c *Callee) OnBeforeRemove(self Entityer, sender Entityer) int {
	return 1
}
func (c *Callee) OnRemove(self Entityer, sender Entityer, index int) int {
	return 1
}
func (c *Callee) OnLoad(self Entityer, typ int) int {
	return 1
}
func (c *Callee) OnStore(self Entityer, typ int) int {
	return 1
}
func (c *Callee) OnDisconnect(self Entityer) int {
	return 1
}
func (c *Callee) OnEntryScene(self Entityer) int {
	return 1
}
func (c *Callee) OnEnterScene(self Entityer) int {
	return 1
}
func (c *Callee) OnCommand(self Entityer, sender Entityer, msgid int, msg interface{}) int {
	return 1
}
func (c *Callee) OnUse(self Entityer, sender Entityer) int {
	return 1
}
func (c *Callee) OnUseTo(self Entityer, sender Entityer, target Entityer) int {
	return 1
}
func (c *Callee) OnEquip(self Entityer, sender Entityer, idx int) int {
	return 1
}
func (c *Callee) OnUnEquip(self Entityer, sender Entityer, idx int) int {
	return 1
}
func (c *Callee) OnPropertyChange(self Entityer, prop string, old interface{}) int {
	return 1
}

func (c *Callee) OnTimer(self Entityer, count int32, args interface{}) int {
	return 1
}

func (c *Callee) OnReady(self Entityer, first bool) int {
	return 1
}
