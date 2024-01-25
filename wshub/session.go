package wshub

import (
	"time"

	"github.com/gorilla/websocket"
)

type Session struct {
	id   string
	hub  *Hub
	conn *websocket.Conn
	send chan []byte
}

func newSession(id string, hub *Hub, conn *websocket.Conn) *Session {
	conn.SetReadLimit(wsMaxMsgSize)
	conn.SetReadDeadline(time.Now().Add(wsPongWait))
	conn.SetPongHandler(func(string) error { conn.SetReadDeadline(time.Now().Add(wsPongWait)); return nil })

	return &Session{id: id, hub: hub, conn: conn, send: make(chan []byte, wsMaxMsgSize)}
}

func (s *Session) run() {
	go s.read()
	go s.write()
}

func (s *Session) close() {
	s.conn.Close()
	close(s.send)
}

func (s *Session) read() {
	defer func() {
		s.hub.unregister <- s
	}()

	for {
		_, msg, err := s.conn.ReadMessage()
		if err != nil {
			break
		}

		s.hub.sub <- Message{s.id, msg}
	}
}

func (s *Session) write() {
	ticker := time.NewTicker(wsPingPeriod)
	defer func() {
		ticker.Stop()
	}()

	for {
		select {
		case message, ok := <-s.send:
			s.conn.SetWriteDeadline(time.Now().Add(wsWriteWait))
			if !ok {
				s.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := s.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			s.conn.SetWriteDeadline(time.Now().Add(wsWriteWait))
			if err := s.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
