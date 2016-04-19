/*
 * Auther: CFC4N (cfc4n@cnxct.com)
 * WebSite: http://www.cnxct.com
 * Date: 2015/11/07
 */

package protocol

import (
	"bytes"
	"encoding/binary"
	"gotcp"
	"net"
	"time"
)

const (
	HEARTBEAT_TIME_INTERVAL                                  = 25 * time.Second //心跳包时间间隔
	PACKET_HEADER_SIZE                                       = 16
	MAESTROMESSAGETYPE_GAMECLIENT_CREATE                     = 0
	MAESTROMESSAGETYPE_GAMECLIENT_STOPPED                    = 1
	MAESTROMESSAGETYPE_GAMECLIENT_CRASHED                    = 2
	MAESTROMESSAGETYPE_CLOSE                                 = 3
	MAESTROMESSAGETYPE_HEARTBEAT                             = 4
	MAESTROMESSAGETYPE_REPLY                                 = 5
	MAESTROMESSAGETYPE_LAUNCHERCLIENT                        = 6
	MAESTROMESSAGETYPE_GAMECLIENT_ABANDONED                  = 7
	MAESTROMESSAGETYPE_GAMECLIENT_LAUNCHED                   = 8
	MAESTROMESSAGETYPE_GAMECLIENT_VERSION_MISMATCH           = 9
	MAESTROMESSAGETYPE_GAMECLIENT_CONNECTED_TO_SERVER        = 10 //0x0a
	MAESTROMESSAGETYPE_CHATMESSAGE_TO_GAME                   = 11 //0x0b
	MAESTROMESSAGETYPE_CHATMESSAGE_FROM_GAME                 = 12 //0x0c
	MAESTROMESSAGETYPE_GAMECLIENT_CREATE_VERSION             = 13 //0x0d
	MAESTROMESSAGETYPE_GAMECLIENT_INSTALL_VERSION            = 14 //0x0e
	MAESTROMESSAGETYPE_GAMECLIENT_CANCEL_INSTALL             = 15 //0x0f
	MAESTROMESSAGETYPE_GAMECLIENT_INSTALL_PROGRESS           = 16 //0x10
	MAESTROMESSAGETYPE_GAMECLIENT_INSTALL_PREVIEW            = 17 //0x11
	MAESTROMESSAGETYPE_GAMECLIENT_CANCEL_PREVIEW             = 18 //0x12
	MAESTROMESSAGETYPE_GAMECLIENT_PREVIEW_PROGRESS           = 19 //0x13
	MAESTROMESSAGETYPE_PLAY_REPLAY                           = 20 //0x14
	MAESTROMESSAGETYPE_GAMECLIENT_UNINSTALL_VERSION          = 21 //0x15
	MAESTROMESSAGETYPE_GAMECLIENT_CANCEL_UNINSTALL           = 22 //0x16
	MAESTROMESSAGETYPE_GAMECLIENT_UNINSTALL_PROGRESS         = 23 //0x17
	MAESTROMESSAGETYPE_GAMECLIENT_UNINSTALL_PREVIEW          = 24 //0x18
	MAESTROMESSAGETYPE_GAMECLIENT_CANCEL_UNINSTALL_PREVIEW   = 25 //0x19
	MAESTROMESSAGETYPE_GAMECLIENT_PREVIEW_UNINSTALL_PROGRESS = 26 //0x1a
	MAESTROMESSAGETYPE_GAMECLIENT_ENUMERATE_VERSIONS         = 27 //0x1b
	MAESTROMESSAGETYPE_GAMECLIENT_CREATE_CLIENT_AND_PRELOAD  = 28 //0x1c
	MAESTROMESSAGETYPE_GAMECLIENT_START_PRELOADED_GAME       = 29 //0x1d
	//MAESTROMESSAGETYPE_INVALID
)

type Header struct {
	pHead0   uint32 //默认0x10
	pHead1   uint32 //默认0x01
	pCommand uint32 //默认0x00
	pLength  uint32 //默认0x00
}

/*
 * 进入游戏后，游戏的相关信息
 */
type LolGameInfo struct {
	lolgamekey    string //	进入游戏时的KEY
	lolgamestatus int8
	buddylist     string //好友列表
	ignorelist    string //黑名单列表
}

func NewLolGameInfo() *LolGameInfo {
	return &LolGameInfo{}
}
func (this *Header) Bytes() [PACKET_HEADER_SIZE]byte {
	var p [PACKET_HEADER_SIZE]byte
	binary.LittleEndian.PutUint32(p[:4], this.pHead0)
	binary.LittleEndian.PutUint32(p[4:8], this.pHead1)
	binary.LittleEndian.PutUint32(p[8:12], this.pCommand)
	binary.LittleEndian.PutUint32(p[12:16], this.pLength)
	return p
}

func (head *Header) Read(buf []byte) {
	head.pHead0 = binary.LittleEndian.Uint32(buf[:4])
	head.pHead1 = binary.LittleEndian.Uint32(buf[4:8])
	head.pCommand = binary.LittleEndian.Uint32(buf[8:12])
	head.pLength = binary.LittleEndian.Uint32(buf[12:16])
}

// Packet
type LolLauncherPacket struct {
	pHead Header
	pData []byte
}

func (p *LolLauncherPacket) Serialize() []byte {
	// 拼装head部分
	length := len(p.pData)
	p.pHead.pLength = uint32(length)

	buff := make([]byte, PACKET_HEADER_SIZE+length)
	head := p.pHead.Bytes()
	copy(buff[:PACKET_HEADER_SIZE], head[:PACKET_HEADER_SIZE])
	copy(buff[PACKET_HEADER_SIZE:], p.pData)
	return buff
}

func (p *LolLauncherPacket) GetCommand() uint32 {
	return p.pHead.pCommand
}

func (p *LolLauncherPacket) GetHeader() Header {
	return p.pHead
}

func (p *LolLauncherPacket) GetData() []byte {
	return p.pData
}

func NewLolLauncherPacket(pCommand uint32, pData []byte) *LolLauncherPacket {
	head := Header{
		pHead0:   0x10,
		pHead1:   0x01,
		pCommand: pCommand,
		pLength:  uint32(len(pData)),
	}
	return &LolLauncherPacket{
		pHead: head,
		pData: pData,
	}
}

type LolLauncherProtocol struct {
    unfinished bool		//是否为未完成的TCP 包，默认false
    header Header
    body []byte
}

func (this *LolLauncherProtocol) ReadPacket(conn *net.TCPConn) (gotcp.Packet, error) {
	fullBuf := bytes.NewBuffer([]byte{})
	for {
		data := make([]byte, 1024)

		//暂时不支持超过1024字节长度的单个TCP包
		readLength, err := conn.Read(data)

		if err != nil { //EOF, or worse
			return nil, err
		}

		if readLength == 0 {
			return nil, gotcp.ErrConnClosing
		} else {
			fullBuf.Write(data[:readLength])
		}

		//粘包处理，先判断header.pLenght 是否为0，若为0，怎为新的包
		if (this.unfinished ==true) {
		    if this.header.pLength <=0 {
		        //内部错误
		        return nil, gotcp.ErrReadBlocking
		    }
		    //是未完成的包
		    body := fullBuf.Next(int(this.header.pLength))
			if len(body) != int(this.header.pLength) {
				return nil, gotcp.ErrReadBlocking
			}
			this.body = body
		} else {
		    //新包
		    //转化到header中
			head := fullBuf.Next(16)
			if len(head) < 16 {
				return nil, gotcp.ErrReadBlocking
			}
			this.header = Header{}
			this.header.Read(head)
			length := int(this.header.pLength)
			body := make ([]byte,length)
			body = fullBuf.Next(length)
			if len(body) < length {
			    this.unfinished = true
				return nil, nil
			}
			this.body = body
		}
		
		//设置为已经完成的包
		this.unfinished = false
		return NewLolLauncherPacket(this.header.pCommand, this.body), nil
	}
}
