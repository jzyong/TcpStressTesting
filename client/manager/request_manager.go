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
	//关卡
	delayTime := util.RandomInt32(1000, 60000)
	GetCustomPassManager().CustomsPassListRequest(player, delayTime)

	//获取在线服务器列表
	delayTime = util.RandomInt32(100, 3000)
	player.AddScheduleJob(delayTime, 1, func() {
		SendMessage(player, hall.MID_GameOnlineListReqMid, &hall.Req3000807{})
	})

	//好友列表
	delayTime = util.RandomInt32(100, 10000)
	player.AddScheduleJob(delayTime, 1, func() {
		GetFriendManager().FriendListRequest(player)
		GetFriendManager().RecommendFriendListRequest(player)
		//卡包列表
		SendMessage(player, hall.MID_GiftListReqMid, &hall.Req3000073{})
	})

	//任务
	delayTime = util.RandomInt32(100, 5000)
	player.AddScheduleJob(delayTime, 3, func() {
		//每日任务列表
		SendMessage(player, hall.MID_TaskListReqMid, &hall.Req3000030{})
	})

	//乐透
	delayTime = util.RandomInt32(100, 5000)
	player.AddScheduleJob(delayTime, 1, func() {
		SendMessage(player, hall.MID_EnterNewLotteryReqMid, &hall.Req3001001{})
	})

	//商城
	delayTime = util.RandomInt32(100, 5000)
	player.AddScheduleJob(delayTime, 1, func() {
		//商城配置
		SendMessage(player, hall.MID_GetShopConfigReqMid, &hall.Req3000011{})
		//商城二次购第一次道具
		SendMessage(player, hall.MID_ShopFirstBuyItemReqMid, &hall.Req3000021{})
		//商城钻石活动
		SendMessage(player, hall.MID_DiamondActivitiesReqMid, &hall.Req3000209{})
	})

	//收集
	delayTime = util.RandomInt32(100, 5000)
	player.AddScheduleJob(delayTime, 1, func() {
		//进入收集大厅
		SendMessage(player, hall.MID_EnterCollectionReqMid, &hall.Req3000701{})
	})

	//俱乐部
	delayTime = util.RandomInt32(100, 5000)
	player.AddScheduleJob(delayTime, 1, func() {
		//俱乐部活动列表
		SendMessage(player, hall.MID_ClubActivityListReqMid, &hall.Req3000236{})
	})
	//聊天
	delayTime = util.RandomInt32(100, 5000)
	player.AddScheduleJob(delayTime, 1, func() {
		//进入聊天
		var rType int32 = 1
		SendMessage(player, hall.MID_EnterChatReqMid, &hall.Req3000901{Type: &rType})
	})

	//轮替游戏  应该服务器主动推送小游戏信息，或者客户端根据当前开放的小游戏进行请求 2022-11-16
	delayTime = util.RandomInt32(100, 5000)
	player.AddScheduleJob(delayTime, 1, func() {
		//跳跳棋 信息
		var requestType int32 = 0
		SendMessage(player, hall.MID_JumpDraughtsReqMid, &hall.Req3000322{Type: &requestType})
		//Bingo 信息
		SendMessage(player, hall.MID_BingoInfoReqMid, &hall.Req3000300{})
		//bingo商城列表
		SendMessage(player, hall.MID_BingoShopListReqMid, &hall.Req3000304{})
		//魔法帽 信息
		SendMessage(player, hall.MID_MagicHatInfoReqMid, &hall.Req3000311{})
		//推金币 信息
		SendMessage(player, hall.MID_PushCoinInfoReqMid, &hall.Req3000315{})
		//进入大富翁
		requestType = 1
		SendMessage(player, hall.MID_EnterDaFuWenReqMid, &hall.Req3000305{Type: &requestType})
	})

	//活动
	delayTime = util.RandomInt32(100, 5000)
	player.AddScheduleJob(delayTime, 1, func() {
		//进入签到
		SendMessage(player, hall.MID_EnterSignReqMid, &hall.Req3000053{})
		//活动面板信息
		var requestType = util.RandomInt32(0, 1)
		SendMessage(player, hall.MID_ActivityPanelListReqMid, &hall.Req3000149{Type: &requestType})
		//获取循环高返利礼包信息
		SendMessage(player, hall.MID_LoopRebateInfoReqMid, &hall.Req3000142{})
		//新人限时礼包
		requestType = util.RandomInt32(1, 2)
		SendMessage(player, hall.MID_NewbieLimitInfoReqMid, &hall.Req3000144{})
		//圣诞活动信息
		SendMessage(player, hall.MID_ChristmasInfoReqMid, &hall.Req3000152{})
		//免费金币
		SendMessage(player, hall.MID_FreeGoldInfoReqMid, &hall.Req3000159{})
		// 热点红点信息
		SendMessage(player, hall.MID_HotNewsReqMid, &hall.Req3000163{})
		//老玩家回归
		SendMessage(player, hall.MID_OldPlayerReturnReqMid, &hall.Req3000168{})
		//金币返利信息
		SendMessage(player, hall.MID_GoldRebateReqMid, &hall.Req3000171{})
		//vegas leagues 当前赛季排行榜排名信息
		SendMessage(player, hall.MID_VegasLeaguesInfoReqMid, &hall.Req3000181{})
		// 进入幸运蛋 遗弃
		//SendMessage(player, hall.MID_EnterLuckyEggReqMid, &hall.Req3000190{})
		//等级成长信息
		SendMessage(player, hall.MID_LevelGrowthReqMid, &hall.Req3000211{})
		//进入存钱罐
		SendMessage(player, hall.MID_EnterPiggyBankReqMid, &hall.Req3000034{})
		//循环礼包
		SendMessage(player, hall.MID_LoopGiftInfoReqMid, &hall.Req3000214{})
		//大额付费礼包充值信息
		SendMessage(player, hall.MID_BigPremiumPackageInfoReqMid, &hall.Req3000213{})
		// 三合一基本信息
		SendMessage(player, hall.MID_ThreeInOneReqMid, &hall.Req3000198{})
		// 进入复活节
		SendMessage(player, hall.MID_EnterEasterReqMid, &hall.Req3000199{})
		//幸运值双倍活动
		SendMessage(player, hall.MID_LuckyActivityReqMid, &hall.Req3000196{})

	})

}

// 定时器 权重请求
func (m *RequestManager) wightRequest(player *model2.Player) {
	// [1,100]底，[101,500]中，[1000,2000]高，

	//============================ slots ============================
	//权重 100
	//此消息不能统计，因为服务器逻辑没有区分请求和推送，没有返回序列号
	player.AddWightJob(100, func() {
		SendMessage(player, hall.MID_HallJackPotDisPlayReqMid, &hall.Req3000058{})
	})

	//============================ 关卡 ============================
	//权重 150
	player.AddWightJob(100, func() {
		GetCustomPassManager().CustomsPassListRequest(player, 0)
	})
	player.AddWightJob(50, func() {
		GetCustomPassManager().CustomsPassRankRequest(player)
	})

	//============================ 玩家 ============================
	//权重 801
	player.AddWightJob(1, func() {
		GetPlayerManager().bindEmailRequest(player)
	})
	player.AddWightJob(500, func() {
		GetPlayerManager().hallInfoRequest(player)
	})
	player.AddWightJob(50, func() {
		//查看道具列表
		SendMessage(player, hall.MID_ItemListReqMid, &hall.Req3000004{})
	})
	player.AddWightJob(50, func() {
		//下一级信息
		SendMessage(player, hall.MID_NextLvInfoReqMid, &hall.Req3000082{})
	})
	player.AddWightJob(200, func() {
		//所有红点
		SendMessage(player, hall.MID_RedPointListReqMid, &hall.Req3000047{})
	})

	//============================ 好友 ============================
	//权重 120
	player.AddWightJob(100, func() {
		GetFriendManager().FriendListRequest(player)
	})
	player.AddWightJob(10, func() {
		GetFriendManager().RecommendFriendListRequest(player)
	})
	player.AddWightJob(10, func() {
		GetFriendManager().FriendApplyListRequest(player)
	})

	//============================ 任务 ============================
	//权重 700
	player.AddWightJob(150, func() {
		//每日任务列表
		SendMessage(player, hall.MID_TaskListReqMid, &hall.Req3000030{})
	})
	player.AddWightJob(150, func() {
		//钻石任务列表
		SendMessage(player, hall.MID_DiamondTaskReqMid, &hall.Req3000128{})
	})
	player.AddWightJob(50, func() {
		//新手列表
		SendMessage(player, hall.MID_NewTaskListReqMid, &hall.Req3000601{})
	})
	player.AddWightJob(50, func() {
		//钻石任务排行榜
		SendMessage(player, hall.MID_DiamondPointRankReqMid, &hall.Req3000133{})
	})
	player.AddWightJob(50, func() {
		//钻石任务商城
		SendMessage(player, hall.MID_DiamondTaskShopReqMid, &hall.Req3000134{})
	})
	player.AddWightJob(150, func() {
		//徽章任务列表
		SendMessage(player, hall.MID_BadgeHomeReqMid, &hall.Req3001021{})
	})
	player.AddWightJob(100, func() {
		//子游戏兑换任务
		SendMessage(player, hall.MID_ExchangeTaskHomeReqMid, &hall.Req3001031{})
	})

	//============================ 邮件 ============================
	//权重 150
	player.AddWightJob(150, func() {
		//邮件列表
		SendMessage(player, hall.MID_EnterMailReqMid, &hall.Req3000015{})
	})

	//============================ 高倍场&建造系统 ============================
	//权重 300
	player.AddWightJob(150, func() {
		//进入贵宾厅
		SendMessage(player, hall.MID_EnterLoungeReqMid, &hall.Req3000811{})
	})
	player.AddWightJob(150, func() {
		//进入城市建造
		SendMessage(player, hall.MID_EnterCityReqMid, &hall.Req3000815{})
	})

	//============================ 乐透 ============================
	//权重 260
	player.AddWightJob(100, func() {
		//乐透信息
		SendMessage(player, hall.MID_EnterNewLotteryReqMid, &hall.Req3001001{})
	})
	player.AddWightJob(10, func() {
		//历史开奖记录
		SendMessage(player, hall.MID_HistoryNewLotteryRecordReqMid, &hall.Req3001003{})
	})
	player.AddWightJob(10, func() {
		//投注历史记录
		SendMessage(player, hall.MID_BetNewLotteryRecordReqMid, &hall.Req3001004{})
	})
	player.AddWightJob(10, func() {
		//开奖结果
		SendMessage(player, hall.MID_ResultNewLotteryReqMid, &hall.Req3001006{})
	})
	player.AddWightJob(10, func() {
		//新乐透商城
		SendMessage(player, hall.MID_ShopNewLotteryReqMid, &hall.Req3001008{})
	})
	player.AddWightJob(10, func() {
		//查询上期中奖玩家信息
		SendMessage(player, hall.MID_LastIssueNewLotteryReqMid, &hall.Req3001009{})
	})
	player.AddWightJob(10, func() {
		//查询当期下注历史
		SendMessage(player, hall.MID_BetNowIssueNewLotteryReqMid, &hall.Req3001010{})
	})

	//============================ 俱乐部 ============================
	//权重 100
	player.AddWightJob(100, func() {
		//进入俱乐部
		var requestType int32 = 1
		SendMessage(player, hall.MID_EnterClubReqMid, &hall.Req3000221{Type: &requestType})
	})

	//============================ 聊天 ============================
	//权重 100
	player.AddWightJob(100, func() {
		//进入俱乐部
		var requestType int32 = 1
		SendMessage(player, hall.MID_EnterChatReqMid, &hall.Req3000901{Type: &requestType})
	})

	//============================ 收集 ============================
	//权重 150
	player.AddWightJob(100, func() {
		//进入收集
		var requestType int32 = 1
		SendMessage(player, hall.MID_EnterCollectionReqMid, &hall.Req3000701{Type: &requestType})
	})
	player.AddWightJob(50, func() {
		//获取任务游戏信息
		SendMessage(player, hall.MID_CollectMissionGameInfoReqMid, &hall.Req3000712{})
	})

	//============================ 小游戏 ============================
	//权重 150
	player.AddWightJob(100, func() {
		GetMiniGameManager().RequestMiniGameInfo(player)
	})
	player.AddWightJob(50, func() {
		//魔法帽
		SendMessage(player, hall.MID_MagicHatInfoReqMid, &hall.Req3000311{})
	})

	//============================ 活动 ============================
	GetActivityManager().wightRequest(player)

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
