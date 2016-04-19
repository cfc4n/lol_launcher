/*
 * Auther: CFC4N (cfc4n@cnxct.com)
 * WebSite: http://www.cnxct.com
 * Date: 2015/11/07
 */
package protocol

import (
	"gotcp"
	"log"
	"time"
)

// 对接LolClient port的callback
// 对接League of legends port的callback
type LolLauncherGameCallback struct {
	PacketSendChanToMain      chan *LolLauncherPacket // packet send chanel
	PacketReceiveChanFromMain chan *LolLauncherPacket // packeet receive chanel
	Lolgameinfo               *LolGameInfo
}

func (this *LolLauncherGameCallback) OnConnect(c *gotcp.Conn) bool {
	addr := c.GetRawConn().RemoteAddr()
	c.PutExtraData(addr)
	log.Println("游戏进程已连接:", addr)
	//	c.AsyncWritePacket(NewLolLauncherPacket(0, []byte("onconnect")), 0)
	// TODO 启动OnMessageFromMain\HeartBeat函数
	go this.OnMessageFromMain(c)
	go this.OnHeartBeat(c)
	return true
}

func (this *LolLauncherGameCallback) OnHeartBeat(c *gotcp.Conn) {
	heartbeat := time.Tick(HEARTBEAT_TIME_INTERVAL)
	for {
		select {
		case <-heartbeat:
			c.AsyncWritePacket(NewLolLauncherPacket(MAESTROMESSAGETYPE_HEARTBEAT, []byte{}), 0)
			return
		}
	}
}

func (this *LolLauncherGameCallback) OnMessageFromMain(c *gotcp.Conn) {
	for {
		packet := <-this.PacketReceiveChanFromMain
		data := packet.GetData()
		commandType := packet.GetCommand()
		switch commandType {
		case MAESTROMESSAGETYPE_CHATMESSAGE_TO_GAME:
			//0x0b 来自游戏大厅的消息，需要转发至游戏进程(在ClientCallback中实现)
			c.AsyncWritePacket(packet, 0)
		default:
			//MAESTROMESSAGETYPE_INVALID
			log.Println("Game－>OnMessageFromMain－>IGNOREMESSAGE:", commandType, " -- ", packet.pHead, " -- ", data)
		}
	}
}

func (this *LolLauncherGameCallback) OnMessage(c *gotcp.Conn, p gotcp.Packet) bool {
	packet := p.(*LolLauncherPacket)
	data := packet.GetData()
	commandType := packet.GetCommand()
	//	fmt.Println("Game－>OnMessage:",commandType," -- ",packet.pHead," -- ",data)
	//	this.PacketSendChanToMain <- packet	//test
	switch commandType {
	case MAESTROMESSAGETYPE_GAMECLIENT_STOPPED:
		//0x01 游戏停止（游戏结束...）
		c.AsyncWritePacket(NewLolLauncherPacket(MAESTROMESSAGETYPE_REPLY, []byte{}), 0)
	case MAESTROMESSAGETYPE_CLOSE:
		// 0x03 游戏进程关闭
		//		fmt.Println("MAESTROMESSAGETYPE_CLOSE:League Of Legends is closed.")
		this.PacketSendChanToMain <- packet
	case MAESTROMESSAGETYPE_HEARTBEAT:
		// 0x04 回复收到心跳
		c.AsyncWritePacket(NewLolLauncherPacket(MAESTROMESSAGETYPE_REPLY, []byte{}), 0)
	case MAESTROMESSAGETYPE_REPLY:
		//0x05 确认收到消息包的回复(可以不做处理)
	case MAESTROMESSAGETYPE_GAMECLIENT_ABANDONED:
		// 0x07 游戏异常退出
		c.AsyncWritePacket(NewLolLauncherPacket(MAESTROMESSAGETYPE_REPLY, []byte{}), 0)
		this.PacketSendChanToMain <- packet
	case MAESTROMESSAGETYPE_GAMECLIENT_LAUNCHED:
		// 08 游戏客户端已启动(league of legends进程会主动发送给launcher)
		c.AsyncWritePacket(NewLolLauncherPacket(MAESTROMESSAGETYPE_REPLY, []byte{}), 0)
		//消息转发到launcher,再转给client最外层程序， 状态标识
		this.PacketSendChanToMain <- packet
	case MAESTROMESSAGETYPE_GAMECLIENT_CONNECTED_TO_SERVER:
		// 0x0a League of legends已经连接到腾讯服务器
		c.AsyncWritePacket(NewLolLauncherPacket(MAESTROMESSAGETYPE_REPLY, []byte{}), 0)
		this.PacketSendChanToMain <- packet
		// @TODO 传值
	case MAESTROMESSAGETYPE_CHATMESSAGE_TO_GAME:
		//0x0b 来自游戏大厅的消息，需要转发至游戏进程(在ClientCallback中实现)
		c.AsyncWritePacket(NewLolLauncherPacket(MAESTROMESSAGETYPE_REPLY, []byte{}), 0)
		this.PacketSendChanToMain <- packet
	case MAESTROMESSAGETYPE_CHATMESSAGE_FROM_GAME:
		//0x0c 来自游戏进程(League of legends)的聊天消息,##根据样本协议包分析，当收到此消息后，立刻回复一个收到消息包 MAESTROMESSAGETYPE_REPLY ##
		c.AsyncWritePacket(NewLolLauncherPacket(MAESTROMESSAGETYPE_REPLY, []byte{}), 0)
		//@TODO 传到最外层程序，然后，外层程序转发消息至LolClient
		this.PacketSendChanToMain <- packet
	default:
		//MAESTROMESSAGETYPE_INVALID
		log.Println("Game－>OnMessage－>MAESTROMESSAGETYPE_INVALID:", commandType, " -- ", packet.pHead, " -- ", data)
	}

	return true
}

func (this *LolLauncherGameCallback) OnClose(c *gotcp.Conn) {
	log.Println("游戏进程已退出:", c.GetExtraData())
}
