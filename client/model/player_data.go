package model

import (
	"github.com/jzyong/TcpStressTesting/client/network"
	"github.com/jzyong/TcpStressTesting/core/proto"
	"github.com/jzyong/golib/log"
	"github.com/jzyong/golib/util"
)

// 玩家对象
type Player struct {
	Id               int64                              //玩家id
	Nick             string                             //昵称
	Imel             string                             //设备号
	Account          string                             //账号
	Password         string                             //密码
	MessageSeq       int32                              //消息序列号，自增长
	MessageAck       int32                              //消息确认序号，等服务器返回，获取最小序列号
	GateUrl          string                             //网关地址
	TestType         int32                              //测试类型  0模拟测试
	TcpClient        network.TcpClient                  //玩家tcp连接
	ScheduleJobs     []*network.ScheduleJob             //调度任务
	StatisticMessage map[int32]*proto.PlayerMessageInfo //缓存的统计消息
	WightJobs        []*PlayerWightJob                  //权重任务
	WightMax         int32                              //权重最大值
	BetGold          int64                              //下注金币
}

// 添加权重任务
func (p *Player) AddWightJob(wight int32, f func()) {
	job := &PlayerWightJob{
		Wight: wight,
		Fun:   f,
	}
	p.WightMax += wight
	p.WightJobs = append(p.WightJobs, job)
}

// AddScheduleJob 添加掉单任务 ms
func (p *Player) AddScheduleJob(intervalTime, executeCount int32, f func()) {
	job := network.NewScheduleJob(intervalTime, executeCount, f)
	p.ScheduleJobs = append(p.ScheduleJobs, job)
}

// 添加统计消息
func (p *Player) AddMessageInfo(messageId, sequenceNo int32, requestLength int32) {
	m := &proto.PlayerMessageInfo{
		MessageId:          messageId,
		SequenceNo:         sequenceNo,
		StartTime:          util.Now().UnixNano(),
		RequestMessageSize: requestLength,
	}
	//log.Debug("添加消息统计：%v",m)
	p.StatisticMessage[sequenceNo] = m
}

// 移除统计消息
func (p *Player) RemoveMessageInfo(sequenceNo int32) *proto.PlayerMessageInfo {
	if sequenceNo < 1 {
		return nil
	}
	m := p.StatisticMessage[sequenceNo]
	if p != nil {
		delete(p.StatisticMessage, sequenceNo)
		return m
	}
	log.Warn("%v 返回消息序列号:%v 缓存消息不存在 ", p.Id, sequenceNo)
	return nil
}

// 权重任务
type PlayerWightJob struct {
	Wight int32 //权重
	Fun   func()
}
