package manager

import (
	model2 "github.com/jzyong/TcpStressTesting/client/model"
	"github.com/jzyong/TcpStressTesting/client/network"
	"github.com/jzyong/golib/util"
	"math"
	"math/rand"
	"sync"
)

// RequestManager 消息请不编排管理
type RequestManager struct {
	util.DefaultModule
}

var requestManager *RequestManager
var requestSingletonOnce sync.Once

func GetRequestManager() *RequestManager {
	requestSingletonOnce.Do(func() {
		requestManager = &RequestManager{}
	})
	return requestManager
}

// Init 初始化
func (m *RequestManager) Init() error {

	return nil
}

// 运行
func (m *RequestManager) Run() {
}

// 关闭
func (m *RequestManager) Stop() {
}

// 登录一次性请求
// 2022-09-05 客户端登录接口请求
func (m *RequestManager) loginRequestOnce(player *model2.Player) {

}

// 定时器 权重请求
func (m *RequestManager) wightRequest(player *model2.Player) {

}

// LoginOtherDataRequest 登录请求其他数据，随机一个定时器请求
func (m *RequestManager) LoginOtherDataRequest(player *model2.Player) {

	m.loginRequestOnce(player)

	m.wightRequest(player)

	//权重随机请求 间隔时间进行随机
	player.AddScheduleJob(util.RandomInt32(3000, 10000), math.MaxUint16, func() {
		random := rand.Int31n(player.WightMax)
		var wight int32 = 0
		for _, job := range player.WightJobs {
			wight += job.Wight
			if random <= wight {
				network.FuncJob(job.Fun).Run()
				break
			}
		}
	})
}
