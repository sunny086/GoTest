package main

import (
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
)

var randnum = rand.New(rand.NewSource(time.Now().Unix()))

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// Client is a middleman between the websocket connection and the hub.
type Client struct {

	// The websocket connection.
	conn *websocket.Conn

	// Buffered channel of outbound messages.
	send chan []byte

	Data *Data
}

// Data 定义传输的数据，
type Data struct {
	IP      string `json:"ip"`      // 本地ip
	Type    string `json:"type"`    // 类型
	User    string `json:"user"`    // 接受信息的用户名
	Form    string `json:"from"`    // 发送信息的用户名
	FormIP  string `json:"FromIp"`  // 发送信息的IP
	Content string `json:"content"` // 信息的内容。
}

func (c *Client) initWebsocketClient() {
	defer func() {
		h.unregister <- c
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
}

// readPump pumps messages from the websocket connection to the hub.
//
// The application runs readPump in a per-connection goroutine. The application
// ensures that there is at most one reader on a connection by executing all
// reads from this goroutine.
func (c *Client) readPump() {
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
		// 从websocket收到的数据是网页端用户输入的数据，属于data中的content
		c.Data.Content = string(message)
		// 将从c.data编码成json,传入到hub的broadcast通道。
		message, _ = json.Marshal(c.Data)
		h.broadcast <- message
		sendMessage := SendMessage{"user5", []byte("111111111111111")}
		h.sendMessage <- &sendMessage
	}
}

// writePump pumps messages from the hub to the websocket connection.
//
// A goroutine running writePump is started for each connection. The
// application ensures that there is at most one writer to a connection by
// executing all writes from this goroutine.
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod) // ticker是一个定时发送的定时器。
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			c.conn.WriteMessage(websocket.TextMessage, message)
		case <-ticker.C:
			//定时器,websocket的心跳机制。
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// serveWs handles websocket requests from the peer.
func serveWs(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	data := &Data{
		IP:      r.RemoteAddr,
		Type:    "handshake",
		User:    "user" + strconv.Itoa(randnum.Intn(10)),
		Content: "xxx上线了",
	}
	client := &Client{conn: conn, send: make(chan []byte, 256), Data: data}
	message, _ := json.Marshal(client.Data)
	conn.WriteMessage(websocket.TextMessage, message)
	h.register <- client

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go client.writePump()
	go client.readPump()
}

func (c *Client) SendBroadcast(message []byte) {
	h.broadcast <- message
}

func (c *Client) SendMessage(message *SendMessage) {
	h.sendMessage <- message
}
