package main

import (
	"flag"
	"github.com/jzyong/TcpStressTesting/client/handler"
	"github.com/jzyong/TcpStressTesting/client/manager"
	"github.com/jzyong/TcpStressTesting/config"
	manager2 "github.com/jzyong/TcpStressTesting/core/manager"
	"github.com/jzyong/TcpStressTesting/core/rpc"
	"github.com/jzyong/golib/log"
	"github.com/jzyong/golib/util"
	"runtime"
)

// 模块管理
type ModuleManager struct {
	*util.DefaultModuleManager
	//核心模块
	StatisticManager *manager2.StatisticManager
	ControlManager   *manager2.ControlManager
	NetworkManager   *manager2.NetworkManager
	GrpcManager      *rpc.GRpcManager

	//客户端逻辑模块
	PlayerManager  *manager.PlayerManager
	RequestManager *manager.RequestManager
}

// 初始化模块
func (m *ModuleManager) Init() error {
	//核心模块
	m.StatisticManager = m.AppendModule(manager2.GetStatisticManager()).(*manager2.StatisticManager)
	m.ControlManager = m.AppendModule(manager2.GetControlManager()).(*manager2.ControlManager)
	m.NetworkManager = m.AppendModule(manager2.GetNetworkManager()).(*manager2.NetworkManager)
	m.GrpcManager = m.AppendModule(rpc.GetGrpcManager()).(*rpc.GRpcManager)

	//客户端逻辑模块
	m.PlayerManager = m.AppendModule(manager.GetPlayerManager()).(*manager.PlayerManager)
	m.RequestManager = m.AppendModule(manager.GetRequestManager()).(*manager.RequestManager)
	return m.DefaultModuleManager.Init()
}

var m = &ModuleManager{
	DefaultModuleManager: util.NewDefaultModuleManager(),
}

// 后台统计入口类
func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	initConfigAndLog()

	log.Info("启动节点:%s  是否为主节点：%v", config.ApplicationConfigInstance.Name, config.ApplicationConfigInstance.Master)

	var err error
	err = m.Init()
	if err != nil {
		log.Error("启动错误: %s", err.Error())
		return
	}

	//注册消息
	handler.RegisterHandlers()

	m.Run()
	util.WaitForTerminate()
	m.Stop()

	util.WaitForTerminate()
}

// 初始化项目配置和日志
func initConfigAndLog() {
	//1.配置文件路径
	configPath := flag.String("config", "D:\\Go\\TcpStressTesting\\config\\application_config_jzy_master.json", "配置文件加载路径")
	flag.Parse()

	config.FilePath = *configPath
	config.ApplicationConfigInstance.Reload()

	//2.关闭debug
	if "DEBUG" != config.ApplicationConfigInstance.LogLevel {
		log.CloseDebug()
	}
	log.SetLogFile("log", "stress-testing-service")

}
