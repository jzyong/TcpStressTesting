package network

import (
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/jzyong/golib/log"
	"io"
	"net"
	"sync"
)

const TcpName = "TcpName" //玩家账号

//定义连接接口
type TcpChannel interface {
	Start()                                      //启动连接，让当前连接开始工作
	Stop()                                       //停止连接，结束当前连接状态M
	GetTCPConnection() *net.TCPConn              //从当前连接获取原始的socket TCPConn
	GetConnID() uint32                           //获取当前连接ID
	SetConnID(connId uint32)                     //设置连接ID
	RemoteAddr() net.Addr                        //获取远程客户端地址信息
	SendMsg(message TcpMessage) error            //发送消息
	IsClose() bool                               //连接是否关闭
	SetProperty(key string, value interface{})   //设置链接属性
	GetProperty(key string) (interface{}, error) //获取链接属性
	RemoveProperty(key string)                   //移除链接属性
	GetMsgBuffChan() chan []byte                 //有缓冲管道，用于读、写两个goroutine之间的消息通信
}

//连接会话 实现Channel
type tcpClientChannelImpl struct {
	TcpChannel
	Conn              net.Conn               //当前连接的socket TCP套接字
	ConnID            uint32                 //当前连接的ID 也可以称作为SessionID，ID全局唯一
	IsClosed          bool                   //当前连接的关闭状态
	MessageDistribute MessageDistribute      //消息管理MsgId和对应处理方法的消息管理模块
	ExitBuffChan      chan bool              //告知该链接已经退出/停止的channel
	msgBuffChan       chan []byte            //有缓冲管道，用于读、写两个goroutine之间的消息通信
	property          map[string]interface{} //链接属性
	propertyLock      sync.RWMutex           //保护链接属性修改的锁
	Client            TcpClient              //连接的客户端
}

//创建连接的方法
func NewClientChannel(conn net.Conn, messageDistribute MessageDistribute, client TcpClient) TcpChannel {
	//初始化Conn属性
	c := &tcpClientChannelImpl{
		Conn:              conn,
		IsClosed:          false,
		MessageDistribute: messageDistribute,
		ExitBuffChan:      make(chan bool, 1),
		msgBuffChan:       make(chan []byte, 1024),
		property:          make(map[string]interface{}),
		Client:            client,
	}
	return c
}

//启动连接，让当前连接开始工作
func (tcpChannel *tcpClientChannelImpl) Start() {
	//1 开启用户从客户端读取数据流程的Goroutine
	go tcpChannel.startReader()
	//2 开启用于写回客户端数据流程的Goroutine
	go tcpChannel.startWriter()
	//按照用户传递进来的创建连接时需要处理的业务，执行钩子方法
	tcpChannel.Client.ChannelActive(tcpChannel)
}

//停止连接，结束当前连接状态M
func (tcpChannel *tcpClientChannelImpl) Stop() {
	log.Info("断开连接：%s", tcpChannel.RemoteAddr().String())
	//如果当前链接已经关闭
	if tcpChannel.IsClosed == true {
		return
	}
	tcpChannel.IsClosed = true

	//如果用户注册了该链接的关闭回调业务，那么在此刻应该显示调用
	tcpChannel.Client.ChannelInactive(tcpChannel)
	// 关闭socket链接
	tcpChannel.Conn.Close()
	//关闭Writer
	tcpChannel.ExitBuffChan <- true

	//关闭该链接全部管道
	close(tcpChannel.ExitBuffChan)
	close(tcpChannel.msgBuffChan)
}

//	写消息Goroutine， 用户将数据发送给客户端
func (tcpChannel *tcpClientChannelImpl) startWriter() {
	defer tcpChannel.Stop()

	for {
		select {
		case data, ok := <-tcpChannel.msgBuffChan:
			if ok {
				//有数据要写给客户端
				if _, err := tcpChannel.Conn.Write(data); err != nil {
					log.Warn("发送数据错误： %v", err)
					return
				}
			} else {
				log.Warn("msgBuffChan is Closed")
				break
			}
		case <-tcpChannel.ExitBuffChan:
			return
		}
	}
}

//	读消息Goroutine，用于从客户端中读取数据
func (tcpChannel *tcpClientChannelImpl) startReader() {
	defer tcpChannel.Stop()
	for {
		// 创建拆包解包的对象
		buffMsgLength := make([]byte, 4)

		// read len
		var decoder = NewClientDataPack()

		//服务器主动断开连接会进入此处
		if _, err := io.ReadFull(tcpChannel.Conn, buffMsgLength); err != nil {
			tcpName, e := tcpChannel.GetProperty(TcpName)
			if e == nil {
				log.Warn("%v: read msg length error:%v", tcpName, err)
			} else {
				log.Warn("read msg length error:%v", err)
			}
			break
		}
		//TODO 需要对长度进行移位操作，前面是标识位
		var msgLength = binary.LittleEndian.Uint32(buffMsgLength)
		// 最大长度验证
		if msgLength > 10000 {
			log.Warn("消息太长：%d\n", msgLength)
		}

		msgData := make([]byte, msgLength)

		if _, err := io.ReadFull(tcpChannel.Conn, msgData); err != nil {
			fmt.Println("read msg data error ", err)
			break
		}
		//拆包，得到msgid 和 数据 放在msg中
		msg, err := decoder.Unpack(msgData, msgLength)
		if err != nil {
			log.Error("unpack error: %v", err)
			break
		}
		//TODO 设置flag
		msg.SetTcpChannel(tcpChannel)
		// 使用调度池 处理任务
		tcpChannel.MessageDistribute.SendMessageToTaskQueue(msg)
		//tcpChannel.MessageDistribute.RunHandler(msg)
	}
}

//获取远程客户端地址信息
func (tcpChannel *tcpClientChannelImpl) RemoteAddr() net.Addr {
	return tcpChannel.Conn.RemoteAddr()
}

func (tcpChannel *tcpClientChannelImpl) GetConnID() uint32 {
	return tcpChannel.ConnID
}

func (tcpChannel *tcpClientChannelImpl) SetConnID(connectId uint32) {
	tcpChannel.ConnID = connectId
}

//直接将Message数据发送数据给远程的TCP客户端
func (tcpChannel *tcpClientChannelImpl) SendMsg(message TcpMessage) error {
	if tcpChannel.IsClosed == true {
		return errors.New("connection closed when send msg")
	}
	//将data封包，并且发送
	var decoder = NewClientDataPack()
	msg, err := decoder.Pack(message)
	if err != nil {
		log.Warn("发送消息[%v]错误", message.GetMessageId())
		return errors.New("Pack error msg ")
	}
	//写回客户端
	tcpChannel.msgBuffChan <- msg
	return nil
}

//是否关闭
func (tcpChannel *tcpClientChannelImpl) IsClose() bool {
	return tcpChannel.IsClosed
}

func (tcpChannel *tcpClientChannelImpl) GetMsgBuffChan() chan []byte {
	return tcpChannel.msgBuffChan
}

//设置链接属性
func (tcpChannel *tcpClientChannelImpl) SetProperty(key string, value interface{}) {
	tcpChannel.propertyLock.Lock()
	defer tcpChannel.propertyLock.Unlock()
	tcpChannel.property[key] = value
}

//获取链接属性
func (tcpChannel *tcpClientChannelImpl) GetProperty(key string) (interface{}, error) {
	tcpChannel.propertyLock.RLock()
	defer tcpChannel.propertyLock.RUnlock()
	if value, ok := tcpChannel.property[key]; ok {
		return value, nil
	} else {
		return nil, errors.New("no property found")
	}
}

//移除链接属性
func (tcpChannel *tcpClientChannelImpl) RemoveProperty(key string) {
	tcpChannel.propertyLock.Lock()
	defer tcpChannel.propertyLock.Unlock()
	delete(tcpChannel.property, key)
}
