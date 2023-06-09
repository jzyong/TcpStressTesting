# 简介

压力测试集群客户端，参考Locust特点进行开发。 Locust对Http支持良好，TCP需要自己扩展，需要学习Python，不能很好进行功能扩展，自己实现额外功能。

| 目录      | 描述                         |
|---------|----------------------------|
| client  | 客户端逻辑，client调用core，不能反向调用  |
| config  | 配置文件                       |
| core    | 公共基础包,和游戏逻辑无关，实现类似Locust功能 |
| example | 示例服务器                      |
| res     | 文档，protobuf等               |

## 特性

* 集群实现，支持最大压测负荷
* 消息延迟（最小值，最大值，平均值）
* 流量统计（平均，总共）
* 主机CPU，内存消耗
* 单独统计请求协议和返回协议
* 协议统计支持lua表达式过滤

## 玩家调度模型

1. 创建go routine池
2. 请求返回消息根据玩家id选择routine执行
3. cron定时器检测玩家，执行逻辑封装成job丢入routine池，根据玩家id分发
4. 玩家调度任务封装成ScheduleJob加入玩家slice中，PlayerManager定时检测，时间到了丢入routine池，根据玩家id分发

## TODO
* README编写
* GitHub page文档
* example测试服务器，protobuf
* 协议生成



