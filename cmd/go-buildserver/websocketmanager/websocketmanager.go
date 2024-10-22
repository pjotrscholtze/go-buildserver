package websocketmanager

import (
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type wsConn struct {
	listeners     []time.Time
	listenersLock sync.Mutex
	lastMsg       interface{}
	channel       chan interface{}
}

func (wc *wsConn) addListener() {
	wc.listenersLock.Lock()
	defer wc.listenersLock.Unlock()
	wc.listeners = append(wc.listeners, time.Now())
}

func (wc *wsConn) getListeners() int {
	wc.listenersLock.Lock()
	defer wc.listenersLock.Unlock()
	keep := []time.Time{}
	for i := range wc.listeners {
		delta := time.Now().Sub(wc.listeners[i])
		if delta < pongWait {
			keep = append(keep, wc.listeners[i])
		}
	}
	wc.listeners = keep
	return len(keep)
}

type WebsocketManager struct {
	// Unbuffered, since we are not expecting huge amounts of traffic.
	channels map[string]map[string](*wsConn)
}

const (
	// Time allowed to write the file to the client.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the client.
	pongWait = 3 * time.Second

	// Send pings to client with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Poll file for changes with this period.
	filePeriod = 10 * time.Second
)

var (
	filename = "../../release.txt"
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
)

func reader(ws *websocket.Conn) {
	defer ws.Close()
	ws.SetReadLimit(512)
	ws.SetReadDeadline(time.Now().Add(pongWait))
	ws.SetPongHandler(func(test string) error {
		ws.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})
	for {
		_, _, err := ws.ReadMessage()
		if err != nil {
			break
		}
	}
}

func (wn *WebsocketManager) writer(ws *websocket.Conn, endpoint, id string) {
	pingTicker := time.NewTicker(pingPeriod)
	// fileTicker := time.NewTicker(filePeriod)

	defer func() {
		pingTicker.Stop()
		// fileTicker.Stop()
		ws.Close()
	}()
	chnlMgr := wn.getChannel(endpoint, id)
	chnlMgr.addListener()
	ws.WriteJSON(chnlMgr.lastMsg)

	for {
		select {
		case data := <-chnlMgr.channel:
			ws.WriteJSON(data)
		case <-pingTicker.C:
			ws.SetWriteDeadline(time.Now().Add(writeWait))
			// if err := ws.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
			// 	return
			// }
			if err := ws.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(writeWait)); err != nil {
				log.Println("ping:", err)
			}
			chnlMgr.addListener()
		}
	}
}

func (wn *WebsocketManager) getChannel(endpoint, id string) *wsConn {
	if _, ok := wn.channels[endpoint]; !ok {
		wn.channels[endpoint] = make(map[string]*wsConn)
	}
	if _, ok := wn.channels[endpoint][id]; !ok {
		wn.channels[endpoint][id] = &wsConn{
			listeners:     []time.Time{},
			listenersLock: sync.Mutex{},
			channel:       make(chan interface{}),
			lastMsg:       nil,
		}
	}
	return wn.channels[endpoint][id]
}
func (wn *WebsocketManager) BroadcastOnEndpoint(endpoint, id string, data interface{}) {
	conn := wn.getChannel(endpoint, id)
	listners := conn.getListeners()
	conn.lastMsg = data
	if listners > 0 {
		conn.channel <- data
	}
}
func (wn *WebsocketManager) Setup(w http.ResponseWriter, r *http.Request) {
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

	go wn.writer(ws, endpoint, id)
	reader(ws)
}

func NewWebsocketManager() WebsocketManager {
	return WebsocketManager{
		channels: make(map[string]map[string]*wsConn),
	}
}
