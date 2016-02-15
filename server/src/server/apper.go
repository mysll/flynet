package server

type Apper interface {
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
