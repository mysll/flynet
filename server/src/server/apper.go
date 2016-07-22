package server

import (
	"server/data/datatype"
	"server/libs/rpc"
)

type Apper interface {
	//当前是否是base服务
	IsBase() bool
	//自主控制socket
	RawSock() bool
	//系统准备回调，底层已经初始化完成
	OnPrepare() bool
	//系统启动回调
	OnStart()
	//其它应用就绪回调
	OnReady(app string)
	//其它应用断开连接回调
	OnLost(app string)
	//关键应用已经准备好
	OnMustAppReady()
	//系统关闭回调
	OnShutdown() bool
	//app心跳回调
	OnBeatRun()
	//更新前回调
	OnBeginUpdate()
	//更新回调
	OnUpdate()
	//更新结束回调
	OnLastUpdate()
	//刷新状态回调(AOI等)
	OnFlush()
	//每一帧回调
	OnFrame()

	//事件回调
	OnEvent(e string, args map[string]interface{})
	//有新客户端连接回调
	OnClientConnected(id int64)
	//客户端断开连接
	OnClientLost(id int64)

	//从base从传送来的角色
	OnTeleportByBase(args []interface{}, player datatype.Entityer) bool
	//传送到场景完成回调
	OnSceneTeleported(mailbox rpc.Mailbox, result bool)
}

func (svr *Server) IsBase() bool {
	return false
}

//自主控制socket
func (svr *Server) RawSock() bool {
	return false
}

//系统准备回调，底层已经初始化完成
func (svr *Server) OnPrepare() bool {
	return true
}

//系统启动回调
func (svr *Server) OnStart() {

}

//其它应用就绪回调
func (svr *Server) OnReady(app string) {
}

//关键应用已经准备好
func (svr *Server) OnMustAppReady() {
}

//其它应用断开连接回调
func (svr *Server) OnLost(app string) {
}

//系统关闭回调
func (svr *Server) OnShutdown() bool {
	return true
}

//app心跳回调
func (svr *Server) OnBeatRun() {

}

//更新前回调
func (svr *Server) OnBeginUpdate() {

}

//更新回调
func (svr *Server) OnUpdate() {

}

//更新结束回调
func (svr *Server) OnLastUpdate() {

}

//每一帧回调
func (svr *Server) OnFrame() {

}

//刷新状态回调(AOI等)
func (svr *Server) OnFlush() {

}

//事件回调
func (svr *Server) OnEvent(e string, args map[string]interface{}) {

}

//有新客户端连接回调
func (svr *Server) OnClientConnected(id int64) {

}

//客户端断开连接
func (svr *Server) OnClientLost(id int64) {

}

//从base从传送来的角色
func (svr *Server) OnTeleportByBase(args []interface{}, player datatype.Entityer) bool {
	return false
}

//传送到场景完成回调
func (svr *Server) OnSceneTeleported(mailbox rpc.Mailbox, result bool) {

}
