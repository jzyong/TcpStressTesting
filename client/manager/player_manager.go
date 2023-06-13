package manager

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/jzyong/TcpStressTesting/client/message"
	model2 "github.com/jzyong/TcpStressTesting/client/model"
	network "github.com/jzyong/TcpStressTesting/client/network"
	"github.com/jzyong/TcpStressTesting/config"
	"github.com/jzyong/TcpStressTesting/core/manager"
	"github.com/jzyong/TcpStressTesting/core/model"
	proto2 "github.com/jzyong/TcpStressTesting/core/proto"
	"github.com/jzyong/golib/log"
	"github.com/jzyong/golib/util"
	"runtime"
	"sync"
)

// 客户端
type PlayerManager struct {
	util.DefaultModule
	players           sync.Map                  //map[int64]*model2.Player  //所有玩家
	imeiPlayers       sync.Map                  //map[string]*model2.Player //所有玩家
	MessageDistribute network.MessageDistribute //消息分发
	loginCount        int32                     //已经登录人数
}

var playerManager *PlayerManager
var playerSingletonOnce sync.Once

func GetPlayerManager() *PlayerManager {
	playerSingletonOnce.Do(func() {
		playerManager = &PlayerManager{}
	})
	return playerManager
}

// Init 初始化
func (m *PlayerManager) Init() error {

	log.Info("[玩家] 初始化")
	m.MessageDistribute = network.NewMessageDistribute(uint32(runtime.NumCPU()), nil, handlerExecuteFinish)
	m.MessageDistribute.StartWorkerPool()
	return nil
}

// Run 运行
func (m *PlayerManager) Run() {
	manager.GetControlManager().CronScheduler.AddFunc("@every 1s", m.updateSecond)
	manager.GetControlManager().CronScheduler.AddFunc("@every 5s", m.updateFiveSecond)

	//接收分发的登录信息
	go m.playerLoginRun()

	//设置总压测消息个数
	manager.GetStatisticManager().MessageInterfaceCount = len(message.MID_name)
}

// Stop 关闭
func (m *PlayerManager) Stop() {
}

// 每秒更新
// 进行玩家定时任务检测，执行转发到对应的chan中
func (m *PlayerManager) updateSecond() {

	//master 批量 登录玩家
	m.batchPlayerLogin()

	//批量退出
	m.batchPlayerLogout()

	nowTime := util.CurrentTimeMillisecond()

	m.players.Range(func(_, value any) bool {
		player := value.(*model2.Player)
		if !player.TcpClient.GetChannel().IsClose() {
			m.playerSecondUpdate(player, nowTime)
		}
		return true
	})
}

// 每秒更新
// 进行玩家定时任务检测，执行转发到对应的chan中
func (m *PlayerManager) updateFiveSecond() {
	m.players.Range(func(key, value any) bool {
		player := value.(*model2.Player)
		if player.TcpClient.GetChannel().IsClose() {
			// 进行断线重连，离线服务器可能踢掉死玩家，如一直不给网关或大厅发送消息
			// 先直接从走登录流程，后面优化走断线重连
			loginInf := make([]string, 0, 3)
			loginInf = append(loginInf, player.Imei, fmt.Sprintf("%v", player.TestType), player.GateUrl)
			log.Debug("%v %v %v 重连进入游戏", player.Id, player.Imei, player.GateUrl)
			m.playerLogin(loginInf)
		} else {
			m.playerFiveSecondUpdate(player)
		}
		return true
	})
}

// 玩家每5秒执行逻辑
func (m *PlayerManager) playerFiveSecondUpdate(player *model2.Player) {
	m.MessageDistribute.ExecuteFunc(player.Id, func() {
		//发送心跳

		now := util.Now().UnixNano()
		//检测 请求失败消息
		for _, info := range player.StatisticMessage {
			if (now - info.StartTime) > model.MessageRequestFailTime {
				info.ExecuteTime = now - info.StartTime
				//将消息发送的 StatisticManager进行统计
				player.RemoveMessageInfo(info.SequenceNo)
				manager.GetStatisticManager().PlayerMessageInfoChan <- info
			}
		}

	})
}

// 玩家每秒执行逻辑
func (m *PlayerManager) playerSecondUpdate(player *model2.Player, nowTime int64) {

	//非线程安全
	//定时任务 检测
	if player.ScheduleJobs != nil {
		m.MessageDistribute.ExecuteFunc(player.Id, func() {
			deleteIndex := -1
			for i, job := range player.ScheduleJobs {
				remainCount := job.Execute(m.MessageDistribute, nowTime, player.Id)
				if remainCount < 1 {
					deleteIndex = i
					//log.Debug("%v 删除定时任务%v", player.Id, job.IntervalTime)
				}
			}
			if deleteIndex >= 0 {
				player.ScheduleJobs = append(player.ScheduleJobs[:deleteIndex], player.ScheduleJobs[deleteIndex+1:]...)
			}
		})
	}
}

// 通过设备id登录，每次都发注册消息，本地不保存账号密码 后台请求执行
func (m *PlayerManager) batchPlayerLogin() {
	//未开始压测
	if manager.GetControlManager().TestState != model.TestRunning {
		return
	}
	if !config.ApplicationConfigInstance.Master {
		return
	}

	//玩家已经完全登录
	if m.loginCount >= config.ApplicationConfigInstance.UserCount {
		return
	}

	spawnRate := config.ApplicationConfigInstance.SpawnRate
	accounts := make([]string, 0, spawnRate)
	for i := 0; i < int(spawnRate); i++ {
		m.loginCount++
		accounts = append(accounts, fmt.Sprintf("%v_%v", model2.PlayerImeiPrefix, m.loginCount))
		//防止多登录
		if m.loginCount >= config.ApplicationConfigInstance.UserCount {
			break
		}
	}
	manager.GetControlManager().DistributePlayerLogin(accounts)
}

// 批量退出
func (m *PlayerManager) batchPlayerLogout() {
	//未开始压测
	if manager.GetControlManager().TestState != model.TestQuit {
		return
	}

	var playerCount = 0
	m.players.Range(func(key, value any) bool {
		player := value.(*model2.Player)
		log.Debug("%v 退出游戏", player.Id)
		player.TcpClient.GetChannel().Stop()
		player.TcpClient.Stop()
		playerCount++
		return true
	})
	m.players = sync.Map{}
	m.imeiPlayers = sync.Map{}
	m.loginCount = 0
	manager.GetStatisticManager().ResetStatisticData() //重置一下统计
	log.Info("退出玩家数量：%v", playerCount)
	manager.GetControlManager().TestState = model.TestIdle
}

// 玩家登录 登录来着master 分配的玩家，从chan中取出玩家
func (m *PlayerManager) playerLoginRun() {
	for {
		select {
		case loginInfo, ok := <-manager.GetControlManager().LoginChannel:
			if ok {
				m.playerLogin(loginInfo)
			} else {
				log.Warn("LoginChannel 已经关闭")
				break
			}
		}
	}
}

// 玩家登录 登录来着master 分配的玩家，创建socket，进行账号获取，账号登陆
func (m *PlayerManager) playerLogin(loginInfo []string) {
	imei := loginInfo[0]
	testType := util.ParseInt32(loginInfo[1])
	gateUrl := loginInfo[2]
	tcpClient, _ := network.NewTcpClient(imei, gateUrl, m.MessageDistribute)
	tcpClient.SetChannelInactive(clientChannelInactive)
	tcpClient.SetChannelActive(clientChannelActive)
	go tcpClient.Start()

	player := &model2.Player{
		Imei:             imei,
		TestType:         testType,
		TcpClient:        tcpClient,
		GateUrl:          gateUrl,
		StatisticMessage: make(map[int32]*proto2.PlayerMessageInfo),
		ScheduleJobs:     make([]*network.ScheduleJob, 0, 10),
		WightJobs:        make([]*model2.PlayerWightJob, 0, 30),
	}
	m.imeiPlayers.Store(imei, player)

	//log.Debug("%v 开始登录", account)
}

// AddPlayer 添加玩家
func (m *PlayerManager) AddPlayer(player *model2.Player) {
	m.players.Store(player.Id, player)
}

// 移除玩家
func (m *PlayerManager) removePlayer(player *model2.Player) {
	m.players.Delete(player.Id)
}

// 获取玩家
func (m *PlayerManager) GetPlayer(id int64) *model2.Player {
	p, ok := m.players.Load(id)
	if ok {
		return p.(*model2.Player)
	}
	return nil
}

// GetPlayerByImei 获取玩家 通过设备号
func (m *PlayerManager) GetPlayerByImei(imei string) *model2.Player {
	p, ok := m.imeiPlayers.Load(imei)
	if ok {
		return p.(*model2.Player)
	}
	return nil
}

// GetPlayerByMsg 根据服务器返回消息获取玩家
func (m *PlayerManager) GetPlayerByMsg(msg network.TcpMessage) *model2.Player {
	imeiObject, _ := msg.GetTcpChannel().GetProperty(network.TcpName)
	imei := imeiObject.(string)
	player := m.GetPlayerByImei(imei)
	return player
}

// SendMessage 发送消息 （所有消息必须从此次发送，添加序列号）
func SendMessage(player *model2.Player, mid message.MID, message proto.Message) {
	messageId := int32(mid)
	if player.TcpClient == nil || player.TcpClient.GetChannel() == nil {
		log.Info("%v-%v 未创建socket连接 消息:%v 发送失败", player.Id, player.Imei, messageId)
		return
	}
	c := player.TcpClient.GetChannel()
	if c.IsClose() {
		log.Info("%v-%v channel已关闭 消息:%v 发送失败", player.Id, player.Imei, messageId)
		return
	}
	data, err := proto.Marshal(message)
	if err != nil {
		log.Warn("%v Pack error msg id = %v  %v", player.Id, messageId, err)
		return
	}

	//封装数据，未使用dataPack，减少创建两个对象
	//创建一个存放bytes字节的缓冲
	dataBuff := bytes.NewBuffer([]byte{})
	//写dataLen 不包含自身长度
	length := 12 + uint32(len(data))
	if err := binary.Write(dataBuff, binary.LittleEndian, length); err != nil {
		log.Info("%v-%v channel已关闭 消息:%v 编码错误：%v", player.Id, player.Imei, messageId, err)
		return
	}
	//写msgID
	if err := binary.Write(dataBuff, binary.LittleEndian, messageId); err != nil {
		log.Info("%v-%v channel已关闭 消息:%v 编码错误：%v", player.Id, player.Imei, messageId, err)
		return
	}
	//写确认序列号
	if err := binary.Write(dataBuff, binary.LittleEndian, player.MessageAck); err != nil {
		log.Info("%v-%v channel已关闭 消息:%v 编码错误：%v", player.Id, player.Imei, messageId, err)
		return
	}
	player.MessageSeq += 1
	//写 序列号
	if err := binary.Write(dataBuff, binary.LittleEndian, player.MessageSeq); err != nil {
		log.Info("%v-%v channel已关闭 消息:%v 编码错误：%v", player.Id, player.Imei, messageId, err)
		return
	}
	//写data数据
	if err := binary.Write(dataBuff, binary.LittleEndian, data); err != nil {
		log.Info("%v-%v channel已关闭 消息:%v 编码错误：%v", player.Id, player.Imei, messageId, err)
		return
	}
	player.AddMessageInfo(messageId, player.MessageSeq, int32(length))
	//dataBuff.Bytes()
	player.TcpClient.GetChannel().GetMsgBuffChan() <- dataBuff.Bytes()
}

// 链接激活
func clientChannelActive(channel network.TcpChannel) {
	//登录
	imeiObject, _ := channel.GetProperty(network.TcpName)
	imei := imeiObject.(string)
	player := GetPlayerManager().GetPlayerByImei(imei)
	if player == nil {
		log.Warn("%v : 获取密钥玩家未找到", imei)
		return
	}
	request := &message.UserLoginRequest{Account: imei, Password: imei, Imei: imei}
	SendMessage(player, message.MID_UserLoginReq, request)

}

// 链接断开
func clientChannelInactive(channel network.TcpChannel) {
	log.Info("网关连接断开：%v", channel.RemoteAddr())
}

// 消息处理完后回调统计协议
func handlerExecuteFinish(msg network.TcpMessage) {
	p := GetPlayerManager().GetPlayerByMsg(msg)
	if p != nil {

		//请求消息
		if msg.GetSeq() > 0 {
			info := p.RemoveMessageInfo(msg.GetSeq())
			p.MessageAck = msg.GetSeq() //直接移除，没有筛选最小的，可能把没收到的消息也移除掉
			if info != nil {
				info.ExecuteTime = util.Now().UnixNano() - info.StartTime
				info.MessageSize = int32(msg.GetLength())
				//将消息发送的 StatisticManager进行统计
				//log.Debug("消息处理：%v %v %v", info,info.ExecuteTime,util.Now().UnixNano())
				manager.GetStatisticManager().PlayerMessageInfoChan <- info
			}
			//主推消息
		} else {
			info := &proto2.PlayerMessageInfo{
				MessageId:   msg.GetMessageId(),
				StartTime:   util.Now().UnixNano(),
				MessageSize: int32(msg.GetLength()),
			}
			manager.GetStatisticManager().PlayerMessageInfoChan <- info
		}

	}
}
