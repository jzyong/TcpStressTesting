package network

import (
	"bytes"
	"encoding/binary"
)

/*
	封包数据和拆包数据
	直接面向TCP连接中的数据流,为传输数据添加头部信息，用于处理TCP粘包问题。
*/
type DataPack interface {
	GetHeaderLength() uint32                   //获取包头长度方法
	Pack(msg TcpMessage) ([]byte, error)       //封包方法
	Unpack([]byte, uint32) (TcpMessage, error) //拆包方法
}

//客户端封包拆包类实例，暂时不需要成员 实现DataPack
type clientDataPackImpl struct{}

//封包拆包实例初始化方法
func NewClientDataPack() DataPack {
	return &clientDataPackImpl{}
}

//获取包头长度方法 ,不包括消息长度
func (dp *clientDataPackImpl) GetHeaderLength() uint32 {
	//消息id4+客户端已收到最小序号4+消息序号4
	return 12
}

//封包方法(压缩数据)
func (dp *clientDataPackImpl) Pack(msg TcpMessage) ([]byte, error) {
	//创建一个存放bytes字节的缓冲
	dataBuff := bytes.NewBuffer([]byte{})

	//写dataLen 不包含自身长度
	if err := binary.Write(dataBuff, binary.LittleEndian, msg.GetLength()); err != nil {
		return nil, err
	}
	//写msgID
	if err := binary.Write(dataBuff, binary.LittleEndian, msg.GetMessageId()); err != nil {
		return nil, err
	}
	//写确认序列号
	if err := binary.Write(dataBuff, binary.LittleEndian, msg.GetAck()); err != nil {
		return nil, err
	}
	//写 序列号
	if err := binary.Write(dataBuff, binary.LittleEndian, msg.GetSeq()); err != nil {
		return nil, err
	}
	//写data数据
	if err := binary.Write(dataBuff, binary.LittleEndian, msg.GetData()); err != nil {
		return nil, err
	}
	return dataBuff.Bytes(), nil
}

//拆包方法(解压数据) 消息长度已经被截取 msgLength 保护消息头12字节 ，没有读取标识位，在外层设置
func (dp *clientDataPackImpl) Unpack(binaryData []byte, msgLength uint32) (TcpMessage, error) {
	//创建一个从输入二进制数据的ioReader
	dataBuff := bytes.NewReader(binaryData)

	//只解压head的信息，得到dataLen和msgID
	msg := &tcpMessageImpl{}

	//读msgID
	if err := binary.Read(dataBuff, binary.LittleEndian, &msg.MessageId); err != nil {
		return nil, err
	}
	//读确认序列号
	if err := binary.Read(dataBuff, binary.LittleEndian, &msg.Ack); err != nil {
		return nil, err
	}
	//读 序列号
	if err := binary.Read(dataBuff, binary.LittleEndian, &msg.Seq); err != nil {
		return nil, err
	}
	//读取数据
	data := make([]byte, msgLength-dp.GetHeaderLength())
	if err := binary.Read(dataBuff, binary.LittleEndian, data); err != nil {
		return nil, err
	}
	msg.SetData(data)
	return msg, nil
}
