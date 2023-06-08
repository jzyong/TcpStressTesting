package model

// 测试状态
const (
	TestIdle    = 0 //空闲
	TestRunning = 1 //测试中
	TestQuit    = 2 //退出中
)

// 消息请求失败时间 纳秒
const MessageRequestFailTime = 5000 * 1000000

// Mb
const Mb = 1024 * 1024
