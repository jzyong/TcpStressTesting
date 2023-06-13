package rpc

import (
	"context"
	"github.com/jzyong/TcpStressTesting/core/manager"
	"github.com/jzyong/TcpStressTesting/core/model"
	"github.com/jzyong/TcpStressTesting/core/proto"
	"github.com/jzyong/golib/util"
	"time"
)

// OuterRpcService 外部rpc请求，逻辑不写在这里面，后面可能会http请求
type OuterRpcService struct {
	proto.UnimplementedStressTestingOuterServiceServer
}

// StartTest 启动测试
func (service *OuterRpcService) StartTest(ctx context.Context, request *proto.StartTestRequest) (*proto.StartTestResponse, error) {
	status, result := manager.GetControlManager().StartTestSend(request.GetServerHosts(), request.GetTestType(), request.GetPlayerCount(), request.GetSpawnRate())
	response := &proto.StartTestResponse{
		Status: status,
		Result: result,
	}
	return response, nil
}

// StopTest 停止压力测试
func (service *OuterRpcService) StopTest(ctx context.Context, request *proto.StopTestRequest) (*proto.StopTestResponse, error) {
	status := manager.GetControlManager().StopTestSend()
	response := &proto.StopTestResponse{
		Status: status,
	}
	return response, nil
}

// RequestInterfaceInfo 请求接口信息
func (service *OuterRpcService) RequestInterfaceInfo(ctx context.Context, request *proto.RequestInterfaceRequest) (*proto.RequestInterfaceResponse, error) {

	response := &proto.RequestInterfaceResponse{}

	if manager.GetControlManager().TestState == model.TestRunning {
		infos := manager.GetStatisticManager().MasterStatisticMessageInfos
		list := make([]*proto.RequestInterfaceResponse_MessageInterfaceInfo, 0, len(infos))
		for _, info := range infos {
			if info.IsRequestMessage() {
				pastSecond := info.PastSecond()
				interfaceInfo := &proto.RequestInterfaceResponse_MessageInterfaceInfo{
					MessageName:     manager.GetStatisticManager().MessageNameFun(info.MessageId),
					MessageId:       info.MessageId,
					RequestCount:    info.RequestCount(),
					FailCount:       info.FailCount,
					DelayAverage:    info.DelayAverage(),
					DelayMin:        int32(info.DelayMin / int64(time.Millisecond)),
					DelayMax:        int32(info.DelayMax / int64(time.Millisecond)),
					SizeAverage:     info.SizeAverage(),
					Rps:             info.Rps(pastSecond),
					FailSecondCount: info.FailRps(pastSecond),
				}
				list = append(list, interfaceInfo)
			}
		}
		response.InterfaceInfo = list
	}

	return response, nil

}

// 推送接口信息
func (service *OuterRpcService) PushInterfaceInfo(ctx context.Context, request *proto.PushInterfaceRequest) (*proto.PushInterfaceResponse, error) {
	response := &proto.PushInterfaceResponse{}
	if manager.GetControlManager().TestState == model.TestRunning {
		infos := manager.GetStatisticManager().MasterStatisticMessageInfos
		list := make([]*proto.PushInterfaceResponse_MessageInterfaceInfo, 0, len(infos))
		for _, info := range infos {
			if !info.IsRequestMessage() {
				interfaceInfo := &proto.PushInterfaceResponse_MessageInterfaceInfo{
					MessageName: manager.GetStatisticManager().MessageNameFun(info.MessageId),
					Count:       info.PushCount,
					SizeAverage: info.SizeAverage(),
					Rps:         info.PushRps(0),
					MessageId:   info.MessageId,
				}
				list = append(list, interfaceInfo)
			}
		}
		response.InterfaceInfo = list
	}

	return response, nil

}

func (service *OuterRpcService) StatisticsLog(ctx context.Context, request *proto.StatisticsLogRequest) (*proto.StatisticsLogResponse, error) {
	state := manager.GetControlManager().TestState
	response := &proto.StatisticsLogResponse{Status: state}
	e := manager.GetStatisticManager().StatisticLogs.Back()
	if e != nil {
		info := e.Value.(*proto.MessageInterfaceInfo)
		pastSecond := info.PastSecond()
		log := &proto.StatisticLog{
			TimeStamp:        int32(util.CurrentTimeSecond()),
			Rps:              info.Rps(pastSecond),
			FailRps:          info.FailRps(pastSecond),
			PushRps:          info.PushRps(pastSecond),
			ResponseTime:     info.DelayAverage(),
			PlayerCount:      manager.GetNetworkManager().LoginUserCount(),
			RequestBytes:     info.RequestFlow(),
			ResponseBytes:    info.ResponseFlow(),
			RequestByteRate:  info.RequestFlowRate(pastSecond),
			ResponseByteRate: info.ResponseFlowRate(pastSecond),
			RequestCount:     info.RequestCount(),
			PushCount:        info.PushCount,
			FailCount:        info.FailCount,
			CoverRate:        manager.GetStatisticManager().CoverRate(),
		}
		response.StatisticLog = log
	}
	return response, nil
}

// WorkerServerInfo 获取worker信息
func (service *OuterRpcService) WorkerServerInfo(ctx context.Context, request *proto.WorkerServerInfoRequest) (*proto.WorkerServerInfoResponse, error) {
	workerServerInfos := make([]*proto.WorkerServerInfoResponse_WorkerServerInfo, 0, 3)
	clients := manager.GetNetworkManager().WorkerClientList
	for _, c := range clients {
		info := &proto.WorkerServerInfoResponse_WorkerServerInfo{
			Name:        c.Name,
			RpcHost:     c.RpcHost,
			CpuRate:     c.CpuUsePercent,
			MemorySize:  c.MemoryUsePercent,
			PlayerCount: c.UserCount,
		}
		workerServerInfos = append(workerServerInfos, info)
	}

	response := &proto.WorkerServerInfoResponse{WorkerServer: workerServerInfos}
	//log.Debug("服务器信息：%v", response)
	return response, nil
}
