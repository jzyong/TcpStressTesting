package handler

import (
	"github.com/jzyong/TcpStressTesting/client/manager"
	"github.com/jzyong/TcpStressTesting/client/message"
	"github.com/jzyong/TcpStressTesting/client/network"
)

// RegisterHandlers 注册消息
func RegisterHandlers() {
	//玩家
	messageDistribute := manager.GetPlayerManager().MessageDistribute
	messageDistribute.RegisterHandler(int32(message.MID_UserLoginRes), network.NewTcpHandler(UserLoginRes))
	messageDistribute.RegisterHandler(int32(message.MID_ReconnectRes), network.NewTcpHandler(ReconnectRes))
	messageDistribute.RegisterHandler(int32(message.MID_HeartRes), network.NewTcpHandler(HeartRes))

}
