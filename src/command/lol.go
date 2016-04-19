// +build darwin

/*
 * Auther: CFC4N (cfc4n@cnxct.com)
 * WebSite: http://www.cnxct.com
 * Date: 2015/11/07
 */
package command

import (
	"errors"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"
)

const layout = "2006-01-02 15:04:05"

type launcher_commands struct {
	/*
	*	来自腾讯 lolclient.exe的参数
	 */
	GameSignatureLength string
	SzGameSignature     string
	CltkeyLength        string
	Cltkey              string
	UId                 string
	Host                string
	Xmpp_server_url     string
	Lq_uri              string
	GetClientIpURL      string

	/**
	 * 游戏客户端路径
	 */
	gameroot_path      string
	gamelog_path       string
	lolclientbin_path  string
	lolgamebin_path    string
	lolpatcherbin_path string
	runtime_path       string
	client_cn_dirname  string
}

func (this *launcher_commands) Parse(args []string) error {
	if len(args) < 10 {
		return errors.New("参数不足，请参考 -h指令")
	}
	//	if args[1] != "--" {
	//		return errors.New("参数错误，请参考help指令")
	//	}
	i := 0
	for _, v := range args {
		i++
		var kv = strings.SplitN(strings.TrimSpace(v), "=", 2)
		if len(kv) != 2 {
			i--
			//			return errors.New("参数错误，找不到 ＝ 分隔符")
		}
		switch kv[0] {
		case "gameSignatureLength":
			this.GameSignatureLength = kv[1]
		case "szGameSignature":
			this.SzGameSignature = kv[1]
		case "cltkeyLength":
			this.CltkeyLength = kv[1]
		case "cltkey":
			this.Cltkey = kv[1]
		case "uId":
			this.UId = kv[1]
		case "--host":
			this.Host = kv[1]
		case "--xmpp_server_url":
			this.Xmpp_server_url = kv[1]
		case "--lq_uri":
			this.Lq_uri = kv[1]
		case "--getClientIpURL":
			this.GetClientIpURL = kv[1]
		default:

		}
	}
	if i < 9 {
		return errors.New("参数解析时，发生错误...")
	}
	return nil
}

//启动大厅客户端
func (this *launcher_commands) LolClientCommand(lolclient_port string) {
	log.Println("启动游戏大厅...")
	outfile, err := os.OpenFile(this.gamelog_path+"launcher_client.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0777)
	if err != nil {
		log.Fatalf("%v", err)
	}
	defer outfile.Close()

	t := time.Now()
	outfile.WriteString("\n\n" + t.Format(layout) + ": Client start...\n")
	os.Chdir(this.lolclientbin_path + "deploy/bin")
	this.LolGameSetenv()

	cmd := exec.Command(this.lolclientbin_path+"deploy/bin/LolClient",
		"-runtime",
		"./",
		"-nodebug",
		"META-INF/AIR/application.xml",
		"./",
		"--",
		lolclient_port,
		"gameSignatureLength="+this.GameSignatureLength,
		"szGameSignature="+this.SzGameSignature,
		"cltkeyLength="+this.CltkeyLength,
		"cltkey="+this.Cltkey,
		"uId="+this.UId,
		"--host="+this.Host,
		"--xmpp_server_url="+this.Xmpp_server_url,
		"--lq_uri="+this.Lq_uri,
		"--getClientIpURL="+this.GetClientIpURL,
	)

	cmd.Stderr = outfile
	cmd.Stdout = outfile
	//	cmd.Stdin = os.Stdin
	if err := cmd.Run(); err != nil {
		log.Fatalf("Client-Start: %v", err) //每局游戏结束后,会收到signal: killed的错 @TODO dtruss 跟踪进程号,查找信号是谁发来的
	}
}

//Unable to connect to the server. Please check your network connection and attempt to reconnect to your game.
//You have disconnected. Please check your internet connection and try again.
//启动游戏进程
func (this *launcher_commands) LolGameCommand(lolgame_port, params string) {
	log.Println("启动游戏进程:")
	outfile, err := os.OpenFile(this.gamelog_path+"launcher_game.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0777)
	if err != nil {
		panic(err)
	}
	defer outfile.Close()
	t := time.Now()
	outfile.WriteString("\n\n" + t.Format(layout) + ": Game start...\n")
	os.Chdir(this.lolgamebin_path + "deploy") //  /Applications/League of Legends.app/Contents/LoL/RADS/solutions/lol_game_client_sln/releases/0.0.0.192/deploy
	this.LolGameSetenv()
	cmd := exec.Command(this.lolgamebin_path+"deploy/LeagueOfLegends.app/Contents/MacOS/LeagueofLegends",
		lolgame_port,
		"LoLPatcher",
		this.lolclientbin_path+"deploy/bin/LolClient",
		params,
	)
	cmd.Stderr = outfile
	cmd.Stdout = outfile
	//	cmd.Stdin = os.Stdin
	if err := cmd.Run(); err != nil {
		log.Fatalf("Game-Start: %v", err)
	}
}

func (this *launcher_commands) LolGameSetenv() {
	//若更新,从  ps -p <PID> -wwE  获取
	var LolEnvKeys = [12]string{"LOGNAME", "SHELL", "USER", "SSH_AUTH_SOCK", "XPC_SERVICE_NAME", "HOME", "__CF_USER_TEXT_ENCODING", "TMPDIR", "PATH", "Apple_PubSub_Socket_Render", "XPC_FLAGS", "riot_launched"}
	var LolEnvs map[string]string
	LolEnvs = make(map[string]string)
	for _, v := range LolEnvKeys {
		switch v {
		case "riot_launched":
			LolEnvs[v] = "true"
		case "XPC_SERVICE_NAME":
			LolEnvs[v] = "com.riotgames.MacContainer.67552"
		default:
			LolEnvs[v] = os.Getenv(v)
		}
	}
	os.Clearenv()
	for k1, v1 := range LolEnvs {
		os.Setenv(k1, v1)
	}
}

func (this *launcher_commands) LolSetConfigPath(gameroot_path, runtime_path string) error {
	_, e := ioutil.ReadDir(gameroot_path)
	if e != nil {
		return e
	}
	this.gameroot_path = gameroot_path

	_, e = ioutil.ReadDir(runtime_path)
	if e != nil {
		return e
	}
	this.runtime_path = runtime_path

	this.gamelog_path = this.runtime_path + "/logs/"
	if _, e = os.Stat(this.gamelog_path); e != nil {
		e = os.Mkdir(this.gamelog_path, 0777)
		if e != nil {
			return e
		}
	}

	this.client_cn_dirname = "lol_air_client" //默认英文客户端路径
	//	this.client_cn_dirname = "lol_air_client_tencent"	//中文客户端路径
	return nil
}

func (this *launcher_commands) LolGetConfig() error {
	if len(this.lolclientbin_path) == 0 {
		// 报错，退出
		this.gameroot_path = "/Applications/League of Legends.app/"
	}
	os.Chdir(this.gameroot_path + "Contents/LoL/RADS/") //@todo 修正
	this.lolclientbin_path = this.gameroot_path + "Contents/LoL/RADS/projects/" + this.client_cn_dirname + "/releases/"
	dir_list, e := ioutil.ReadDir(this.lolclientbin_path)
	if e != nil {
		return e
	}
	for _, v := range dir_list {
		if v.IsDir() == true {
			this.lolclientbin_path += v.Name()
			break
		}
	}
	this.lolclientbin_path += "/"

	this.lolgamebin_path = this.gameroot_path + "Contents/LoL/RADS/solutions/lol_game_client_sln/releases/"
	gamedir_list, e := ioutil.ReadDir(this.lolgamebin_path)
	if e != nil {
		return e
	}
	for _, v := range gamedir_list {
		if v.IsDir() == true {
			this.lolgamebin_path += v.Name()
			break
		}
	}

	this.lolgamebin_path += "/"

	this.lolpatcherbin_path = this.gameroot_path + "Contents/LoL/RADS/projects/lol_patcher/releases/"
	pather_list, e := ioutil.ReadDir(this.lolpatcherbin_path)
	if e != nil {
		return e
	}
	for _, v := range pather_list {
		if v.IsDir() == true {
			this.lolpatcherbin_path += v.Name()
			break
		}
	}

	this.lolpatcherbin_path += "/"

	return nil
}

//func (this *launcher_commands) LoLPatcher() {
//	log.Println("启动Patcher进程:")
//	outfile, err := os.OpenFile(this.gamelog_path+"patcher_game.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0777)
//	if err != nil {
//		panic(err)
//	}
//	defer outfile.Close()
//	t := time.Now()
//	outfile.WriteString("\n\n" + t.Format(layout) + ": LoLPatcher start...\n")
//	os.Chdir(this.lolpatcherbin_path+"deploy")	//
//	this.LolGameSetenv()
//	cmd := exec.Command(this.lolpatcherbin_path+"deploy/LoLPatcher.app/Contents/MacOS/LoLPatcher")
//	cmd.Stderr = outfile
//	cmd.Stdout = outfile
//	//	cmd.Stdin = os.Stdin
//	if err := cmd.Run(); err != nil {
//		log.Fatalf("Patcher-Start: %v", err)
//	}
//}

//

func NewLolCommand() *launcher_commands {
	return &launcher_commands{}
}
