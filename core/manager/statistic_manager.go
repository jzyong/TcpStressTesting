package manager

import (
	"container/list"
	"context"
	"github.com/jzyong/TcpStressTesting/config"
	"github.com/jzyong/TcpStressTesting/core/model"
	"github.com/jzyong/TcpStressTesting/core/proto"
	"github.com/jzyong/golib/log"
	"github.com/jzyong/golib/util"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	"sync"
	"time"
)

// 压测数据统计，主节点操作
// master相关的统计处理丢到一个 go routine，防止多线程问题
type StatisticManager struct {
	util.DefaultModule
	UserCount                        int32                                            //用户数 (worker)
	MemoryUsePercent                 int32                                            //内存使用百分比
	CpuUsePercent                    float64                                          //CPU使用百分比
	PlayerMessageInfoChan            chan *proto.PlayerMessageInfo                    //玩家消息队列
	StatisticMessageChan             chan *proto.UploadStatisticsRequest              //上传消息
	uploadStatisticChan              chan bool                                        //上传统计信息，必须在同一例程中，否则有并发修改map
	WorkerStatisticMessageInfos      map[int32]*proto.MessageInterfaceInfo            //worker消息统计
	MasterStatisticMessageInfos      map[int32]*proto.MessageInterfaceInfo            //master消息统计汇总
	WorkerStatisticMessageInfoCaches map[string]map[int32]*proto.MessageInterfaceInfo //master缓存所有worker的统计 进行求和取平均值计算
	statisticUploadCount             int                                              //统计消息上传次数
	StatisticLogs                    *list.List                                       //统计日志
	MessageInterfaceCount            int
	statisticLock                    sync.RWMutex //锁
}

var statisticManager *StatisticManager
var statisticOnce sync.Once

func GetStatisticManager() *StatisticManager {
	statisticOnce.Do(func() {
		statisticManager = &StatisticManager{}
	})
	return statisticManager
}

// 开始启动
func (receiver *StatisticManager) Init() error {
	receiver.MessageInterfaceCount = 1
	receiver.PlayerMessageInfoChan = make(chan *proto.PlayerMessageInfo, 1024)
	receiver.WorkerStatisticMessageInfos = make(map[int32]*proto.MessageInterfaceInfo)
	receiver.uploadStatisticChan = make(chan bool, 64)
	if config.ApplicationConfigInstance.Master {
		receiver.WorkerStatisticMessageInfoCaches = make(map[string]map[int32]*proto.MessageInterfaceInfo)
		receiver.MasterStatisticMessageInfos = make(map[int32]*proto.MessageInterfaceInfo)
		receiver.StatisticMessageChan = make(chan *proto.UploadStatisticsRequest, 1024)
		receiver.StatisticLogs = list.New()
	}
	log.Info("[统计] 初始化")
	return nil
}

// 开始启动
func (receiver *StatisticManager) Run() {
	GetControlManager().CronScheduler.AddFunc("@every 1s", receiver.updateSecond)
	GetControlManager().CronScheduler.AddFunc("@every 5s", receiver.updateFiveSecond)
	go receiver.calculateStatistic()

}

// 关闭连接
func (receiver *StatisticManager) Stop() {
}

// 每秒更新
func (receiver *StatisticManager) updateSecond() {
	receiver.uploadStatisticChan <- true //将发送消息放入chan中，不能直接发生，否则并发修改
	//receiver.calculateWorkerStatistic()

	//采集内存和cpu
	v, _ := mem.VirtualMemory()
	receiver.MemoryUsePercent = int32(v.UsedPercent)
	cpuPercents, _ := cpu.Percent(time.Second*1, true)
	for _, cpuPercent := range cpuPercents {
		receiver.CpuUsePercent = cpuPercent
		//log.Debug("cpu :%v", cpuPercent)
	}

}

// 每5秒更新
func (receiver *StatisticManager) updateFiveSecond() {

}

// 计算合并worker的统计信息到 master
func (receiver *StatisticManager) calculateWorkerStatistic() {
	receiver.statisticUploadCount++
	// 添加次数频率统计，不然每次收到消息计算？ 每收到两次消息统计一次
	index := receiver.statisticUploadCount % (len(receiver.WorkerStatisticMessageInfoCaches))
	if index != 0 {
		return
	}

	//统计每条协议的
	for messageId, _ := range receiver.WorkerStatisticMessageInfos {
		masterMessageInfo := receiver.MasterStatisticMessageInfos[messageId]
		if masterMessageInfo == nil {
			masterMessageInfo = &proto.MessageInterfaceInfo{
				MessageId: messageId,
				DelayMin:  model.MessageRequestFailTime,
			}
			receiver.MasterStatisticMessageInfos[messageId] = masterMessageInfo
		}

		masterMessageInfo.MergeClear()
		for _, infos := range receiver.WorkerStatisticMessageInfoCaches {
			workerMessageInfo := infos[messageId]
			if workerMessageInfo == nil {
				continue
			}
			masterMessageInfo.MergeAdd(workerMessageInfo)
		}
	}

	//统计总共的
	info := &proto.MessageInterfaceInfo{DelayMin: model.MessageRequestFailTime}
	for _, interfaceInfo := range receiver.MasterStatisticMessageInfos {
		info.MergeAdd(interfaceInfo)
	}
	receiver.StatisticLogs.PushBack(info)
}

// 计算合并玩家的统计信息到 worker
func (receiver *StatisticManager) calculateStatistic() {
	for {
		select {
		// 从chan中读取玩家消息进行统计
		case playerMessageInfo := <-receiver.PlayerMessageInfoChan:
			info := receiver.WorkerStatisticMessageInfos[playerMessageInfo.MessageId]
			if info == nil {
				info = &proto.MessageInterfaceInfo{
					MessageId: playerMessageInfo.MessageId,
					StartTime: util.Now().UnixNano(),
					DelayMin:  model.MessageRequestFailTime,
				}
				receiver.WorkerStatisticMessageInfos[playerMessageInfo.MessageId] = info
			}
			info.Add(playerMessageInfo)
		case statisticRequest := <-receiver.StatisticMessageChan:
			receiver.WorkerStatisticMessageInfoCaches[statisticRequest.WorkerServer.Name] = statisticRequest.InterfaceInfo
			receiver.calculateWorkerStatistic()
		case uploadStatistics := <-receiver.uploadStatisticChan:
			if uploadStatistics {
				receiver.uploadStatisticSend()
			}

		}
	}
}

// 上传统计信息，无论是否在压测
// 同时充当心跳检测，发送消息失败移除master客户端
func (receiver *StatisticManager) uploadStatisticSend() {
	receiver.statisticLock.Lock()
	defer receiver.statisticLock.Unlock()
	masterClient := GetNetworkManager().MasterClient
	if masterClient == nil {
		return
	}

	request := &proto.UploadStatisticsRequest{}

	//服务器主机信息：
	workerInfo := &proto.UploadStatisticsRequest_WorkerServerInfo{
		Name:        config.ApplicationConfigInstance.Name,
		RpcHost:     config.ApplicationConfigInstance.RpcHost,
		CpuRate:     receiver.CpuUsePercent,
		MemorySize:  receiver.MemoryUsePercent,
		PlayerCount: receiver.UserCount,
	}
	request.WorkerServer = workerInfo

	//统计数据
	if GetControlManager().TestState == model.TestRunning {
		request.InterfaceInfo = receiver.WorkerStatisticMessageInfos
	}

	c := proto.NewStressTestingInnerServiceClient(masterClient.GrpcClient)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()
	////执行对应处理方法
	//defer func() {
	//	//都加锁了，还存在并发迭代？ fatal error: concurrent map iteration and map write
	//	if r := recover(); r != nil {
	//		log.Warn("同步数据错误 %v ", r)
	//	}
	//}()
	_, err := c.UploadStatistics(ctx, request)

	//移除master客户端，从新创建
	if err != nil {
		log.Warn("master 连接不可用：%v", err)
		masterClient.GrpcClient.Close()
		GetNetworkManager().MasterClient = nil
		return
	}
}

// 接收统计数据 worker --> master
func (receiver *StatisticManager) UploadStatisticReceive(request *proto.UploadStatisticsRequest) {
	masterClient := GetNetworkManager().MasterClient
	if masterClient == nil {
		return
	}

	workerServerInfo := request.WorkerServer
	if workerServerInfo != nil {
		GetNetworkManager().UpdateWorkerClient(workerServerInfo)
	}
	interfaceInfos := request.InterfaceInfo
	//log.Info("%v 接口信息：%v", workerServerInfo.Name, interfaceInfos)
	if interfaceInfos != nil {
		receiver.StatisticMessageChan <- request //避免并发的写操作
	}
}

// 重置统计数据
func (receiver *StatisticManager) ResetStatisticData() {
	receiver.UserCount = 0
	receiver.WorkerStatisticMessageInfos = make(map[int32]*proto.MessageInterfaceInfo)
	if config.ApplicationConfigInstance.Master {
		receiver.StatisticLogs.Init()
		receiver.MasterStatisticMessageInfos = make(map[int32]*proto.MessageInterfaceInfo)
		receiver.WorkerStatisticMessageInfoCaches = make(map[string]map[int32]*proto.MessageInterfaceInfo)
	}
}

// 接口测试覆盖率
func (receiver *StatisticManager) CoverRate() float32 {
	return float32(len(receiver.MasterStatisticMessageInfos)) / float32(receiver.MessageInterfaceCount)
}
