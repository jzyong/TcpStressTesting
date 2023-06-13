# Tcp Stress Test Tool

&nbsp;&nbsp;Distributed customized TCP protocol stress testing,statistics tool.
Because the TCP protocol of each project is different, the project cannot be directly used.
You need to modify the TCP packet logic and application logic by yourself.
[Document](https://jzyong.github.io/TcpStressTesting/)

| Directory | Description                   |
|-----------|-------------------------------|
| client    | tcp client logic(need modify) |
| config    | config file                   |
| core      | core logic                    |
| example   | server demo                   |
| res       | document,script               |

## Features

* Cluster implementation, support for maximum pressure load
* Message delay (minimum, maximum, average)
* Traffic statistics (average, total)
* Server CPU and memory consumption
* Request message and respond message are counted separately
* Message statistics support lua expression filtering

## Usage
### Command line

```shell
# 1. Clone Project
git clone https://github.com/jzyong/TcpStressTesting.git
cd TcpStressTesting

# 2. Build and run tool
go build
.\TcpStressTesting.exe --config config/application_config_jzy_master.json

# 3. Run example server
cd .\example\
go test

# 4. Start and stop test
 cd .\core\rpc\
 go test -v -run StartTest
 go test -v -run StopTest
```

### Docker
 TODO

### Statistical UI
 TODO

## TODO

* GitHub page文档
* Unity 界面统计整理
* docker 部署测试


