package handler

import (
	"github.com/jzyong/TcpStressTesting/client/network"
)

// UserLoginRes 用户登录
func UserLoginRes(msg network.TcpMessage) bool {

	return true
}

// ReconnectRes 重连
func ReconnectRes(msg network.TcpMessage) bool {

	return true
}

// HeartRes 心跳
func HeartRes(msg network.TcpMessage) bool {

	return true
}
