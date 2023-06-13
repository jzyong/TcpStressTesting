package handler

import (
	"github.com/jzyong/TcpStressTesting/client/manager"
	"github.com/jzyong/TcpStressTesting/client/message"
	"github.com/jzyong/TcpStressTesting/client/network"
	manager2 "github.com/jzyong/TcpStressTesting/core/manager"
)

// RegisterHandlers 注册消息
func RegisterHandlers() {
	manager2.GetStatisticManager().MessageNameFun = messageName

	//玩家
	messageDistribute := manager.GetPlayerManager().MessageDistribute
	messageDistribute.RegisterHandler(int32(message.MID_UserLoginRes), network.NewTcpHandler(UserLoginRes))
	messageDistribute.RegisterHandler(int32(message.MID_ReconnectRes), network.NewTcpHandler(ReconnectRes))
	messageDistribute.RegisterHandler(int32(message.MID_HeartRes), network.NewTcpHandler(HeartRes))

}

// 获取消息名称
func messageName(messageId int32) string {
	return message.MID_name[messageId]
}
