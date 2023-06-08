package network

import (
	"github.com/jzyong/TcpStressTesting/core/manager"
	"github.com/jzyong/TcpStressTesting/core/model"
	"github.com/jzyong/golib/log"
	"github.com/jzyong/golib/util"
	"runtime"
	"strconv"
)

/*
消息分发管理抽象层
*/
type MessageDistribute interface {
	RunHandler(message TcpMessage)                    //处理消息
	RegisterHandler(msgId int32, handler *TcpHandler) //为消息添加具体的处理逻辑
	StartWorkerPool()                                 //启动worker工作池
	SendMessageToTaskQueue(message TcpMessage)        //将消息交给TaskQueue,由worker进行处理
	ExecuteFunc(distributeId int64, f func())         //执行函数
}

// 处理未注册消息，如转发到大厅
type HandUnregisterMessageMethod func(message TcpMessage)

// 消息处理完成后的处理逻辑，这里用于统计协议
type HandMessageFinishMethod func(message TcpMessage)

// Handler 处理器
type messageDistributeImpl struct {
	handlers              map[int32]*TcpHandler       //存放每个MsgId 所对应的处理方法的map属性
	WorkerPoolSize        uint32                      //业务工作Worker池的数量
	TaskQueue             []chan TcpMessage           //Worker负责取任务的消息队列
	JobQueue              []chan Job                  //自定义函数
	HandUnregisterMessage HandUnregisterMessageMethod //处理未注册消息，如转发到大厅
	HandMessageFinish     HandMessageFinishMethod     //消息处理完成后的处理逻辑，这里用于统计协议
}

func NewMessageDistribute(workPoolSize uint32, unregisterMethod HandUnregisterMessageMethod, handFinishMethod HandMessageFinishMethod) MessageDistribute {
	return &messageDistributeImpl{
		handlers:       make(map[int32]*TcpHandler),
		WorkerPoolSize: workPoolSize,
		//一个worker对应一个queue
		TaskQueue:             make([]chan TcpMessage, workPoolSize),
		JobQueue:              make([]chan Job, workPoolSize),
		HandUnregisterMessage: unregisterMethod,
		HandMessageFinish:     handFinishMethod,
	}
}

// 执行函数 ，distributeId 一般为玩家id
func (mh *messageDistributeImpl) ExecuteFunc(distributeId int64, f func()) {
	workerID := uint32(distributeId) % mh.WorkerPoolSize
	//将请求消息发送给任务队列
	mh.JobQueue[workerID] <- FuncJob(f)
}

// 将消息交给TaskQueue,由worker进行处理
func (mh *messageDistributeImpl) SendMessageToTaskQueue(request TcpMessage) {
	//根据ConnID来分配当前的连接应该由哪个worker负责处理
	//轮询的平均分配法则

	//得到需要处理此条连接的workerID
	workerID := request.GetTcpChannel().GetConnID() % mh.WorkerPoolSize
	//将请求消息发送给任务队列
	mh.TaskQueue[workerID] <- request
}

// 马上以非阻塞方式处理消息
func (mh *messageDistributeImpl) RunHandler(msg TcpMessage) {

	//压测关闭后 服务器可能还在推送消息，直接忽略
	if manager.GetControlManager().TestState == model.TestIdle {
		return
	}

	handler, ok := mh.handlers[msg.GetMessageId()]
	if !ok {
		if mh.HandUnregisterMessage != nil {
			mh.HandUnregisterMessage(msg)
			return
		}
		log.Warn("协议 %d Handler 不存在 ", msg.GetMessageId())
		return
	}
	//执行对应处理方法
	defer func() {
		if r := recover(); r != nil {
			const size = 64 << 10
			buf := make([]byte, size)
			buf = buf[:runtime.Stack(buf, false)]
			log.Warn("handler 执行错误: %v %s", r, buf)
		}
	}()
	handler.run(msg)
	mh.HandMessageFinish(msg)
}

// 为消息添加具体的处理逻辑
func (mh *messageDistributeImpl) RegisterHandler(msgId int32, handler *TcpHandler) {
	//1 判断当前msg绑定的API处理方法是否已经存在
	if _, ok := mh.handlers[msgId]; ok {
		panic("repeated api , msgId = " + strconv.Itoa(int(msgId)))
	}
	//2 添加msg与handler的绑定关系
	mh.handlers[msgId] = handler
	//	log.Debug("Add handler %d ", msgId)
}

// 启动一个Worker工作流程
func (mh *messageDistributeImpl) startOneWorker(workerID int, taskQueue chan TcpMessage, job chan Job) {
	log.Info("工作组 ID %v 启动 ", workerID)
	//不断的等待队列中的消息
	for {
		select {
		//有消息则取出队列的Request，并执行绑定的业务方法
		case request := <-taskQueue:
			mh.RunHandler(request)
		case j := <-job:
			if r := recover(); r != nil {
				const size = 64 << 10
				buf := make([]byte, size)
				buf = buf[:runtime.Stack(buf, false)]
				log.Error("cron: panic running job: %v\n%s", r, buf)
			}
			j.Run()
		}

	}
}

// 启动worker工作池
func (mh *messageDistributeImpl) StartWorkerPool() {
	//遍历需要启动worker的数量，依此启动
	for i := 0; i < int(mh.WorkerPoolSize); i++ {
		//一个worker被启动
		//给当前worker对应的任务队列开辟空间
		mh.TaskQueue[i] = make(chan TcpMessage, 1024)
		mh.JobQueue[i] = make(chan Job, 1024)
		//启动当前Worker，阻塞的等待对应的任务队列是否有消息传递进来
		go mh.startOneWorker(i, mh.TaskQueue[i], mh.JobQueue[i])
	}
}

// 提交执行的任务
type Job interface {
	Run()
}
type FuncJob func()

func (f FuncJob) Run() { f() }

// 调度任务 ms为单位
type ScheduleJob struct {
	ExecuteTime  int64  //执行时间
	ExecuteCount int32  //重复执行次数
	IntervalTime int32  //间隔时间
	Fun          func() //执行的函数
}

// 构造调度任务 f执行函数
func NewScheduleJob(intervalTime, executeCount int32, f func()) *ScheduleJob {
	executeTime := util.CurrentTimeMillisecond() + int64(intervalTime)
	return &ScheduleJob{
		ExecuteTime:  executeTime,
		ExecuteCount: executeCount,
		IntervalTime: intervalTime,
		Fun:          f,
	}
}

// Execute 执行 返回剩余次数 0需要移除调度任务
func (s *ScheduleJob) Execute(messageDistribute MessageDistribute, nowTime, distributeId int64) int32 {
	if s.ExecuteCount < 1 {
		return 0
	}
	if nowTime > s.ExecuteTime {
		messageDistribute.ExecuteFunc(distributeId, s.Fun)
		s.ExecuteTime = nowTime + int64(s.IntervalTime)
		s.ExecuteCount--
	}
	return s.ExecuteCount
}
