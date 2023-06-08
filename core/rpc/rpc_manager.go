package rpc

import (
	"fmt"
	"github.com/jzyong/TcpStressTesting/config"
	"github.com/jzyong/TcpStressTesting/core/proto"
	"github.com/jzyong/golib/log"
	"github.com/jzyong/golib/util"
	"google.golang.org/grpc"
	"net"
	"strings"
)

// grpc 管理
type GRpcManager struct {
	*util.DefaultModule
	GrpcServer *grpc.Server
}

var grpcManager = &GRpcManager{}

func GetGrpcManager() *GRpcManager {
	return grpcManager
}

func (m *GRpcManager) Init() error {
	server := grpc.NewServer()
	m.GrpcServer = server
	// 添加grpc服务
	proto.RegisterStressTestingInnerServiceServer(server, new(InnerRpcService))
	proto.RegisterStressTestingOuterServiceServer(server, new(OuterRpcService))
	//容器中运行 绑定ip地址不一定正确走配置
	portStr := strings.Split(config.ApplicationConfigInstance.RpcHost, ":")[1]
	listen, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%v", portStr))
	if err != nil {
		log.Fatal("%v", err)
	}
	log.Info("grpc listen on:%v", config.ApplicationConfigInstance.RpcHost)
	go server.Serve(listen)
	log.Info("[Grpc] 初始化")

	return nil
}

// 开始启动
func (m *GRpcManager) Run() {
}

func (m *GRpcManager) Stop() {
	if m.GrpcServer != nil {
		m.GrpcServer.Stop()
	}
}
