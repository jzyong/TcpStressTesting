package config

import (
	"encoding/json"
	"github.com/jzyong/golib/log"
	"io/ioutil"
	"os"
)

//配置
var ApplicationConfigInstance *ApplicationConfig

//配置文件路径
var FilePath string

//配置 ，需要命令行支持
type ApplicationConfig struct {
	RpcHost    string   `json:"rpcHost"`    //自己 rpc 地址
	MasterHost string   `json:"masterHost"` //主节点host
	Master     bool     `json:"master"`     //true 为主节点，主节点需要协调worker节点
	UserCount  int32    `json:"userCount"`  //用户个数
	SpawnRate  int32    `json:"spawnRate"`  //用户生成速率 个/s
	ClientHost string   `json:"clientHost"` //客户端地址，http://为http请求，否则TCP
	Profile    string   `json:"profile"`    //个性化配置
	LogLevel   string   `json:"logLevel"`   //日志级别
	Name       string   `json:"name"`       //名称
	GateUrls   []string `json:"gateUrls"`   //网关地址列表
	TestType   int32    `json:"testType"`   //测试类型，0标准测试，其他为slotsId针对slots小游戏测试
}

func init() {
	ApplicationConfigInstance = &ApplicationConfig{
		RpcHost:   "127.0.0.1:5001",
		Master:    true,
		UserCount: 1,
		SpawnRate: 1,
		LogLevel:  "DEBUG",
		Profile:   "jzy",
		Name:      "master",
	}
	//ApplicationConfigInstance.Reload()
}

//判断一个文件是否存在
func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

//读取用户的配置文件
func (statsConfig *ApplicationConfig) Reload() {
	if confFileExists, _ := PathExists(FilePath); confFileExists != true {
		//fmt.Println("Config File ", g.ConfFilePath , " is not exist!!")
		log.Warn("config file ", FilePath, "not find, use default config")
		return
	}
	data, err := ioutil.ReadFile(FilePath)
	if err != nil {
		panic(err)
	}
	//将json数据解析到struct中
	err = json.Unmarshal(data, statsConfig)
	if err != nil {
		log.Error("%v", err)
		panic(err)
	}
}
