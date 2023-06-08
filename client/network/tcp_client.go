package network

import (
	"github.com/jzyong/golib/log"
	"net"
)

// Socket客户端
type TcpClient interface {
	Start()                              //启动
	Stop()                               //停止
	Run()                                //开启业务服务
	GetChannel() TcpChannel              //得到链接
	SetChannelActive(func(TcpChannel))   //设置该Client的连接创建时Hook函数
	SetChannelInactive(func(TcpChannel)) //设置该Client的连接断开时的Hook函数
	ChannelActive(channel TcpChannel)    //调用连接OnConnStart Hook函数
	ChannelInactive(channel TcpChannel)  //调用连接OnConnStop Hook函数
}

// Client 接口实现
type tcpClientImpl struct {
	Name              string                //名称
	ServerUrl         string                //服务绑定的地址
	MessageDistribute MessageDistribute     //消息管理模块，用来绑定MsgId和对应的处理方法
	Channel           TcpChannel            //当前的链接管理器
	channelActive     func(conn TcpChannel) //该Client的连接创建时Hook函数
	channelInactive   func(conn TcpChannel) //该Client的连接断开时的Hook函数
	Conn              net.Conn              //网络连接
}

// 创建客户端
func NewTcpClient(name, url string, messageDistribute MessageDistribute) (TcpClient, error) {
	return &tcpClientImpl{
		Name:              name,
		ServerUrl:         url,
		MessageDistribute: messageDistribute,
	}, nil
}

//============== 实现 Client 里的全部接口方法 ========

// 开启网络服务 用go启动
func (s *tcpClientImpl) Start() {
	//log.Debug("%s 连接服务器：%s", s.Name, s.ServerUrl)

	//开启一个go去连接服务器
	go func() {

		//1 连接服务器地址
		conn, err := net.Dial("tcp", s.ServerUrl)
		if err != nil {
			log.Error("Game start err, exit! %v", err)
			return
		}

		//2 已经监听成功
		log.Info("客户端：%s %s连接服务器：%s 成功", s.Name, conn.LocalAddr().String(), s.ServerUrl)
		s.Conn = conn
		channel := NewClientChannel(conn, s.MessageDistribute, s)
		s.Channel = channel
		s.Channel.SetProperty(TcpName, s.Name)

		//3 启动当前链接的处理业务
		go channel.Start()
	}()
	//阻塞,否则主Go退出， listener的go将会退出
	select {}
}

// 停止服务
func (s *tcpClientImpl) Stop() {
	log.Info("%s 连接 %s 关闭", s.Name, s.ServerUrl)
	s.Conn.Close()
}

// 运行服务
func (s *tcpClientImpl) Run() {
	s.Start()
	////阻塞,否则主Go退出， listener的go将会退出
	//select {}
}

// 得到链接
func (s *tcpClientImpl) GetChannel() TcpChannel {
	return s.Channel
}

// 设置该Server的连接创建时Hook函数
func (s *tcpClientImpl) SetChannelActive(channelActiveFun func(TcpChannel)) {
	s.channelActive = channelActiveFun
}

// 设置该Server的连接断开时的Hook函数
func (s *tcpClientImpl) SetChannelInactive(channelInactiveFunc func(TcpChannel)) {
	s.channelInactive = channelInactiveFunc
}

// 调用连接OnConnStart Hook函数
func (s *tcpClientImpl) ChannelActive(conn TcpChannel) {
	if s.channelActive != nil {
		s.channelActive(conn)
	}
}

// 调用连接OnConnStop Hook函数
func (s *tcpClientImpl) ChannelInactive(conn TcpChannel) {
	if s.channelInactive != nil {
		s.channelInactive(conn)
	}
}
