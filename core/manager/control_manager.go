package manager

import (
	"context"
	"fmt"
	"github.com/jzyong/TcpStressTesting/config"
	"github.com/jzyong/TcpStressTesting/core/model"
	"github.com/jzyong/TcpStressTesting/core/proto"
	"github.com/jzyong/golib/log"
	"github.com/jzyong/golib/util"
	"github.com/robfig/cron"
	"math/rand"
	"strings"
	"sync"
	"time"
)

// ControlManager 控制器 ，master节点 进行work节点的压测管理，处理来着管理员的操控
type ControlManager struct {
	util.DefaultModule
	TestState        int32         //测试状态
	WorkerLoginIndex int32         //分发登录索引 ，采用轮询worker进行登录
	LoginChannel     chan []string // 玩家登录chan
	CronScheduler    *cron.Cron    //定时器
}

var controlManager *ControlManager
var singletonOnce sync.Once

func GetControlManager() *ControlManager {
	singletonOnce.Do(func() {
		controlManager = &ControlManager{}
	})
	return controlManager
}

// 开始启动
func (receiver *ControlManager) Init() error {
	receiver.LoginChannel = make(chan []string, 1024)

	//初始化定时器
	receiver.CronScheduler = cron.New()
	receiver.CronScheduler.Start()

	log.Info("[控制器] 初始化")
	return nil
}

// 开始启动
func (receiver *ControlManager) Run() {
	receiver.CronScheduler.AddFunc("0/1 * * * * *", receiver.updateSecond)
}

// 关闭连接
func (receiver *ControlManager) Stop() {
}

// 每秒更新
func (receiver *ControlManager) updateSecond() {
}

// 开启测试 后台请求
// return 0成功，1服务器正在运行中，2非master节点
func (receiver *ControlManager) StartTestSend(serverHosts string, slotsId, playerCount, spawnRate int32) (int32, string) {
	if !config.ApplicationConfigInstance.Master {
		return 2, "请求非主节点"
	}
	if receiver.TestState == model.TestRunning {
		return 1, "正在压力测试中，请先暂停之前的测试"
	}
	log.Info("请求测试：服务器 %v 类型 %v 总玩家数 %v 登录速率 %v", serverHosts, slotsId, playerCount, spawnRate)

	if len(serverHosts) == 0 {
		return 3, "请求服务器地址为空"
	}
	gateUrls := strings.Split(serverHosts, ",")
	for _, gateUrl := range gateUrls {
		ipPort := strings.Split(gateUrl, ":")
		if len(ipPort) != 2 {
			return 3, fmt.Sprintf("请求地址格式错误：%v", serverHosts)
		}
	}

	workerCount := int32(GetNetworkManager().WorkerCount())
	if playerCount < 1 || playerCount/workerCount > 5000 {
		return 4, fmt.Sprintf("请求人数%v 不合适，每个工作节点最多运行5000人", playerCount)
	}
	if slotsId < 0 {
		return 5, "测试类型错误，0为常规测试，其他为游戏id"
	}

	config.ApplicationConfigInstance.UserCount = playerCount
	config.ApplicationConfigInstance.SpawnRate = spawnRate
	config.ApplicationConfigInstance.TestType = slotsId
	config.ApplicationConfigInstance.GateUrls = gateUrls
	GetStatisticManager().ResetStatisticData()
	receiver.TestState = model.TestRunning //此处设置只对master节点管用，worker还需要设置

	return 0, "开始压力测试"
}

// 停止压力测试 backend-->master
func (receiver *ControlManager) StopTestSend() int32 {
	if receiver.TestState == model.TestIdle {
		log.Info("压力测试已经结束，重复请求")
		return 0
	}

	////打印未测试接口
	//for k, v := range hall.MID_name {
	//	info := GetStatisticManager().MasterStatisticMessageInfos[k]
	//	if info != nil {
	//		continue
	//	}
	//	log.Debug("%v - %v 未测试", k, v)
	//}

	// 广播worker 停止压力测试
	logoutRequest := &proto.PlayerAllQuitRequest{}
	for _, c := range GetNetworkManager().WorkerClientList {
		go func(client *RpcClient) {
			rpcClient := proto.NewStressTestingInnerServiceClient(client.GrpcClient)
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
			defer cancel()
			response, err := rpcClient.PlayerAllQuit(ctx, logoutRequest)
			if err != nil {
				log.Error("%v 退出压测错误： %v", client.Name, err)
				return
			}
			if response.Status != 0 {
				log.Warn("%v ：暂停压测失败：%v", client.Name, response.Status)
			}
		}(c)
	}
	return 0
}

// PlayerLoginReceive 玩家登录 master请求
// 收到登录 放入channel中，PlayerManager 从channel中依次取出来进行登录
func (receiver *ControlManager) PlayerLoginReceive(account, gateUrl string, testType int32) {
	receiver.TestState = model.TestRunning
	log.Info("收到登录玩家：%v", account)
	loginInfo := make([]string, 0, 3)
	loginInfo = append(loginInfo, account, fmt.Sprintf("%v", testType), gateUrl)
	GetStatisticManager().UserCount++
	receiver.LoginChannel <- loginInfo
}

// 分发玩家登录 ,批量登录
func (receiver *ControlManager) DistributePlayerLogin(accounts []string) {

	workerLoginInfos := make(map[*RpcClient][]*proto.PlayerLoginRequest_LoginInfo, GetNetworkManager().WorkerCount())

	//分配worker
	for _, account := range accounts {
		clients := GetNetworkManager().WorkerClientList
		client := clients[receiver.WorkerLoginIndex]
		receiver.WorkerLoginIndex++
		receiver.WorkerLoginIndex = receiver.WorkerLoginIndex % int32(GetNetworkManager().WorkerCount())
		loginInfos := workerLoginInfos[client]
		if loginInfos == nil {
			loginInfos = make([]*proto.PlayerLoginRequest_LoginInfo, 0, (len(accounts)/GetNetworkManager().WorkerCount())+1)
		}
		gateUrls := config.ApplicationConfigInstance.GateUrls
		gateUrl := gateUrls[rand.Intn(len(gateUrls))]
		loginInfo := &proto.PlayerLoginRequest_LoginInfo{
			Account:  account,
			TestType: config.ApplicationConfigInstance.TestType,
			GateUrl:  gateUrl,
		}
		loginInfos = append(loginInfos, loginInfo)
		workerLoginInfos[client] = loginInfos
	}

	//批量登录
	for client, loginInfos := range workerLoginInfos {
		go func(info []*proto.PlayerLoginRequest_LoginInfo, rpcClient *RpcClient) {
			c := proto.NewStressTestingInnerServiceClient(rpcClient.GrpcClient)
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
			defer cancel()
			_, err := c.PlayerLogin(ctx, &proto.PlayerLoginRequest{LoginInfo: info})
			if err != nil {
				log.Warn("%v ：登录 %v 失败：%v", info, rpcClient.Name, err)
			}
			//log.Debug("%v :登录结果：%v", a, response.Status)
		}(loginInfos, client)
	}
}

// PlayerAllQuitReceived 玩家全部退出压力测试 master-->worker  0成功,1已经退出压测
func (receiver *ControlManager) PlayerAllQuitReceived() int32 {
	if receiver.TestState == model.TestIdle {
		log.Info("已经退出压测，重复请求")
		return 1
	}
	//只设置状态，client自己定时 获取状态进行退出
	receiver.TestState = model.TestQuit
	receiver.WorkerLoginIndex = 0
	GetStatisticManager().ResetStatisticData()
	log.Info("请求暂停压测")
	return 0
}
