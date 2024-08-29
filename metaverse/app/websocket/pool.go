package websocket

import (
	"fmt"
	"log"
	"time"
)

type Pool struct {
	Clients    map[*Client]bool
	Broadcast  chan Message
	Register   chan *Client
	Unregister chan *Client
	MikeOnList []Info
}

func NewPool() *Pool {
	return &Pool{
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Clients:    make(map[*Client]bool),
		Broadcast:  make(chan Message),
		MikeOnList: make([]Info, 0),
	}
}

func (pool *Pool) Start() {
	for {
		select {
		case client := <-pool.Register:
			pool.Clients[client] = true
		case client := <-pool.Unregister:
			fmt.Println("断开连接")
			if _, ok := pool.Clients[client]; ok {
				delete(pool.Clients, client)
			}
			for c := range pool.Clients {
				if err := c.Conn.WriteJSON(&Message{Id: client.Id, Operation: 1, SceneId: client.SceneId}); err != nil {
					log.Println("Error during message writing when unregister:", err)
					continue
				}
			}
		case message := <-pool.Broadcast:
			for client := range pool.Clients {
				if err := client.Conn.WriteJSON(message); err != nil {
					log.Println("Error during message writing:", err)
					break
				}
			}
		}
	}
}

// 心跳检测
func (pool *Pool) Heartbeat() {
	for {
		for client := range pool.Clients {
			if time.Now().Unix()-client.LastHeartbeatTime > HeartbeatTime {
				client.Pool.Unregister <- client
			}
		}
		time.Sleep(time.Second * HeartbeatCheckTime)
	}
}
