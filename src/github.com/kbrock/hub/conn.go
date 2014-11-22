package hub

import (
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"time"
	"strconv"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	deferPeriod = 250 * time.Millisecond

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// connection is an middleman between the websocket connection and the hub.
type connection struct {
	name string
	// The websocket connection.
	ws *websocket.Conn

	// Buffered channel of outbound messages.
	//spray2coalesce
	send chan *message
	// channel of message ready to send to websocket
	//coalesce2receiver
	buffer chan *message
}

// assign a random "name" to each new connection
var debugCounter = 0
func NewConnection(ws *websocket.Conn) *connection {
	name := strconv.Itoa(debugCounter)
	debugCounter += 1
	c := connection{
		name: name,
		ws: ws,
		// GOAL: no buffering here
		send: make(chan *message, 256),
		buffer: make (chan *message),
	}
	return &c
}

// readPump pumps messages from the websocket connection to the hub.
func (c *connection) readPump() {
	defer func() {
		h.unregister <- c
		c.ws.Close()
	}()
	c.ws.SetReadLimit(maxMessageSize)
	c.ws.SetReadDeadline(time.Now().Add(pongWait))
	c.ws.SetPongHandler(func(string) error { c.ws.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		if _, msg, err := c.ws.ReadMessage() ; err == nil {
			h.broadcast <- NewMessage(c.name, msg)
		} else {
			break
		}
	}
}

// write writes a message with the given message type and payload.
func (c *connection) write(mt int, payload []byte) error {
	c.ws.SetWriteDeadline(time.Now().Add(writeWait))
	return c.ws.WriteMessage(mt, payload)
}

// aggregator takes messages from broadcast to buffer
func (c *connection) aggregator() {
  msg := EmptyMessage()
  timer := time.NewTimer(0)

  var timerCh <-chan time.Time
  var outCh chan *message

  for {
    select {
    case e, ok := <-c.send:
		// un registered, time to go away
		if !ok {
			close(c.buffer)
			timer.Stop()
			return
		}
      msg = msg.Merge(e)
      if timerCh == nil {
        timer.Reset(deferPeriod)
        timerCh = timer.C
      }
    case <-timerCh:
      outCh = c.buffer
      timerCh = nil
    case outCh <- msg:
      msg = EmptyMessage()
      outCh = nil
    }
  }
}

// writePump pumps messages from the buffer to the websocket connection.
func (c *connection) writePump() {
	pingTicker := time.NewTicker(pingPeriod)

	defer func() {
		pingTicker.Stop()
		c.ws.Close()
	}()

	for {
		select {
		case msg, ok := <-c.buffer:
			if !ok {
				c.write(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.write(websocket.TextMessage, msg.Bytes()); err != nil {
				return
			}
		case <-pingTicker.C:
			if err := c.write(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		}
	}
}

// serveWs handles websocket requests from the peer.
func ServeWs(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		return
	}
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	c := NewConnection(ws)
	h.register <- c
	go c.writePump()
	go c.aggregator()
	c.readPump()
}
