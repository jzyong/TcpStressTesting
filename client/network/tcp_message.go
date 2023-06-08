package network

/*
	将请求的一个消息封装到message中，定义抽象层接口
*/
type TcpMessage interface {
	GetLength() uint32         //获取消息数据段长度，包括消息头(消息长度4+消息id4+客户端已收到最小序号4+消息序号4+protobuf消息体)
	GetMessageId() int32       //获取消息ID
	GetSeq() int32             //获取消息序列号
	GetAck() int32             //获取消息确认号
	GetFlag() int32            //获取标识数据，底层和长度公用一个int32
	GetData() []byte           //获取消息内容，protobuf数据
	GetObjectId() int64        //对象唯一id
	SetMessageId(int32)        //设置消息ID
	SetData([]byte)            //设置消息内容
	SetSeq(int32)              //设置消息序列号
	SetObjectId(int64)         //设置对象唯一id
	SetAck(int32)              //设置消消息确认号
	SetFlag(int32)             //设置标识数据，底层和长度公用一个int32
	GetTcpChannel() TcpChannel //获取Channel
	SetTcpChannel(TcpChannel)  //设置Channel
}

//客户端消息体 实现 Message
type tcpMessageImpl struct {
	MessageId  int32      //消息的ID
	Data       []byte     //消息的内容
	Seq        int32      //序列号
	Ack        int32      //确认序列号
	Flag       int32      //消息标识
	TcpChannel TcpChannel //连接会话
}

//创建一个Message消息包
func NewTcpMessage(messageId, seq, ack, flag int32, data []byte) TcpMessage {
	return &tcpMessageImpl{
		MessageId: messageId,
		Seq:       seq,
		Ack:       ack,
		Flag:      flag,
		Data:      data,
	}
}

func (msg *tcpMessageImpl) GetLength() uint32 {
	return 12 + uint32(len(msg.Data))
}

func (msg *tcpMessageImpl) GetSeq() int32 {
	return msg.Seq
}

func (msg *tcpMessageImpl) GetAck() int32 {
	return msg.Ack
}

func (msg *tcpMessageImpl) GetFlag() int32 {
	return msg.Flag
}

func (msg *tcpMessageImpl) SetMessageId(i int32) {
	msg.MessageId = i
}

func (msg *tcpMessageImpl) SetSeq(u int32) {
	msg.Seq = u
}

func (msg *tcpMessageImpl) SetObjectId(u int64) {
	panic("implement me")
}

func (msg *tcpMessageImpl) SetAck(u int32) {
	msg.Ack = u
}

func (msg *tcpMessageImpl) SetFlag(u int32) {
	msg.Flag = u
}

func (msg *tcpMessageImpl) GetTcpChannel() TcpChannel {
	return msg.TcpChannel
}

func (msg *tcpMessageImpl) SetTcpChannel(channel TcpChannel) {
	msg.TcpChannel = channel
}

//获取消息ID
func (msg *tcpMessageImpl) GetMessageId() int32 {
	return msg.MessageId
}

//获取消息内容
func (msg *tcpMessageImpl) GetData() []byte {
	return msg.Data
}

//设计消息内容
func (msg *tcpMessageImpl) SetData(data []byte) {
	msg.Data = data
}

func (msg *tcpMessageImpl) GetObjectId() int64 {
	panic("implement me")
}
