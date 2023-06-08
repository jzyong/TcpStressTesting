package proto

import (
	"github.com/jzyong/TcpStressTesting/core/model"
	"github.com/jzyong/golib/util"
	"time"
)

// ---------------MessageInterfaceInfo proto定义协议方法扩充------------------
// 请求总次数
func (s *MessageInterfaceInfo) RequestCount() int32 {
	return s.FailCount + s.SuccessCount
}

// 过去时间
func (s *MessageInterfaceInfo) PastSecond() int64 {
	pastTime := (util.Now().UnixNano() - s.StartTime) / int64(time.Second)
	return pastTime
}

// 平均延迟 ms
func (s *MessageInterfaceInfo) DelayAverage() int32 {
	if s.SuccessCount < 1 {
		return 0
	}
	delay := s.ExecuteTime / int64(time.Millisecond) / int64(s.SuccessCount)
	return int32(delay)
}

// 平均大小 bytes
func (s *MessageInterfaceInfo) SizeAverage() int32 {
	count := int64(s.SuccessCount + s.PushCount)
	if count < 1 {
		return 0
	}
	size := s.MessageSize / count
	return int32(size)
}

// 每秒请求次数
func (s *MessageInterfaceInfo) Rps(pastTime int64) float32 {
	requestCount := s.RequestCount()
	if requestCount < 1 {
		return 0
	}
	if pastTime < 1 {
		pastTime = s.PastSecond()
	}

	if pastTime < 1 {
		return 0
	}
	rps := float32(requestCount) / float32(pastTime)
	return rps
}

// 每秒失败次数
func (s *MessageInterfaceInfo) FailRps(pastTime int64) float32 {
	if s.FailCount < 1 {
		return 0
	}
	if pastTime < 1 {
		pastTime = s.PastSecond()
	}
	if pastTime < 1 {
		return 0
	}
	rps := float32(s.FailCount) / float32(pastTime)
	return rps
}

// 每秒失败次数
func (s *MessageInterfaceInfo) PushRps(pastTime int64) float32 {
	if s.PushCount < 1 {
		return 0
	}
	if pastTime < 1 {
		pastTime = s.PastSecond()
	}
	if pastTime < 1 {
		return 0
	}
	rps := float32(s.PushCount) / float32(pastTime)
	return rps
}

// 累计统计 主推消息和请求消息id可能重复
func (s *MessageInterfaceInfo) Add(info *PlayerMessageInfo) {
	if info.ExecuteTime >= model.MessageRequestFailTime {
		s.FailCount++
	} else {
		//主推消息
		if info.ExecuteTime < 1 && info.GetSequenceNo() < 1 {
			s.PushCount++
			//请求消息
		} else {
			s.SuccessCount++
			s.ExecuteTime += info.ExecuteTime
			if info.ExecuteTime < s.DelayMin {
				s.DelayMin = info.ExecuteTime
			}
			if info.ExecuteTime > s.DelayMax {
				s.DelayMax = info.ExecuteTime
			}
			s.RequestMessageSize += int64(info.RequestMessageSize)
		}
		s.MessageSize += int64(info.MessageSize)
	}
	//log.Debug("%v 累计结果:%v", info, s)
}

// 清除，并不是完全清除，消息id和开始时间不清楚
func (s *MessageInterfaceInfo) MergeClear() {
	s.ExecuteTime = 0
	s.FailCount = 0
	s.SuccessCount = 0
	s.MessageSize = 0
	s.RequestMessageSize = 0
	s.PushCount = 0
}

// 合并累加
func (s *MessageInterfaceInfo) MergeAdd(info *MessageInterfaceInfo) {
	s.ExecuteTime += info.ExecuteTime
	s.FailCount += info.FailCount
	s.SuccessCount += info.SuccessCount
	s.PushCount += info.PushCount
	if info.DelayMin < s.DelayMin {
		s.DelayMin = info.DelayMin
	}
	if info.DelayMax > s.DelayMax {
		s.DelayMax = info.DelayMax
	}
	s.MessageSize += info.MessageSize
	s.RequestMessageSize += info.RequestMessageSize
	if s.StartTime < 1 || info.StartTime < s.StartTime {
		s.StartTime = info.StartTime
	}
}

// 请求流量 M
func (s *MessageInterfaceInfo) RequestFlow() float32 {
	return float32(s.RequestMessageSize) / model.Mb
}

// 返回流量 M
func (s *MessageInterfaceInfo) ResponseFlow() float32 {
	size := float32(s.MessageSize) / model.Mb
	return size
}

// 请求流量 M
func (s *MessageInterfaceInfo) RequestFlowRate(pastTime int64) float32 {
	flow := s.RequestFlow()
	if pastTime < 1 {
		pastTime = s.PastSecond()
	}
	if pastTime < 1 {
		return 0
	}
	return flow / float32(pastTime)
}

// 返回流量 M
func (s *MessageInterfaceInfo) ResponseFlowRate(pastTime int64) float32 {
	size := s.ResponseFlow()
	if pastTime < 1 {
		pastTime = s.PastSecond()
	}
	if pastTime < 1 {
		return 0
	}
	return size / float32(pastTime)
}

// IsRequestMessage true 为请求消息
func (s *MessageInterfaceInfo) IsRequestMessage() bool {
	if s.ExecuteTime > 0 || s.SuccessCount > 0 || s.FailCount > 0 {
		return true
	}
	return false
}
