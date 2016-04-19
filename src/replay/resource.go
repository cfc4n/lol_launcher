/*
 * Auther: CFC4N (cfc4n@cnxct.com)
 * WebSite: http://www.cnxct.com
 * Date: 2015/11/07
 */
package replay

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"strconv"
)

type gameInfo struct {
	Src            int    `json:"src"`
	Area_id        int    `json:"area_id"`
	Score          uint32 `json:"score"`
	Game_length    uint16 `json:"game_length"`
	Battle_type    int    `json:"battle_type"`
	Max_tier       []byte `json:"max_tier"`
	Game_id        uint32 `json:"game_id"`
	Start_time     string `json:"start_time"`
	Ob_ver         string `json:"ob_ver"`
	Encryption_key string `json:"encryption_key"`
}

type ChunkInfo struct {
	ChunkId            int `json:"chunkId"`
	AvailableSince     int `json:"availableSince"`
	NextAvailableChunk int `json:"nextAvailableChunk"`
	KeyFrameId         int `json:"keyFrameId"`
	NextChunkId        int `json:"nextChunkId"`
	EndStartupChunkId  int `json:"endStartupChunkId"`
	StartGameChunkId   int `json:"startGameChunkId"`
	EndGameChunkId     int `json:"endGameChunkId"`
	Duration           int `json:"duration"`
}

type MetaData struct {
	GameKey struct {
		GameId     int    `json:"gameId"`
		PlatformId string `json:"platformId"`
	} `json:"gameKey"`
	GameServerAddress         string `json:"gameServerAddress"`
	Port                      int    `json:"port"`
	EncryptionKey             string `json:"encryptionKey"`
	ChunkTimeInterval         int    `json:"chunkTimeInterval"`
	StartTime                 string `json:"startTime"`
	GameEnded                 bool   `json:"gameEnded"`
	LastChunkId               int    `json:"lastChunkId"`
	LastKeyFrameId            int    `json:"lastKeyFrameId"`
	EndStartupChunkId         int    `json:"endStartupChunkId"`
	DelayTime                 int    `json:"delayTime"`
	PendingAvailableChunkInfo []struct {
		Duration     int    `json:"duration"`
		Id           int    `json:"id"`
		ReceivedTime string `json:"receivedTime"`
	} `json:"pendingAvailableChunkInfo"`
	PendingAvailableKeyFrameInfo []struct {
		Id           int    `json:"id"`
		ReceivedTime string `json:"receivedTime"`
		NextChunkId  int    `json:"nextChunkId"`
	} `json:"pendingAvailableKeyFrameInfo"`
	KeyFrameTimeInterval      int    `json:"keyFrameTimeInterval"`
	DecodedEncryptionKey      string `json:"decodedEncryptionKey"`
	StartGameChunkId          int    `json:"startGameChunkId"`
	GameLength                int    `json:"gameLength"`
	ClientAddedLag            int    `json:"clientAddedLag"`
	ClientBackFetchingEnabled bool   `json:"clientBackFetchingEnabled"`
	ClientBackFetchingFreq    int    `json:"clientBackFetchingFreq"`
	InterestScore             int    `json:"interestScore"`
	FeaturedGame              bool   `json:"featuredGame"`
	CreateTime                string `json:"createTime"`
	EndGameChunkId            int    `json:"endGameChunkId"`
	EndGameKeyFrameId         int    `json:"endGameKeyFrameId"`
	FirstChunkId              int    `json:"firstChunkId"`
}


var GameDataChunk map[string][]byte //类型为1  /observer-mode/rest/consumer/getGameDataChunk/:region/:id/:frame/*ignore
var KeyFrame map[string][]byte      //类型为2  /observer-mode/rest/consumer/getKeyFrame/:region/:id/:frame/*ignore
var GameMetaDataJson []byte         //		/observer-mode/rest/consumer/getGameMetaData/:region/:id/*ignore
var LastChunkInfo string            //		/observer-mode/rest/consumer/getLastChunkInfo/:region/:id/:end/*ignore
var ObVersion string                //		/observer-mode/rest/consumer/version
var ObRegion string                 // ob录像所属大区，对应全球服的NA1\EUN1\EUW1\OC1等
var NowChunkId int					//当前chunkid，用于自增
var NowKeyFrameId int					//当前NowKeyFrameId，用于自增
var MaxChunkId int					//最大的chunkid，
var GameInfo gameInfo     //处理ob文件中abstract数据
var encryption_key string // ob watch key
var ChunkKeyListTmp [] []int
var KeyFrameListTmp [] []int
var ChunkKeyList  = make(map[int]int)
var KeyFrameList  = make(map[int]int)
var GameMetaData MetaData

/*
	1,加载ob文件信息到内存
	2,启动http 服务
	3,启动lol进程
	4,接收系统信号
*/

func Loadfile(file string) error {

	body, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}
	alllen := len(body)

	startlen := 0
	/*
		获取tgp lolob版本号信息
	*/

	for i := startlen; i < alllen; i++ {
		if body[i] == 0x0d && body[i+1] == 0x0a {
			//						str1 = body[startlen:i]
			startlen = i + 2
			break
		}
	}

	/*
		获取abstract信息
	*/
	var str2 []byte
	str2 = make([]byte, 0)
	for i := startlen; i < alllen; i++ {
		if body[i] == 0x0d && body[i+1] == 0x0a {
			str2 = body[startlen:i]
			startlen = i + 2
			break
		}
	}

	/*
		获取source信息
	*/
	var str3 []byte
	str3 = make([]byte, 0)
	for i := startlen; i < alllen; i++ {
		if body[i] == 0x0d && body[i+1] == 0x0a {
			str3 = body[startlen:i]
			startlen = i + 2
			break
		}
	}
	fmt.Println()

	/*
		获取obmeta信息
	*/
	var str4 []byte
	str4 = make([]byte, 0)
	for i := startlen; i < alllen; i++ {
		if body[i] == 0x0d && body[i+1] == 0x0a {
			str4 = body[startlen:i]
			startlen = i + 2
			break
		}
	}

	/*
		获取 keyframe_tab 信息
	*/
	var str5 []byte
	str5 = make([]byte, 0)
	for i := startlen; i < alllen; i++ {
		if body[i] == 0x0d && body[i+1] == 0x0a {
			str5 = body[startlen:i]
			startlen = i + 2
			break
		}
	}
	//	fmt.Println(string(str5))
	//	fmt.Println()

	/*
		获取 chunk_tab 信息
	*/
	var str6 []byte
	str6 = make([]byte, 0)
	for i := startlen; i < alllen; i++ {
		if body[i] == 0x0d && body[i+1] == 0x0a {
			str6 = body[startlen:i]
			startlen = i + 2
			break
		}

	}
	//	fmt.Println(string(str6))
	//	fmt.Println()

	var tgp_strs [][]byte
	//	tgp_strs = append(tgp_strs,str1)	//ob tgp 版本信息，忽略
	tgp_strs = append(tgp_strs, str2)
	tgp_strs = append(tgp_strs, str3)
	tgp_strs = append(tgp_strs, str4)
	tgp_strs = append(tgp_strs, str5)
	tgp_strs = append(tgp_strs, str6)

	//查找 ： 分割，取前面部分，判断是否为某字符串
	for n := 0; n < len(tgp_strs); n++ {
		strtmp := tgp_strs[n]
		var name []byte
		var i int
		for i = 0; i < len(strtmp); i++ {
			if strtmp[i] == 0x3a {
				name = strtmp[:i]
				break
			}
		}

		switch string(name) {
		case "obmeta":
			GameMetaDataJson = strtmp[i+1:]
		case "abstract":
			err = json.Unmarshal(strtmp[i+1:], &GameInfo)
			if err != nil {
				return err
			}
			if GameInfo.Ob_ver != "" {
				ObVersion = GameInfo.Ob_ver
			} else {
				return fmt.Errorf("Cant found ob_ver")
			}
			encryption_key = GameInfo.Encryption_key
		case "source":
//			fmt.Println(string(strtmp[i+1:]))
		case "keyframe_tab":
			err := json.Unmarshal(strtmp[i+1:], &KeyFrameListTmp)
			if err != nil {
				return err
			}
			for _,v := range KeyFrameListTmp {
				if len(v) != 2 {
					return errors.New("Can't explin [chunk_tab] string")
				}
				k1 := v[0]
				v1 := v[1]
				KeyFrameList[v1] = k1
			}
		case "chunk_tab":
			err := json.Unmarshal(strtmp[i+1:], &ChunkKeyListTmp)
			if err != nil {
				return err
			}
			for _,v := range ChunkKeyListTmp {
				if len(v) != 2 {
					return errors.New("Can't explin [chunk_tab] string")
				}
				k1 := v[0]
				v1 := v[1]
				if k1 > MaxChunkId {
					MaxChunkId = k1
				}
				ChunkKeyList[k1] = v1
			}
//			fmt.Println(MaxChunkId)
		default:
		}
	}

	GameMetaData = MetaData{}
	err = json.Unmarshal(GameMetaDataJson, &GameMetaData)
	if err != nil {
		return err
	}

	log.Println("场次ID:", GameInfo.Game_id)
	log.Println("所属大区:", GameMetaData.GameKey.PlatformId)
//	log.Println("协议版本:", ObVersion)


	//"spectator 127.0.0.1:8080 4iDght8wUXHhzlP37OnRb2ekRVRWecFj 1643152753 HN1"
	NowKeyFrameId = 1	//初始化
	GameDataChunk = make(map[string][]byte)
	KeyFrame = make(map[string][]byte)
	//增加换行符的长度（文本字符跟16进制字符部分，有两个\r\n ，上面已经加过2了，故这里只加2)
	startlen += 2

	strlen := 0
	head := make([]byte, 0)
	for i := startlen; i < alllen; i++ {
	here:
		startlen += strlen
		if startlen+7 > alllen {
//			fmt.Println("END...")
			break
		}
		head = body[startlen : startlen+7]
//		fmt.Println(head, "\t\t", body[startlen+7:startlen+30]) //这里顺带打印出后面30字节数据，观察确认用
		strlen = int(uint32(head[6]) | uint32(head[5])<<8 | uint32(head[4])<<16 | uint32(head[3])<<24)

		//开始按照第一个字节进行添加到对应map中
		switch head[0] {
		case 1:
			//GameDataChunk
			key := int(head[2] | head[1]<<8)
			GameDataChunk[strconv.Itoa(key)] = body[startlen+7 : startlen+7+strlen]
		case 2:
			//KeyFrame
			key := int(head[2] | head[1]<<8)
			KeyFrame[strconv.Itoa(key)] = body[startlen+7 : startlen+7+strlen]
		}

		//处理完成后，增加startlen长度
		startlen += 7
		goto here
	}
	return nil
}

// Lists the 10 featured games for the regions supported by this server.
//@TODO
func featured() []byte {
	return []byte(`{"gameList":[{"gameId":2132699567,"mapId":11,"gameMode":"CLASSIC","gameType":"MATCHED_GAME","gameQueueConfigId":4,"participants":[{"teamId":100,"spell1Id":3,"spell2Id":4,"championId":45,"skinIndex":0,"profileIconId":749,"summonerName":"민기잉","bot":false},{"teamId":100,"spell1Id":4,"spell2Id":7,"championId":18,"skinIndex":6,"profileIconId":881,"summonerName":"나는상윤","bot":false},{"teamId":100,"spell1Id":12,"spell2Id":4,"championId":39,"skinIndex":2,"profileIconId":540,"summonerName":"시티즌팍","bot":false},{"teamId":100,"spell1Id":14,"spell2Id":4,"championId":412,"skinIndex":4,"profileIconId":7,"summonerName":"병장된에프람","bot":false},{"teamId":100,"spell1Id":11,"spell2Id":4,"championId":421,"skinIndex":0,"profileIconId":7,"summonerName":"Crown is heavy","bot":false},{"teamId":200,"spell1Id":4,"spell2Id":14,"championId":223,"skinIndex":0,"profileIconId":872,"summonerName":"CJ Entus 민기","bot":false},{"teamId":200,"spell1Id":7,"spell2Id":4,"championId":222,"skinIndex":0,"profileIconId":839,"summonerName":"쫀 끄레OI머","bot":false},{"teamId":200,"spell1Id":12,"spell2Id":4,"championId":117,"skinIndex":2,"profileIconId":937,"summonerName":"라꼬기","bot":false},{"teamId":200,"spell1Id":4,"spell2Id":11,"championId":77,"skinIndex":4,"profileIconId":23,"summonerName":"도근보근","bot":false},{"teamId":200,"spell1Id":14,"spell2Id":4,"championId":238,"skinIndex":1,"profileIconId":918,"summonerName":"FNC NeMo","bot":false}],"observers":{"encryptionKey":"8/D9TTkvUJLdcwVdLN0FNg21rf9+kBVi"},"platformId":"KR","gameTypeConfigId":2,"bannedChampions":[{"championId":41,"teamId":100,"pickTurn":1},{"championId":429,"teamId":200,"pickTurn":2},{"championId":76,"teamId":100,"pickTurn":3},{"championId":60,"teamId":200,"pickTurn":4},{"championId":203,"teamId":100,"pickTurn":5},{"championId":13,"teamId":200,"pickTurn":6}],"gameStartTime":1445850945462,"gameLength":178},{"gameId":2132707534,"mapId":11,"gameMode":"CLASSIC","gameType":"MATCHED_GAME","gameQueueConfigId":4,"participants":[{"teamId":100,"spell1Id":4,"spell2Id":7,"championId":18,"skinIndex":6,"profileIconId":778,"summonerName":"give me adcarryz","bot":false},{"teamId":100,"spell1Id":3,"spell2Id":4,"championId":16,"skinIndex":0,"profileIconId":7,"summonerName":"다미갓 2","bot":false},{"teamId":100,"spell1Id":12,"spell2Id":4,"championId":38,"skinIndex":4,"profileIconId":881,"summonerName":"MickeyGod","bot":false},{"teamId":100,"spell1Id":12,"spell2Id":4,"championId":24,"skinIndex":0,"profileIconId":23,"summonerName":"Hcm","bot":false},{"teamId":100,"spell1Id":11,"spell2Id":4,"championId":421,"skinIndex":1,"profileIconId":918,"summonerName":"FNC ReHope","bot":false},{"teamId":200,"spell1Id":11,"spell2Id":4,"championId":60,"skinIndex":2,"profileIconId":597,"summonerName":"SpiriToT","bot":false},{"teamId":200,"spell1Id":14,"spell2Id":4,"championId":25,"skinIndex":5,"profileIconId":532,"summonerName":"Lin4an","bot":false},{"teamId":200,"spell1Id":4,"spell2Id":7,"championId":222,"skinIndex":0,"profileIconId":4,"summonerName":"CoreJJ","bot":false},{"teamId":200,"spell1Id":14,"spell2Id":12,"championId":120,"skinIndex":2,"profileIconId":924,"summonerName":"이 차가 식기전에","bot":false},{"teamId":200,"spell1Id":12,"spell2Id":4,"championId":58,"skinIndex":5,"profileIconId":924,"summonerName":"CJ Entus 고스트","bot":false}],"observers":{"encryptionKey":"ai8AMw4S60uqDkgSfuwnN/k298CnuHbW"},"platformId":"KR","gameTypeConfigId":2,"bannedChampions":[{"championId":203,"teamId":100,"pickTurn":1},{"championId":92,"teamId":200,"pickTurn":2},{"championId":429,"teamId":100,"pickTurn":3},{"championId":41,"teamId":200,"pickTurn":4},{"championId":76,"teamId":100,"pickTurn":5},{"championId":45,"teamId":200,"pickTurn":6}],"gameStartTime":1445850923233,"gameLength":201},{"gameId":2132699888,"mapId":11,"gameMode":"CLASSIC","gameType":"MATCHED_GAME","gameQueueConfigId":4,"participants":[{"teamId":100,"spell1Id":14,"spell2Id":12,"championId":120,"skinIndex":4,"profileIconId":905,"summonerName":"for s0meday","bot":false},{"teamId":100,"spell1Id":4,"spell2Id":14,"championId":103,"skinIndex":1,"profileIconId":25,"summonerName":"zl존태봉123","bot":false},{"teamId":100,"spell1Id":7,"spell2Id":4,"championId":22,"skinIndex":6,"profileIconId":939,"summonerName":"I is peco","bot":false},{"teamId":100,"spell1Id":14,"spell2Id":4,"championId":53,"skinIndex":0,"profileIconId":6,"summonerName":"Like like a fish","bot":false},{"teamId":100,"spell1Id":11,"spell2Id":4,"championId":76,"skinIndex":6,"profileIconId":879,"summonerName":"Ch2ser","bot":false},{"teamId":200,"spell1Id":11,"spell2Id":4,"championId":64,"skinIndex":5,"profileIconId":592,"summonerName":"Awe 착한카직스","bot":false},{"teamId":200,"spell1Id":12,"spell2Id":4,"championId":68,"skinIndex":0,"profileIconId":6,"summonerName":"woongplayer","bot":false},{"teamId":200,"spell1Id":7,"spell2Id":4,"championId":42,"skinIndex":0,"profileIconId":28,"summonerName":"나이스쿤","bot":false},{"teamId":200,"spell1Id":3,"spell2Id":4,"championId":201,"skinIndex":1,"profileIconId":775,"summonerName":"바 이 꿈나무","bot":false},{"teamId":200,"spell1Id":14,"spell2Id":4,"championId":245,"skinIndex":0,"profileIconId":585,"summonerName":"MopPy","bot":false}],"observers":{"encryptionKey":"QBHNVoH/siBeHBPwRBcFkHY5x9F66vuV"},"platformId":"KR","gameTypeConfigId":2,"bannedChampions":[{"championId":41,"teamId":100,"pickTurn":1},{"championId":4,"teamId":200,"pickTurn":2},{"championId":60,"teamId":100,"pickTurn":3},{"championId":28,"teamId":200,"pickTurn":4},{"championId":203,"teamId":100,"pickTurn":5},{"championId":122,"teamId":200,"pickTurn":6}],"gameStartTime":1445850927583,"gameLength":196},{"gameId":2132730318,"mapId":11,"gameMode":"CLASSIC","gameType":"MATCHED_GAME","gameQueueConfigId":4,"participants":[{"teamId":100,"spell1Id":12,"spell2Id":4,"championId":112,"skinIndex":0,"profileIconId":21,"summonerName":"EDG소프트","bot":false},{"teamId":100,"spell1Id":4,"spell2Id":11,"championId":421,"skinIndex":1,"profileIconId":922,"summonerName":"Hardjunglecarry","bot":false},{"teamId":100,"spell1Id":7,"spell2Id":4,"championId":236,"skinIndex":0,"profileIconId":621,"summonerName":"일산마동석","bot":false},{"teamId":100,"spell1Id":12,"spell2Id":4,"championId":13,"skinIndex":3,"profileIconId":718,"summonerName":"원 출","bot":false},{"teamId":100,"spell1Id":14,"spell2Id":4,"championId":412,"skinIndex":4,"profileIconId":19,"summonerName":"clan support","bot":false},{"teamId":200,"spell1Id":12,"spell2Id":4,"championId":39,"skinIndex":1,"profileIconId":758,"summonerName":"THE TIME T0 DIE","bot":false},{"teamId":200,"spell1Id":7,"spell2Id":4,"championId":67,"skinIndex":6,"profileIconId":621,"summonerName":"빅토리 AdcOneTop","bot":false},{"teamId":200,"spell1Id":12,"spell2Id":4,"championId":245,"skinIndex":0,"profileIconId":576,"summonerName":"m1sha","bot":false},{"teamId":200,"spell1Id":4,"spell2Id":11,"championId":64,"skinIndex":1,"profileIconId":938,"summonerName":"Pedeo","bot":false},{"teamId":200,"spell1Id":3,"spell2Id":4,"championId":223,"skinIndex":0,"profileIconId":576,"summonerName":"돈벌러호주감","bot":false}],"observers":{"encryptionKey":"cQXJffIDCCJIJuQYYsJiMcC1nDH4T6SP"},"platformId":"KR","gameTypeConfigId":2,"bannedChampions":[{"championId":41,"teamId":100,"pickTurn":1},{"championId":429,"teamId":200,"pickTurn":2},{"championId":4,"teamId":100,"pickTurn":3},{"championId":92,"teamId":200,"pickTurn":4},{"championId":203,"teamId":100,"pickTurn":5},{"championId":76,"teamId":200,"pickTurn":6}],"gameStartTime":1445850966088,"gameLength":158},{"gameId":2132707537,"mapId":11,"gameMode":"CLASSIC","gameType":"MATCHED_GAME","gameQueueConfigId":4,"participants":[{"teamId":100,"spell1Id":14,"spell2Id":4,"championId":412,"skinIndex":4,"profileIconId":925,"summonerName":"너 왜그래 진짜","bot":false},{"teamId":100,"spell1Id":4,"spell2Id":11,"championId":245,"skinIndex":0,"profileIconId":749,"summonerName":"MMMMMNMM","bot":false},{"teamId":100,"spell1Id":7,"spell2Id":4,"championId":429,"skinIndex":0,"profileIconId":20,"summonerName":"도적인생","bot":false},{"teamId":100,"spell1Id":4,"spell2Id":21,"championId":69,"skinIndex":4,"profileIconId":924,"summonerName":"더샤이리븐개못함","bot":false},{"teamId":100,"spell1Id":14,"spell2Id":12,"championId":120,"skinIndex":0,"profileIconId":918,"summonerName":"get angry","bot":false},{"teamId":200,"spell1Id":11,"spell2Id":4,"championId":421,"skinIndex":2,"profileIconId":786,"summonerName":"Sar2","bot":false},{"teamId":200,"spell1Id":7,"spell2Id":4,"championId":67,"skinIndex":5,"profileIconId":7,"summonerName":"Spieth","bot":false},{"teamId":200,"spell1Id":3,"spell2Id":4,"championId":12,"skinIndex":0,"profileIconId":23,"summonerName":"마스터가서닉변","bot":false},{"teamId":200,"spell1Id":12,"spell2Id":4,"championId":36,"skinIndex":0,"profileIconId":937,"summonerName":"주둥이닫고겜해라","bot":false},{"teamId":200,"spell1Id":12,"spell2Id":4,"championId":105,"skinIndex":8,"profileIconId":871,"summonerName":"미드자르반짱짱맨","bot":false}],"observers":{"encryptionKey":"wVt2TEDlK4/CvRvSfYvoksGFWM3OkrBx"},"platformId":"KR","gameTypeConfigId":2,"bannedChampions":[{"championId":122,"teamId":100,"pickTurn":1},{"championId":28,"teamId":200,"pickTurn":2},{"championId":13,"teamId":100,"pickTurn":3},{"championId":4,"teamId":200,"pickTurn":4},{"championId":203,"teamId":100,"pickTurn":5},{"championId":60,"teamId":200,"pickTurn":6}],"gameStartTime":1445850907191,"gameLength":217}],"clientRefreshInterval":300}`)
}

// Contains the current version for this Region.
func version() string {
	return ObVersion
}

// URL: .../consumer/getGameMetaData/{platformId}/{gameID}/1/token
// Returns information about the given game. This contains the games type and map, summoners involved, champions picked & banned, start time of the game and the encryption key required to read the replay data.
func getGameMetaData(platformId, gameId string) (*MetaData, error) {
	return &GameMetaData, nil
}

// URL: .../consumer/getLastChunkInfo/{platformId}/{gameID}/{param}/token
// Return some information about the last available chunk:

func getLastChunkInfo(platformId, gameId string, param string) (*ChunkInfo, error) {

	//计算该函数调用次数
	Duration := ChunkKeyList[GameMetaData.StartGameChunkId]
	NowChunkId++
	if NowChunkId > GameMetaData.StartGameChunkId && NowChunkId <= MaxChunkId {
		GameMetaData.StartGameChunkId = NowChunkId
		Duration = ChunkKeyList[NowChunkId]
		if KeyFrameList[GameMetaData.StartGameChunkId] != 0 {
			NowKeyFrameId++
		}
	}


	result := ChunkInfo{
		ChunkId:            GameMetaData.StartGameChunkId,
		AvailableSince:     32502, //30000
		NextAvailableChunk: 26976, //30000
		KeyFrameId:         NowKeyFrameId,
		NextChunkId:        GameMetaData.StartGameChunkId,
		EndStartupChunkId:  GameMetaData.EndStartupChunkId,
		StartGameChunkId:   GameMetaData.StartGameChunkId,
		EndGameChunkId:     MaxChunkId,
		Duration:           Duration, //30018
	}

//	//0 Param requires end data for client to know to stream the rest
	if tmp, _ := strconv.Atoi(param); tmp == 0 {
		result = ChunkInfo{
			ChunkId:            GameMetaData.StartGameChunkId,
			AvailableSince:     32502, //30000
			NextAvailableChunk: 26976, //30000
			KeyFrameId:         NowKeyFrameId,
			NextChunkId:        GameMetaData.StartGameChunkId,
			EndStartupChunkId:  GameMetaData.EndStartupChunkId,
			StartGameChunkId:   GameMetaData.StartGameChunkId,
			EndGameChunkId:     MaxChunkId,
			Duration:           Duration, //30018
		}
	}
//
//	if tmp, _ := strconv.Atoi(param); tmp == 0 {
//		result = ChunkInfo{
//			ChunkId:            GameMetaData.LastChunkId,
//			AvailableSince:     30000,
//			NextAvailableChunk: 30000,
//			KeyFrameId:         GameMetaData.LastKeyFrameId,
//			NextChunkId:        GameMetaData.LastChunkId - 1,
//			EndStartupChunkId:  GameMetaData.EndStartupChunkId,
//			StartGameChunkId:   GameMetaData.StartGameChunkId,
//			EndGameChunkId:     GameMetaData.LastChunkId,
//			Duration:           30000,
//		}
//	}

	return &result, nil
}

// URL: .../consumer/getGameDataChunk/{platformId}/{gameID}/{chunkId}/token
// Retrieves a chunk of data for the given game.
func getGameDataChunk(platformId, gameId, chunkId string) ([]byte, error) {
//	log.Printf("[getGameDataChunk]: %s,Not used params:platformId - %s , gameId - %s", chunkId, platformId, gameId)
	if GameDataChunk[chunkId] == nil {
		return []byte{}, errors.New("Cant found chunkID:" + chunkId)
	}
	return GameDataChunk[chunkId], nil
}

// URL: .../consumer/getKeyFrame/{platformId}/{gameID}/{keyFrameId}/token
// Retrieves a key frame for the given game.
func getKeyFrame(platformId, gameId, keyFrameId string) ([]byte, error) {
//	log.Printf("[getGameDataChunk]: %s,Not used params:platformId - %s , gameId - %s", keyFrameId, platformId, gameId)

	if KeyFrame[keyFrameId] == nil {
		return []byte{}, errors.New("Cant found keyFrameId:" + keyFrameId)
	}
	return KeyFrame[keyFrameId], nil
}

// INCOMPLETE
// URL: .../consumer/getLastChunkInfo/{platformId}/{gameID}/null (!)
// Contains data used for the statistics screen after a game.
func endOfGameStats(platformId, gameId string) []byte {
	return nil
}
