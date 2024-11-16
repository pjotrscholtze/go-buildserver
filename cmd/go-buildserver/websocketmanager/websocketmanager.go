package websocketmanager

import (
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write the file to the client.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the client.
	pongWait = 3 * time.Second

	// Send pings to client with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10
)

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
)

type Connection struct {
	knownDead      bool
	lastPing       *time.Time
	channel        chan interface{}
	connectionLock *sync.Mutex
	ws             *websocket.Conn
}

func (c *Connection) reader() {
	defer func() {
		c.connectionLock.Lock()
		defer c.connectionLock.Unlock()
		c.knownDead = true
		c.ws.Close()
	}()
	c.ws.SetReadLimit(512)
	c.ws.SetReadDeadline(time.Now().Add(pongWait))
	c.ws.SetPongHandler(func(test string) error {
		c.connectionLock.Lock()
		defer c.connectionLock.Unlock()
		now := time.Now()
		c.lastPing = &now
		c.ws.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})
	for {
		_, _, err := c.ws.ReadMessage()
		if err != nil {
			return
		}
	}
}
func (c *Connection) writer(lastMessage interface{}) {
	pingTicker := time.NewTicker(pingPeriod)
	// fileTicker := time.NewTicker(filePeriod)

	defer func() {
		pingTicker.Stop()
		// fileTicker.Stop()
		c.ws.Close()
	}()
	if lastMessage != nil {
		c.ws.WriteJSON(lastMessage)
	}
	closeChannel := make(chan struct {
		code int
		text string
	})
	c.ws.SetCloseHandler(func(code int, text string) error {
		//
		log.Println("closing. code:", code, ", text: ", text)
		closeChannel <- struct {
			code int
			text string
		}{code: code, text: text}
		return nil
	})

	for {
		select {
		case data := <-closeChannel:
			log.Println("closing (via channel). code:", data.code, ", text: ", data.text)
			c.knownDead = true
			return
		case data := <-c.channel:
			c.ws.WriteJSON(data)
		case <-pingTicker.C:
			c.ws.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.ws.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(writeWait)); err != nil {
				log.Println("ping:", err)
				c.knownDead = true
				return
			}
		}
	}

}
func (c *Connection) Init(lastMessage interface{}) {
	// go wn.writer(ws, endpoint, id)
	go c.writer(lastMessage)
	c.reader()

}

func (c *Connection) ConnectionIsAlive() bool {
	c.connectionLock.Lock()
	defer c.connectionLock.Unlock()
	if c.knownDead {
		return false
	}
	if c.lastPing == nil {
		return true
	}
	delta := time.Now().Sub(*c.lastPing)
	return delta < pongWait
}

func (c *Connection) Send(data interface{}) {
	c.connectionLock.Lock()
	defer c.connectionLock.Unlock()
	c.channel <- data
}

type WebsocketChannel struct {
	connectionsLock *sync.Mutex
	lastMessage     interface{}
	connections     []*Connection
}

func (wc *WebsocketChannel) Register(ws *websocket.Conn) {
	wc.connectionsLock.Lock()
	defer wc.connectionsLock.Unlock()
	conn := &Connection{
		knownDead:      false,
		lastPing:       nil,
		channel:        make(chan interface{}),
		connectionLock: &sync.Mutex{},
		ws:             ws,
	}
	wc.connections = append(wc.connections, conn)
	go conn.Init(wc.lastMessage)
}

func (wc *WebsocketChannel) Broadcast(data interface{}) {
	wc.connectionsLock.Lock()
	defer wc.connectionsLock.Unlock()
	cleanedConnections := make([]*Connection, 0)
	for _, conn := range wc.connections {
		if conn != nil && conn.ConnectionIsAlive() {
			cleanedConnections = append(cleanedConnections, conn)
			conn.Send(data)
		}
	}
	wc.connections = cleanedConnections
}

type WebsocketManager struct {
	channels     map[string]map[string]*WebsocketChannel
	channelsLock *sync.Mutex
}

func (wsm *WebsocketManager) getChannel(endpoint, id string) *WebsocketChannel {
	wsm.channelsLock.Lock()
	defer wsm.channelsLock.Unlock()
	if _, ok := wsm.channels[endpoint]; !ok {
		wsm.channels[endpoint] = make(map[string]*WebsocketChannel)
	}
	if _, ok := wsm.channels[endpoint][id]; !ok {
		wsm.channels[endpoint][id] = &WebsocketChannel{
			connectionsLock: &sync.Mutex{},
			lastMessage:     nil,
			connections:     make([]*Connection, 0),
		}
	}
	return wsm.channels[endpoint][id]
}

func (wsm *WebsocketManager) BroadcastOnEndpoint(endpoint, id string, data interface{}) {
	channel := wsm.getChannel(endpoint, id)
	if channel == nil {
		// Shouldnt happen
		return
	}
	channel.Broadcast(data)
}

func (wsm *WebsocketManager) Setup(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		if _, ok := err.(websocket.HandshakeError); !ok {
			log.Println(err)
		}
		return
	}
	urlParts := strings.Split(r.RequestURI[4:], "/")
	endpoint := urlParts[0]
	id := ""
	if len(urlParts) > 1 {
		id = urlParts[1]
	}
	channel := wsm.getChannel(endpoint, id)
	channel.Register(ws)

}

func NewWebsocketManager() WebsocketManager {
	return WebsocketManager{
		channels:     make(map[string]map[string]*WebsocketChannel),
		channelsLock: &sync.Mutex{},
	}
}
