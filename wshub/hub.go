package wshub

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/permadao/permaswap/logger"
)

const (
	PubAckOK     = "ok"
	PubAckFailed = "failed"

	// Time allowed to write a message to the peer.
	wsWriteWait = 5 * time.Second
	// Maximum message size allowed from peer.
	wsMaxMsgSize = 1024 * 100

	// Time allowed to read the next pong message from the peer.
	wsPongWait   = 10 * time.Second
	wsPingPeriod = (wsPongWait * 9) / 10
)

var log = logger.New("wshub")

type Message struct {
	ID   string // session id
	Data []byte
}

type Hub struct {
	pub chan Message
	sub chan Message

	register   chan *Session
	unregister chan *Session
	sessions   map[string]*Session
}

func New() *Hub {
	return &Hub{
		pub: make(chan Message),
		sub: make(chan Message),

		register:   make(chan *Session),
		unregister: make(chan *Session),
		sessions:   make(map[string]*Session),
	}
}

func (h *Hub) Run(registerFunc, unregisterFunc func(id string)) {
	go func(unregisterFunc func(id string)) {
		for {
			select {
			case msg := <-h.pub:
				if ses, ok := h.sessions[msg.ID]; ok {
					ses.send <- msg.Data
				}
			case ses := <-h.register:
				h.sessions[ses.id] = ses

				if registerFunc != nil {
					go registerFunc(ses.id)
				}
			case ses := <-h.unregister:
				id := ses.id
				if unregisterFunc != nil {
					go unregisterFunc(id)
				}

				delete(h.sessions, ses.id)
				go ses.close()
			}
		}
	}(unregisterFunc)
}

func (h *Hub) RegisterSession(w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{
		ReadBufferSize:  wsMaxMsgSize,
		WriteBufferSize: wsMaxMsgSize,
		CheckOrigin:     func(r *http.Request) bool { return true },
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error("can not upgrade websocket connection", "err", err)
		return
	}

	s := newSession(uuid.NewString(), h, conn)
	s.run()

	h.register <- s
}

func (h *Hub) Subscribe() <-chan Message {
	return h.sub
}

func (h *Hub) Publish(id string, data []byte) {
	h.pub <- Message{id, data}
}

func (h *Hub) CloseSession(sessionID string) {
	if ses, ok := h.sessions[sessionID]; ok {
		log.Info("closing ws session", "sessionID", sessionID)
		go ses.conn.Close()
	}
}
