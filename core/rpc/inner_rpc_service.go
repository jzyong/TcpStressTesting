package rpc

import (
	"context"
	"github.com/jzyong/TcpStressTesting/core/manager"
	"github.com/jzyong/TcpStressTesting/core/proto"
)

// 内部rpc请求
type InnerRpcService struct {
	proto.UnimplementedStressTestingInnerServiceServer
}

// 连接master worker-->master
func (service *InnerRpcService) ConnectMaster(ctx context.Context, request *proto.ConnectMasterRequest) (*proto.ConnectMasterResponse, error) {
	success := manager.GetNetworkManager().ConnectMasterReceive(request.RpcHost, request.Name)
	var status int32 = 0
	if !success {
		status = 1
	}
	var response = &proto.ConnectMasterResponse{
		Status: status,
	}
	return response, nil
}

// 玩家登录  master-->worker
func (service *InnerRpcService) PlayerLogin(ctx context.Context, request *proto.PlayerLoginRequest) (*proto.PlayerLoginResponse, error) {
	// 收到登录 放入channel中，PlayerManager 从channel中依次取出来进行登录
	for _, loginInfo := range request.LoginInfo {
		manager.GetControlManager().PlayerLoginReceive(loginInfo.Account, loginInfo.GateUrl, loginInfo.TestType)
	}
	response := &proto.PlayerLoginResponse{Status: 0}
	return response, nil
}

// 退出所有玩家
func (service *InnerRpcService) PlayerAllQuit(ctx context.Context, request *proto.PlayerAllQuitRequest) (*proto.PlayerAllQuitResponse, error) {
	status := manager.GetControlManager().PlayerAllQuitReceived()
	response := &proto.PlayerAllQuitResponse{Status: status}
	return response, nil
}

// 处理上传的数据
func (service *InnerRpcService) UploadStatistics(ctx context.Context, request *proto.UploadStatisticsRequest) (*proto.UploadStatisticsResponse, error) {
	// 进行统计数据处理
	manager.GetStatisticManager().UploadStatisticReceive(request)
	response := &proto.UploadStatisticsResponse{
		Status: 0,
	}

	return response, nil

}
