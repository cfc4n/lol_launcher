/*
 * Auther: CFC4N (cfc4n@cnxct.com)
 * WebSite: http://www.cnxct.com
 * Date: 2015/11/07
 */
package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"command"
	"github.com/collinglass/mw"
	"github.com/gorilla/mux"
	"github.com/pborman/getopt"
	"gotcp"
	proto "protocol"
	"replay"
)

const (
	DEFAULT_CLIENT_PORT    uint16 = 8393                  //默认客户端连接的端口
	DEFAULT_GAME_PORT      uint16 = 8394                  //默认游戏程序连接的端口
	DEFAULT_REPLAY_PORT    uint16 = 9527                  //默认录像回放服务的端口
	BUILD_DATE             string = "2015-12-09"          //发布时间
	LAUNCHER_VERSION       string = "beta 0.1"            //发布版本
	TECHNICAL_ANALYSIS_URL string = "http://t.cn/RbkW0FY" //技术分析文章地址
)

type PacketChan struct {
	packetSendChanToMain      chan *proto.LolLauncherPacket // packet send chanel
	packetReceiveChanFromMain chan *proto.LolLauncherPacket // packeet receive chanel
}

type LolLauncher struct {
	Client *gotcp.Server //游戏大厅 TCP服务对象
	Game   *gotcp.Server //游戏进程 TCP服务对象
}

func (this *LolLauncher) watch() {
	// catchs system signal
	chSig := make(chan os.Signal)
	signal.Notify(chSig, syscall.SIGINT, syscall.SIGTERM)
	log.Println("收到系统信号: ", <-chSig)

	// stops services
	this.Client.Stop()
	this.Game.Stop()
}

func main() {

	/*
	 *	lolclient connect port
	 *  lolclient path
	 *  game connect port
	 *  tgp parameter
	 *
	 */

	/*
	 *	本工具参数
	 */
	var lolclient_port uint16
	var lolroot_path string
	var lolobfile string
	var lolgame_port uint16
	var help bool = false
	var replay_modle bool = false
	//	var version bool = false
	fmt.Println()
	fmt.Println("版    本: lol_launcher_mac " + LAUNCHER_VERSION + " build at " + BUILD_DATE + " By CFC4N (cfc4n@cnxct.com)")
	fmt.Println("声    明: 本软件仅供技术交流，游戏娱乐，请勿用于非法用途。\n")

	s := getopt.New()

	/*
	 *	接收本工具参数
	 */
	//	s.StringVarLong(&lolroot_path, "path", 'X', "The root path of League of Legends games", "/Applications/League of Legends.app/")
	//	s.Uint16VarLong(&lolclient_port, "client_port", 'Y',"The port LolClient connected.")
	s.BoolVarLong(&help, "help", 'h', "录像回放模式：-r -f replays/1_123.ob\n游戏模式： gameSignatureLength=数字 szGameSignature=KEY cltkeyLength=数字 cltkey=\"KEY\" uId=QQ号 --host=url --xmpp_server_url=url --lq_uri=url --getClientIpURL=url")
	s.StringVarLong(&lolobfile, "obfile", 'f', "ob录像文件所在路径，建议放在replays目录下。")
	s.BoolVarLong(&replay_modle, "replay", 'r')
	s.Parse(os.Args)

	if help {
		s.PrintUsage(os.Stderr)
		return
	}

	lolroot_path = getCurrentDirectory()

	lolCommands := command.NewLolCommand()
	//设置游戏安装目录，以及日志目录
	lolCommands.LolSetConfigPath("/Applications/League of Legends.app/", lolroot_path)

	//获取游戏中，大厅程序以及游戏进程程序所在目录等配置
	lolCommands.LolGetConfig()

	/*
	 *
	 * 参数判断
	 */
	if lolclient_port <= 0 || lolclient_port >= 65535 {
		lolclient_port = DEFAULT_CLIENT_PORT
	}

	if lolgame_port <= 0 || lolgame_port >= 65535 {
		lolgame_port = DEFAULT_GAME_PORT
	}

	/*
	 *
	 */

	config := &gotcp.Config{
		PacketSendChanLimit:    20,
		PacketReceiveChanLimit: 20,
	}

	srvLauncher := &LolLauncher{}
	clientPacketChan := &PacketChan{
		packetSendChanToMain:      make(chan *proto.LolLauncherPacket, config.PacketSendChanLimit),
		packetReceiveChanFromMain: make(chan *proto.LolLauncherPacket, config.PacketSendChanLimit),
	}

	gamePacketChan := &PacketChan{
		packetSendChanToMain:      make(chan *proto.LolLauncherPacket, config.PacketSendChanLimit),
		packetReceiveChanFromMain: make(chan *proto.LolLauncherPacket, config.PacketSendChanLimit),
	}

	//	@TODO 稍后讲增加对游戏状态管理
	lolgame := proto.LolGameInfo{}

	/*
	 * 监听8393端口
	 */
	tcpAddr, err := net.ResolveTCPAddr("tcp4", "127.0.0.1:"+strconv.Itoa(int(lolclient_port)))
	checkError(err)
	listener, err := net.ListenTCP("tcp", tcpAddr)
	checkError(err)

	srvLauncher.Client = gotcp.NewServer(config, &proto.LolLauncherClientCallback{PacketSendChanToMain: clientPacketChan.packetSendChanToMain, PacketReceiveChanFromMain: clientPacketChan.packetReceiveChanFromMain, Lolgameinfo: &lolgame}, &proto.LolLauncherProtocol{})
	// starts service
	go srvLauncher.Client.Start(listener, time.Second)
	log.Println("大厅服务已监听:", listener.Addr())

	/*
	* 监听8394端口
	 */
	tcpAddrGame, err := net.ResolveTCPAddr("tcp4", "127.0.0.1:"+strconv.Itoa(int(lolgame_port)))
	checkError(err)
	listenerGame, err := net.ListenTCP("tcp", tcpAddrGame)
	checkError(err)

	srvLauncher.Game = gotcp.NewServer(config, &proto.LolLauncherGameCallback{PacketSendChanToMain: gamePacketChan.packetSendChanToMain, PacketReceiveChanFromMain: gamePacketChan.packetReceiveChanFromMain, Lolgameinfo: &lolgame}, &proto.LolLauncherProtocol{})

	// starts service
	go srvLauncher.Game.Start(listenerGame, time.Second)
	log.Println("游戏服务已监听:", listenerGame.Addr())

	//监听客户端通道消息
	go func() {
		for {
			packet := <-clientPacketChan.packetSendChanToMain
			data := packet.GetData()
			commandType := packet.GetCommand()
			switch commandType {
			case proto.MAESTROMESSAGETYPE_GAMECLIENT_CREATE:
				//获取参数，启动游戏进程
				go lolCommands.LolGameCommand(strconv.Itoa(int(lolgame_port)), string(data))
				//消息发送至game客户端
			case proto.MAESTROMESSAGETYPE_CLOSE:
				// 0X03 游戏关闭
			case proto.MAESTROMESSAGETYPE_HEARTBEAT:
				//0x04 回复收到心跳
			case proto.MAESTROMESSAGETYPE_REPLY:
				//0x05 确认收到消息包的回复(可以不做处理)
			case proto.MAESTROMESSAGETYPE_CHATMESSAGE_TO_GAME:
				//0x0b 来自游戏大厅的消息，需要转发至游戏进程(在ClientCallback中实现)
				gamePacketChan.packetReceiveChanFromMain <- packet
			default:
				//MAESTROMESSAGETYPE_INVALID
				log.Println("Client(main)－>OnMessageFromMain－>MAESTROMESSAGETYPE_INVALID:", commandType, " -- ", packet.GetHeader(), " -- ", data)
			}
		}
	}()

	//监听游戏进程通道消息
	go func() {
		for {
			packet := <-gamePacketChan.packetSendChanToMain
			data := packet.GetData()
			commandType := packet.GetCommand()
			switch commandType {
			case proto.MAESTROMESSAGETYPE_CLOSE:
				// 0X03 游戏关闭
				clientPacketChan.packetReceiveChanFromMain <- packet
			case proto.MAESTROMESSAGETYPE_HEARTBEAT:
				//0x04 回复收到心跳
			case proto.MAESTROMESSAGETYPE_REPLY:
				//0x05 确认收到消息包的回复(可以不做处理)
			case proto.MAESTROMESSAGETYPE_GAMECLIENT_ABANDONED:
				//0x07 异常退出
				clientPacketChan.packetReceiveChanFromMain <- packet
			case proto.MAESTROMESSAGETYPE_GAMECLIENT_LAUNCHED:
				// 08 游戏客户端已启动(league of legends进程会主动发送给launcher，launcher通知到client)
				clientPacketChan.packetReceiveChanFromMain <- packet
			case proto.MAESTROMESSAGETYPE_GAMECLIENT_CONNECTED_TO_SERVER:
				// 0x0a League of legends已经连接到服务器
				clientPacketChan.packetReceiveChanFromMain <- packet
			case proto.MAESTROMESSAGETYPE_CHATMESSAGE_FROM_GAME:
				//0x0c 来自游戏进程(League of legends)的聊天消息,##根据样本协议包分析，当收到此消息后，立刻回复一个收到消息包 MAESTROMESSAGETYPE_REPLY ##
				clientPacketChan.packetReceiveChanFromMain <- packet
			case proto.MAESTROMESSAGETYPE_GAMECLIENT_CREATE_VERSION:
				continue
			default:
				//MAESTROMESSAGETYPE_INVALID
				log.Println("Game(main)－>OnMessageFromMain－>MAESTROMESSAGETYPE_INVALID:", commandType, " -- ", packet.GetHeader(), " -- ", data)
			}
		}
	}()

	if replay_modle {
		//观看录像模式
		log.Println("录像观看模式已启动...")
		filePath := lolroot_path + "/" + lolobfile
		log.Println("加载录像文件:", filePath)
		//加载分析OB文件
		err := replay.Loadfile(filePath)
		if err != nil {
			panic(err)
		}
		listenHost := "127.0.0.1:" + strconv.Itoa(int(DEFAULT_REPLAY_PORT))
		//		fmt.Println(replay.GameInfo)
		//		os.Exit(0)
		params := "spectator " + listenHost + " " + replay.GameInfo.Encryption_key + " " + strconv.Itoa(int(replay.GameInfo.Game_id)) + " " + replay.GameMetaData.GameKey.PlatformId

		// log.SetOutput(f)
		log.Println("录像回放服务已监听:", listenHost)

		r := mux.NewRouter()
		mw.Decorate(
			r,
			LoggingMW,
		)

		r.HandleFunc("/observer-mode/rest/featured", replay.FeaturedHandler)
		r.HandleFunc("/observer-mode/rest/consumer/version", replay.VersionHandler)
		r.HandleFunc("/observer-mode/rest/consumer/getGameMetaData/{platformId}/{gameId}/{yolo}/token", replay.GetGameMetaDataHandler)
		r.HandleFunc("/observer-mode/rest/consumer/getLastChunkInfo/{platformId}/{gameId}/{param}/token", replay.GetLastChunkInfoHandler)
		r.HandleFunc("/observer-mode/rest/consumer/getLastChunkInfo/{platformId}/{gameId}/null", replay.EndOfGameStatsHandler)
		r.HandleFunc("/observer-mode/rest/consumer/getGameDataChunk/{platformId}/{gameId}/{chunkId}/token", replay.GetGameDataChunkHandler)
		r.HandleFunc("/observer-mode/rest/consumer/getKeyFrame/{platformId}/{gameId}/{keyFrameId}/token", replay.GetKeyFrameHandler)

		http.Handle("/", r)
		//启动游戏
		//		log.Println("录像播放参数:",params)

		go lolCommands.LolGameCommand(strconv.Itoa(int(lolgame_port)), params)
		if err := http.ListenAndServe(listenHost, nil); err != nil {
			panic(err)
		}
		//监听系统关闭消息
		log.Println("正在监听系统信号，可按Command+C键停止该程序...")
		srvLauncher.watch()
		log.Println("软件退出。")
		os.Exit(0)

	} else {
		//游戏模式
		/*
		 *	接收来自腾讯LolClient的参数
		 */

		err := lolCommands.Parse(os.Args)
		if err != nil {
			checkError(err)
		}

		//启动Patcher进程
		//		go lolCommands.LoLPatcher()
		//启动游戏进程
		go lolCommands.LolClientCommand(strconv.Itoa(int(lolclient_port)))
		//		data:="127.0.0.1 5111 17BLOhi6KZsTtldTsizvHg== 1"
		//		lolCommands.LolGameCommand(strconv.Itoa(int(lolgame_port)), string(data))
		//监听系统关闭消息
		log.Println("正在监听系统信号，可按Command+C键停止该程序...")
		srvLauncher.watch()
		log.Println("软件退出。")
		os.Exit(0)
	}

}

func checkError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func substr(s string, pos, length int) string {
	runes := []rune(s)
	l := pos + length
	if l > len(runes) {
		l = len(runes)
	}
	return string(runes[pos:l])
}
func getParentDirectory(dirctory string) string {
	return substr(dirctory, 0, strings.LastIndex(dirctory, "/"))
}

func getCurrentDirectory() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}
	return strings.Replace(dir, "\\", "/", -1)
}

func LoggingMW(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Path: %s, Method: %s\n", r.URL.Path, r.Method)
		h.ServeHTTP(w, r)
	})
}
