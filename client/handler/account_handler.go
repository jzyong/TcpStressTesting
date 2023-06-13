package handler

import (
	"github.com/jzyong/TcpStressTesting/client/manager"
	"github.com/jzyong/TcpStressTesting/client/message"
	"github.com/jzyong/TcpStressTesting/client/network"
	"github.com/jzyong/golib/log"
	"google.golang.org/protobuf/proto"
	"math"
)

// UserLoginRes 用户登录
func UserLoginRes(msg network.TcpMessage) bool {
	player := manager.GetPlayerManager().GetPlayerByMsg(msg)
	if player == nil {
		log.Warn("玩家初始化异常")
		return false
	}
	response := &message.UserLoginResponse{}
	proto.Unmarshal(msg.GetData(), response)
	log.Info("用户%v 登录成功userId=%v", player.Imei, response.GetUserId())
	player.Id = response.UserId
	manager.GetPlayerManager().AddPlayer(player)

	//定时发送心跳
	player.AddScheduleJob(3000, math.MaxUint16, func() {
		manager.SendMessage(player, message.MID_HeartReq, &message.HeartRequest{})
	})

	return true
}

// ReconnectRes 重连
func ReconnectRes(msg network.TcpMessage) bool {

	return true
}

// HeartRes 心跳
func HeartRes(msg network.TcpMessage) bool {
	player := manager.GetPlayerManager().GetPlayerByMsg(msg)
	if player == nil {
		log.Warn("玩家初始化异常")
		return false
	}
	log.Debug("%v 收到心跳消息", player.Id)
	return true
}
