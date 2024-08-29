package services

import (
	"interactive-server/logger"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	jsoniter "github.com/json-iterator/go"
)

var (
	json2     = jsoniter.ConfigCompatibleWithStandardLibrary
	WsCliPool = make(map[string]Client)
	Mu        = new(sync.Mutex)
)

type Client struct {
	WsConn     *websocket.Conn
	ConnLock   *sync.Mutex
	ConnMsg    Message  //用户连接时的message信息
	CloseCh    chan int //关闭连接订阅
	WriteMsgCh chan Message
}

type Service struct {
	Broadcast chan Message
	BanCh     chan chBanInfo //判断是否被禁言 禁用channel
}

type chBanInfo struct {
	Reason  string //原因
	UserId  string //连接的用户Id
	Ban     bool   //是否禁用/禁言
	BanType string //禁用类型(0禁言1禁用)
}

type OPERATION int

const (
	SYNC_POS            OPERATION = iota //位置同步消息
	DISCONNECT                           //断开连接
	OPEN_MICRO                           //打开麦克风(用户列表)
	CLOS_MICRO                           //关闭麦克风（关闭的用户ID）
	OPEN_SPEAKER                         //打开扬声器（用户列表）
	CLOSE_SPEAKER                        //关闭扬声器
	CONNECT                              //客户端连接
	ENTERSCENE                           //进入场景提示
	SHARE_SCREEN                         //打开屏幕分享（分享列表）
	EXIT_SHARE_SCREEN                    //关闭屏幕分享
	JOIN_SHARE                           //加入分享
	CHAT                                 //聊天
	EXIT_SHARE                           //离开分享
	BAN                                  //禁用
	BAN_RELEASE                          //解除禁用
	BAN_TO_POST                          //禁言
	BAN_TO_POST_RELEASE                  //解除禁言
	OCCUPIED                             //用户被挤出,通知原连接
)

type Message struct {
	UserId       string    `json:"id"`        //用户Id
	Operation    OPERATION `json:"operation"` //操作类型
	SceneId      string    `json:"sceneid"`
	Text         string    `json:"text"`
	Range        int       `json:"range"` //0全局 1场景 2房间
	UserModel    int       `json:"usermodel"`
	Pos          Xyz       `json:"pos"`
	Rot          Xyz       `json:"rot"`
	Move         Xyz       `json:"move"`
	Fname        string    `json:"fname"`
	MikeClients  []Info    `json:"mike_clients"`  //打开麦克风的用户列表
	ShareClients []Info    `json:"share_clients"` // 打开屏幕分享的用户列表
	ShareId      string    `json:"share_id"`      //分享id
	DateTime     int64     `json:"dateteme"`      //时间戳(秒)
}

type Xyz struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
	Z float64 `json:"z"`
}

type Info struct {
	Id      string `json:"id"`
	SceneId string `json:"scene_id"`
	Fname   string `json:"fname"`
}

func (this *Service) Inits() {
	this.Broadcast = make(chan Message)
}

func (this *Service) StartBroadcast(msg Message) {
	Mu.Lock()
	defer Mu.Unlock()
	for uid, cli := range WsCliPool {
		if uid != msg.UserId && cli.WsConn != nil { //不发送给自己
			if msg.Operation == CHAT && msg.Range == 1 { //该消息为聊天并且范围在场景内
				if msg.SceneId == cli.ConnMsg.SceneId {
					cli.WriteMsgCh <- msg
					logger.Infof("broadcast chat msg sent to user:%s, msg:%v", uid, msg)
				}
			} else {
				cli.WriteMsgCh <- msg
				if msg.Operation != SYNC_POS {
					logger.Infof("broadcast none SYNC_POS msg sended to user:%s,msg:%v", uid, msg)
				}
			}
		}
	}
}

// 断开连接
func (this *Service) RemoveConn(uid string) {
	Mu.Lock()
	defer Mu.Unlock()
	delete(WsCliPool, uid)
}

func (this *Service) UpdateConnSceneId(userId string, msg Message) {
	Mu.Lock()
	defer Mu.Unlock()
	user, ok := WsCliPool[userId]
	if ok {
		user.ConnMsg.SceneId = msg.SceneId
		WsCliPool[userId] = user
	}
}

// 更新连接
func (this *Service) UpdateConn(newUserId string, newCli Client) {
	if newUserId == "" {
		return
	}
	Mu.Lock()
	defer Mu.Unlock()
	for loginUserId, orinCli := range WsCliPool { //给新客户端发送当前已连接用户的连接信息
		if newUserId != loginUserId {
			select {
			case newCli.WriteMsgCh <- orinCli.ConnMsg:
			case <-time.After(1 * time.Second):
				logger.Errorf("write msg channel blocked!userID:%s", newUserId)
			}
			continue
		}
		orinCli.ConnLock.Lock()
		if err := orinCli.WsConn.WriteJSON(&Message{
			UserId:    newUserId,
			Operation: OCCUPIED,
		}); err != nil {
			orinCli.ConnLock.Unlock()
			logger.Errorf("OCCUPIED msg send failed!userid:%s", newUserId)
			continue
		}
		// 当前userId对应原ws连接关闭
		if err := orinCli.WsConn.Close(); err != nil {
			orinCli.ConnLock.Unlock()
			logger.Errorf("UpdateConn() close original ws conn failed!userid:%s,error:%v", loginUserId, err)
			continue
		}
		orinCli.ConnLock.Unlock()
		func() {
			Mu.Lock()
			defer Mu.Unlock()
			select {
			case <-orinCli.CloseCh: //等被挤掉的原连接完全关闭后再更新新链接
			case <-time.After(5 * time.Second): //等待超时
				logger.Errorf("orin conn closing wait time out!orin userid:%s", orinCli.ConnMsg.UserId)
			}
		}()
	}
	WsCliPool[newUserId] = newCli
}
