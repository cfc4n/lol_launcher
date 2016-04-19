package parse
import (
	"errors"
	"strings"
)



type launcher_params struct {
	/*
	*	来自腾讯 lolclient.exe的参数
	*/
	GameSignatureLength string
	SzGameSignature string
	CltkeyLength string
	Cltkey string
	UId string
	Host string
	Xmpp_server_url string
	Lq_uri string
	GetClientIpURL string
}


func (this * launcher_params) Parse(args []string) error {
	if len(args) < 10 {
		return errors.New("参数不足，请参考help指令")
	}
//	if args[1] != "--" {
//		return errors.New("参数错误，请参考help指令")
//	}
	i := 0
	for _,v := range args {
		i++
		var kv = strings.SplitN(strings.TrimSpace(v),"=",2)
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

func NewParse () *launcher_params {
	return &launcher_params{}
}