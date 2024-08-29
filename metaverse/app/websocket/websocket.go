package websocket

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{}
var clients = make(map[*websocket.Conn]bool)
var broadcast = make(chan Message)

const (
	HeartbeatCheckTime = 10 // 心跳检测时间(s)
	HeartbeatTime      = 60 // 心跳距离上一次的最大时间(s)
)

type Client struct {
	Conn              *websocket.Conn
	Pool              *Pool
	LastHeartbeatTime int64
	Id                string
	SceneId           string
}

type Message struct {
	Id          string `json:"id"`
	Operation   int    `json:"operation"`
	SceneId     string `json:"sceneid"`
	UserModel   int    `json:"usermodel"`
	Pos         Xyz    `json:"pos"`
	Rot         Xyz    `json:"rot"`
	Move        Xyz    `json:"move"`
	Fname       string `json:"fname"`
	MikeClients []Info `json:"mike_clients"`
}

type Xyz struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
	Z float64 `json:"z"`
}

type Info struct {
	Id      string `json:"id"`
	SceneId string `json:"sceneid"`
}

func RunSocket() {
	http.HandleFunc("/socket", socketHandler)
	log.Println("启动websocket服务成功")
	go func() {
		log.Fatal(http.ListenAndServe("10.0.0.134:8050", nil))
	}()
	go HandleMessage()
}

func RunWebSocket() {
	pool := NewPool()
	go pool.Heartbeat()
	http.HandleFunc("/socket", func(w http.ResponseWriter, r *http.Request) {

	})
}

func HandleMessage() {
	for {
		messageData := <-broadcast
		for client := range clients {
			err := client.WriteJSON(messageData)
			if err != nil {
				log.Println("Error during message writing:", err)
				client.Close()
				delete(clients, client)
			}
		}
	}
}

func socketHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Error during connection upgradation:", err)
		return
	}
	defer conn.Close()
	clients[conn] = true
	for {
		var data Message
		err := conn.ReadJSON(&data)
		if err != nil {
			log.Println("Error during message reading:", err)
			delete(clients, conn)
			break
		}
		fmt.Println("data", data)
		broadcast <- data
	}
}

func (c *Client) Read() {
	for {
		var message Message
		err := c.Conn.ReadJSON(&message)
		if err != nil {
			log.Println("Error during message reading:", err)
			break
		}
		if message.Operation == 4 {
			message.MikeClients = c.Pool.MikeOnList
			c.Conn.WriteJSON(message)
			continue
		}
		c.Pool.Broadcast <- message
		c.LastHeartbeatTime = time.Now().Unix()
		c.Id = message.Id
		c.SceneId = message.SceneId
		if message.Operation == 2 { //打开麦克风
			c.Pool.MikeOnList = append(c.Pool.MikeOnList, Info{message.Id, message.SceneId})
			message.MikeClients = c.Pool.MikeOnList
		} else if message.Operation == 3 { //关闭麦克风
			var i = -1
			for index, val := range c.Pool.MikeOnList {
				if val.Id == message.Id {
					i = index
					break
				}
			}
			if i >= 0 {
				c.Pool.MikeOnList = append(c.Pool.MikeOnList[:i], c.Pool.MikeOnList[i+1:]...)
			}
		}
	}
}

func serveWs(pool *Pool, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	client := &Client{Pool: pool, Conn: conn, LastHeartbeatTime: time.Now().Unix()}
	client.Pool.Register <- client
	go client.Read()
}
