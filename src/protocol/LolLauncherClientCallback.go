package protocol

import (
	"gotcp"
	"log"
	"time"
)

// 对接LolClient port的callback
type LolLauncherClientCallback struct {
	PacketSendChanToMain      chan *LolLauncherPacket // packet send chanel
	PacketReceiveChanFromMain chan *LolLauncherPacket // packeet receive chanel
	Lolgameinfo               *LolGameInfo
}

func (this *LolLauncherClientCallback) OnConnect(c *gotcp.Conn) bool {
	addr := c.GetRawConn().RemoteAddr()
	c.PutExtraData(addr)
	log.Println("游戏大厅已连接:", addr)
	//	c.AsyncWritePacket(NewLolLauncherPacket(0, []byte("onconnect")), 0)
	// TODO 启动OnMessageFromMain\HeartBeat函数
	go this.OnMessageFromMain(c)
	go this.OnHeartBeat(c)
	return true
}

func (this *LolLauncherClientCallback) OnHeartBeat(c *gotcp.Conn) {
	heartbeat := time.Tick(HEARTBEAT_TIME_INTERVAL)
	for {
		select {
		case <-heartbeat:
			c.AsyncWritePacket(NewLolLauncherPacket(MAESTROMESSAGETYPE_HEARTBEAT, []byte{}), 0)
			//			return
		}
	}
}

func (this *LolLauncherClientCallback) OnMessageFromMain(c *gotcp.Conn) {
	for {
		packet := <-this.PacketReceiveChanFromMain
		data := packet.GetData()
		commandType := packet.GetCommand()
		//		fmt.Println("Client－>OnMessageFromMain:",commandType," -- ",packet.pHead," -- ",data)
		switch commandType {
	    case MAESTROMESSAGETYPE_GAMECLIENT_ABANDONED:
	    	c.AsyncWritePacket(packet, 0)
		case MAESTROMESSAGETYPE_GAMECLIENT_LAUNCHED:
			//0X08
			c.AsyncWritePacket(packet, 0)
		case MAESTROMESSAGETYPE_GAMECLIENT_CONNECTED_TO_SERVER:
			//0XOA 连接到服务器
			c.AsyncWritePacket(packet, 0)
		case MAESTROMESSAGETYPE_CHATMESSAGE_TO_GAME:
			//0x0b 来自游戏大厅的消息，需要转发至游戏进程(在ClientCallback中实现)
			c.AsyncWritePacket(packet, 0)
		case MAESTROMESSAGETYPE_CHATMESSAGE_FROM_GAME:
			c.AsyncWritePacket(packet, 0)
		default:
			//MAESTROMESSAGETYPE_INVALID
			log.Println("Client－>OnMessageFromMain－>IGNOREMESSAGE:", commandType, " -- ", packet.pHead, " -- ", data)
		}
	}
}

func (this *LolLauncherClientCallback) OnMessage(c *gotcp.Conn, p gotcp.Packet) bool {
	packet := p.(*LolLauncherPacket)
	data := packet.GetData()
	commandType := packet.GetCommand()
	switch commandType {
	case MAESTROMESSAGETYPE_GAMECLIENT_CREATE:
		// 0x00 存储data数据，此数据为League of legends 启动参数
		c.AsyncWritePacket(NewLolLauncherPacket(MAESTROMESSAGETYPE_REPLY, []byte{}), 0)
		log.Println("LOGIN KEY:", string(data))
		////启动游戏进程
		this.PacketSendChanToMain <- packet
	case MAESTROMESSAGETYPE_CLOSE:
		//0x03 游戏进程退出 @TODO
		this.PacketSendChanToMain <- packet
	case MAESTROMESSAGETYPE_HEARTBEAT:
		//0x04 回复收到心跳
		c.AsyncWritePacket(NewLolLauncherPacket(MAESTROMESSAGETYPE_REPLY, []byte{}), 0)
	case MAESTROMESSAGETYPE_REPLY:
		//0x05 不处理（一般不会有这种消息)
	case MAESTROMESSAGETYPE_CHATMESSAGE_TO_GAME:
		//0x0b 来自游戏大厅的消息，需要转发至游戏进程 @TODO
		this.PacketSendChanToMain <- packet
	default:
		//MAESTROMESSAGETYPE_INVALID
		log.Println("Client－>OnMessage－>MAESTROMESSAGETYPE_INVALID:", commandType, " -- ", packet.pHead, " -- ", data)
	}

	return true
}

func (this *LolLauncherClientCallback) OnClose(c *gotcp.Conn) {
	log.Println("游戏大厅已退出:", c.GetExtraData())
}
