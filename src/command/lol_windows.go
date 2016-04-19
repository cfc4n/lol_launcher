// +build windows

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
	gameroot_path     string
	gamelog_path      string
	lolclientbin_path string
	lolgamebin_path   string
	runtime_path      string
}

func (this *launcher_commands) Parse(args []string) error {
	if len(args) < 10 {
		return errors.New("参数不足，请参考help指令")
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

	const layout = "2006-01-02 15:04:05"
	t := time.Now()
	outfile.WriteString("\n\n" + t.Format(layout) + ": Client start...\n")
	os.Chdir(this.lolclientbin_path + "Air")
	os.Setenv("__COMPAT_LAYER", "ElevateCreateProcess")

	cmd := exec.Command(this.lolclientbin_path+"Air/LolClient.exe",
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
		log.Fatalf("Client-Start: %v", err)
	}
}

//启动游戏进程
func (this *launcher_commands) LolGameCommand(lolgame_port, params string) {
	log.Println("启动游戏进程:")
	outfile, err := os.OpenFile(this.gamelog_path+"launcher_game.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0777)
	if err != nil {
		panic(err)
	}
	defer outfile.Close()
	const layout = "2006-01-02 15:04:05"
	t := time.Now()
	outfile.WriteString("\n\n" + t.Format(layout) + ": Game start...\n")
	os.Chdir(this.lolgamebin_path + "Game")
	os.Setenv("__COMPAT_LAYER", "ElevateCreateProcess")
	cmd := exec.Command(this.lolgamebin_path+"Game/League of Legends.exe",
		lolgame_port,
		"lol.launcher_tencent.exe",
		this.lolclientbin_path+"Air/LolClient.exe",
		params,
	)
	cmd.Stderr = outfile
	cmd.Stdout = outfile
	//	cmd.Stdin = os.Stdin
	if err := cmd.Run(); err != nil {
		log.Fatalf("Game-Start: %v", err)
	}
}

func (this *launcher_commands) LolSetConfigPath(gameroot_path, runtime_path string) error {

	//windows下，游戏路径不稳定，故需要把launcher.exe放到游戏目录，然后用runtime_path赋值给gameroot_path
	gameroot_path = runtime_path
	_, e := ioutil.ReadDir(gameroot_path)
	if e != nil {
		return e
	}
	this.gameroot_path = gameroot_path //临时测试，需要把launcher复制到游戏所在目录

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

	return nil
}
func (this *launcher_commands) LolGetConfig() error {
	os.Chdir(this.gameroot_path)

	this.lolclientbin_path = this.gameroot_path + "/"
	this.lolgamebin_path = this.gameroot_path + "/"
	return nil
}

func (this *launcher_commands) LoLPatcher() {
	log.Println("启动Patcher进程:")
}

func NewLolCommand() *launcher_commands {
	return &launcher_commands{}
}
