package rpc

import (
	"context"
	"github.com/jzyong/TcpStressTesting/core/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"testing"
	"time"
)

const (
	rpcHosts = "127.0.0.1:5010"
	GateUrls = "127.0.0.1:7070"
)

// 开启测试
func TestStartTest(t *testing.T) {
	dialOption := grpc.WithTransportCredentials(insecure.NewCredentials())
	conn, err := grpc.Dial(rpcHosts, dialOption)
	if err != nil {
		log.Fatalf("%v", err)
	}
	defer conn.Close()

	client := proto.NewStressTestingOuterServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	response, err := client.StartTest(ctx, &proto.StartTestRequest{
		ServerHosts: GateUrls,
		SpawnRate:   1,
		PlayerCount: 3,
		TestType:    0,
	})

	if err != nil {
		log.Fatalf("%v", err)
	}
	log.Printf("压测开始返回结果:%v", response)
}

// 停止测试
func TestStopTest(t *testing.T) {
	dialOption := grpc.WithTransportCredentials(insecure.NewCredentials())
	conn, err := grpc.Dial(rpcHosts, dialOption)
	if err != nil {
		log.Fatalf("%v", err)
	}
	defer conn.Close()

	client := proto.NewStressTestingOuterServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	response, err := client.StopTest(ctx, &proto.StopTestRequest{})

	if err != nil {
		log.Fatalf("%v", err)
	}
	log.Printf("压测结束返回结果:%v", response)
}
