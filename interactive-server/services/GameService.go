package services

import (
	"context"
	"encoding/json"
	"interactive-server/clients"
	"interactive-server/logger"
	"interactive-server/pbs"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"google.golang.org/grpc/metadata"
)

type GameService struct {
	Service
}

var GameSvcBanCtrlr *sync.Map

func (this *GameService) Start(w http.ResponseWriter, r *http.Request) {
	this.Inits()
	var upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Errorf("ws conn upgrade failed!error: %v", err)
		return
	}
	defer c.Close()
	var msg Message
	lc := new(sync.Mutex)
	c.SetPingHandler(func(app string) (err error) { //心跳
		lc.Lock()
		err = c.WriteMessage(websocket.PongMessage, []byte(app))
		lc.Unlock()
		return
	})
	var msgBytes []byte
	_, msgBytes, err = c.ReadMessage()
	if err != nil {
		logger.Errorf("ws conn read failed!%v", err)
		return
	}
	logger.Infof("GameService new ws conn,first msg:%s", msgBytes)
	if err = json.Unmarshal(msgBytes, &msg); err != nil {
		logger.Errorf("%v", err)
		return
	}

	if len(msg.UserId) == 0 {
		lc.Lock()
		c.WriteMessage(websocket.CloseMessage, []byte("userid missing"))
		lc.Unlock()
		logger.Infof("ws chat conn first read failed!msg:%s", msgBytes)
		return
	}
	loginUserId := msg.UserId
	// 该client所有发出消息通过订阅该channel发送
	sendMsgCh := make(chan Message, 1000)
	defer close(sendMsgCh)
	closeSyncCh := make(chan int) // 同步关闭监听routine的channel
	defer func() {
		// 连接关闭前，通知后端服务更新用户登录时长
		if len(loginUserId) != 0 {
			_, err = pbs.NewUserServiceClient(clients.GrpcConn).Logout(
				metadata.NewOutgoingContext(context.Background(), metadata.New(map[string]string{"skiptoken": "jumpjumpjump"})),
				&pbs.UserLogoutReq{
					UserId: msg.UserId,
				})
			if err != nil {
				logger.Errorf("%v", err)
			}
			// 连接断开，广播离线消息
			msg.Operation = DISCONNECT
			this.RemoveConn(loginUserId)
			msg.UserId = loginUserId
			this.StartBroadcast(msg)
		}
		close(closeSyncCh)
	}()
	// 注册禁用禁言订阅channel到全局map
	var ch chan chBanInfo
	vch, exist := GameSvcBanCtrlr.Load(loginUserId)
	if !exist {
		ch = make(chan chBanInfo)
		GameSvcBanCtrlr.Store(loginUserId, ch)
	} else {
		ch = vch.(chan chBanInfo)
	}
	this.BanCh = ch
	// 监听禁用通知routine
	go func() {
		defer func() {
			if err := recover(); err != nil {
				logger.Errorf("ws subscribe routine paniced:error%v", err)
			}
		}()
		var err error
		for {
			select {
			case wMsg, opened := <-sendMsgCh: // 从channel订阅消息并发送
				if !opened {
					return
				}
				lc.Lock()
				if err = c.WriteJSON(&wMsg); err != nil {
					logger.Errorf("sending ws msg failed!userid: %s,msg: %v,error: %v", loginUserId, wMsg, err)
					lc.Unlock()
					return
				}
				lc.Unlock()
			case chInfo := <-this.BanCh: //禁用通知从外部触发
				logger.Infof("%s, userid:%s", chInfo.Reason, chInfo.UserId)
				var newMsg Message
				if chInfo.BanType == "0" {
					newMsg.Operation = BAN_TO_POST_RELEASE
					if chInfo.Ban {
						newMsg.Operation = BAN_TO_POST
					}
				} else {
					newMsg.Operation = BAN_RELEASE
					if chInfo.Ban {
						newMsg.Operation = BAN
					}
				}
				newMsg.UserId = chInfo.UserId
				newMsg.Text = chInfo.Reason
				sendMsgCh <- newMsg
				logger.Infof("ban msg sended to userid:%s,msg:%v", chInfo.UserId, newMsg)
				continue
			case <-closeSyncCh: //主逻辑退出
				logger.Infof("chatservice listen routine exit.userid:%s", loginUserId)
				return
			default:
			}
		}
	}()
	if msg.Operation == CONNECT { //广播上线
		this.UpdateConn(loginUserId, Client{ConnMsg: msg, WsConn: c, CloseCh: closeSyncCh, WriteMsgCh: sendMsgCh, ConnLock: lc})
	}
	this.StartBroadcast(msg)
	for {
		_, msgBytes, err = c.ReadMessage()
		if err != nil {
			logger.Errorf("%v", err)
			break
		}

		err = json2.Unmarshal(msgBytes, &msg)

		if err != nil {
			logger.Errorf("%v", err)
			break
		}
		if len(msg.UserId) == 0 {
			logger.Errorf("ws read msg user id missing,orin userid:%s", loginUserId)
			break
		}
		//todo
		if msg.Operation == CHAT {
			logger.Infof("incoming chat msg :%v", msg)
		}

		if msg.Operation == DISCONNECT { //广播离线
			break
		}

		if msg.Operation == ENTERSCENE { //切换场景
			this.UpdateConnSceneId(msg.UserId, msg)
		}
		this.StartBroadcast(msg)
	}
}

func init() {
	GameSvcBanCtrlr = new(sync.Map)
}
