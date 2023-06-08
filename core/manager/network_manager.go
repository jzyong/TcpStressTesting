package manager

import (
	"context"
	"github.com/jzyong/TcpStressTesting/config"
	"github.com/jzyong/TcpStressTesting/core/proto"
	"github.com/jzyong/golib/log"
	"github.com/jzyong/golib/util"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"sync"
	"time"
)

// 自身网络通信，grpc，http
type NetworkManager struct {
	util.DefaultModule
	WorkerClients    map[string]*RpcClient //工作节点 （master接口拥有）
	WorkerClientList []*RpcClient          //工作节点，用于轮询
	MasterClient     *RpcClient            //连接的 master节点 （worker接口拥有）
	netWorkLock      sync.RWMutex
}

var networkManager *NetworkManager
var netWorkOnce sync.Once

func GetNetworkManager() *NetworkManager {
	netWorkOnce.Do(func() {
		networkManager = &NetworkManager{}
	})
	return networkManager
}

// 开始启动
func (receiver *NetworkManager) Init() error {

	//初始化 master节点
	if config.ApplicationConfigInstance.Master {
		receiver.WorkerClients = make(map[string]*RpcClient, 5)
		receiver.WorkerClientList = make([]*RpcClient, 0, 5)
	}
	log.Info("[网络] 初始化")
	return nil
}

// 开始启动
func (receiver *NetworkManager) Run() {

	GetControlManager().CronScheduler.AddFunc("0/1 * * * * *", receiver.updateSecond)
	GetControlManager().CronScheduler.AddFunc("0/5 * * * * *", receiver.updateFiveSecond)

}

// 关闭连接
func (receiver *NetworkManager) Stop() {
}

// 每秒更新
func (receiver *NetworkManager) updateSecond() {
	// 连接master
	receiver.ConnectMasterSend()
	receiver.deleteWorkerClients()
}

// 每五秒秒更新
func (receiver *NetworkManager) updateFiveSecond() {

}

// 工作客户端个数
func (receiver *NetworkManager) WorkerCount() int {
	return len(receiver.WorkerClients)
}

// 更新worker状态
func (receiver *NetworkManager) UpdateWorkerClient(workerInfo *proto.UploadStatisticsRequest_WorkerServerInfo) {
	receiver.netWorkLock.Lock()
	defer receiver.netWorkLock.Unlock()
	client := receiver.WorkerClients[workerInfo.Name]
	if client != nil {
		client.MemoryUsePercent = workerInfo.MemorySize
		client.CpuUsePercent = workerInfo.CpuRate
		client.UserCount = workerInfo.PlayerCount
		client.HeartTime = util.CurrentTimeSecond()
	}
}

// 根据心跳删除连接断开worker
func (receiver *NetworkManager) deleteWorkerClients() {
	nowSecond := util.CurrentTimeSecond()
	receiver.netWorkLock.Lock()
	defer receiver.netWorkLock.Unlock()
	var deleteClient *RpcClient
	for k, c := range receiver.WorkerClients {
		if (nowSecond - c.HeartTime) > 5 {
			deleteClient = c
			delete(receiver.WorkerClients, k)
		}
	}
	if deleteClient != nil {
		list := make([]*RpcClient, 0, 5)
		for _, c := range receiver.WorkerClientList {
			if c.Name != deleteClient.Name {
				list = append(list, c)
			}
		}
		receiver.WorkerClientList = list
		log.Info("%v 节点移除", deleteClient.Name)
	}
}

// worker 请求和 master 创建连接请求，
// 心跳检测通过集群间上传统计信息失败来检测 worker --> master
func (receiver *NetworkManager) ConnectMasterSend() bool {
	//master节点拥有work节点熟悉，因此允许自己连接自己
	if receiver.MasterClient != nil {
		return false
	}
	masterHost := config.ApplicationConfigInstance.MasterHost
	dialOption := grpc.WithTransportCredentials(insecure.NewCredentials())
	rpcClientConn, err := grpc.Dial(masterHost, dialOption)
	if err != nil {
		log.Error("连接master异常： %v", err)
		return false
	}

	masterClient := &RpcClient{
		RpcHost:    masterHost,
		Name:       config.ApplicationConfigInstance.Name,
		GrpcClient: rpcClientConn,
	}
	client := proto.NewStressTestingInnerServiceClient(rpcClientConn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*1)
	defer cancel()
	request := &proto.ConnectMasterRequest{
		RpcHost: config.ApplicationConfigInstance.RpcHost,
		Name:    config.ApplicationConfigInstance.Name,
	}
	response, err := client.ConnectMaster(ctx, request)
	if err != nil {
		log.Warn("连接master返回异常： %v", err)
		return false
	}

	if response.Status != 0 {
		log.Warn("连接master 状态异常：%v ", response.Status)
		return false
	}
	receiver.MasterClient = masterClient
	log.Info("%v 连接master：%v成功", request.RpcHost, masterHost)
	return true

}

// master 收到 worker的连接请求 worker --> master
func (receiver *NetworkManager) ConnectMasterReceive(host, name string) bool {
	dialOption := grpc.WithTransportCredentials(insecure.NewCredentials())
	rpcClientConn, err := grpc.Dial(host, dialOption)
	if err != nil {
		log.Error("连接worker %v - %v 异常： %v", name, host, err)
		return false
	}
	client := &RpcClient{
		RpcHost:    host,
		Name:       name,
		GrpcClient: rpcClientConn,
		HeartTime:  util.CurrentTimeSecond(),
	}
	log.Info("收到 worker %v - %v 的连接", name, host)
	receiver.netWorkLock.Lock()
	defer receiver.netWorkLock.Unlock()
	receiver.WorkerClients[name] = client
	receiver.WorkerClientList = append(receiver.WorkerClientList, client)
	return true
}

// 所有worker登录用户数
func (receiver *NetworkManager) LoginUserCount() int32 {
	var count int32 = 0
	for _, c := range receiver.WorkerClients {
		count += c.UserCount
	}
	return count
}

// 工作节点
type RpcClient struct {
	RpcHost          string  //rpc连接地址
	Name             string  //名称
	UserCount        int32   //用户数
	MemoryUsePercent int32   //内存百分比
	CpuUsePercent    float64 //CPU使用百分比
	HeartTime        int64   //心跳，用于检测worker是否存活
	GrpcClient       *grpc.ClientConn
}
