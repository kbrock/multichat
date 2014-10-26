// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"time"
	"sync"
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
	send chan *message
}

// assign a random "name" to each new connection
var debugCounter = 0
func NewConnection(ws *websocket.Conn) *connection {
	name := strconv.Itoa(debugCounter)
  debugCounter += 1
	c := connection{
		name: name,
		ws: ws,
		send: make(chan *message, 256),
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
			h.broadcast <- &message{sender:c.name,bytes:msg}
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

// writePump pumps messages from the hub to the websocket connection.
func (c *connection) writePump() {
	deferTicker := time.NewTicker(deferPeriod)
	pingTicker := time.NewTicker(pingPeriod)

	var mutex = &sync.Mutex{}
	var outstanding *message = nil

	defer func() {
		pingTicker.Stop()
		deferTicker.Stop()
		c.ws.Close()
	}()

	for {
		select {
		case msg, ok := <-c.send:
			if !ok {
				c.write(websocket.CloseMessage, []byte{})
				return
			}
			mutex.Lock()
			if outstanding != nil{
				outstanding.merge(msg)
			} else {
				outstanding = msg.clone()
			}
			mutex.Unlock()
		case <-deferTicker.C:
			mutex.Lock()
			var tosend *message = nil
			if outstanding != nil {
		 		tosend = outstanding
		 		outstanding = nil
			}
			mutex.Unlock()
			if (tosend != nil) {
	 			if err := c.write(websocket.TextMessage, tosend.allBytes()); err != nil {
		 			return
			 	}
			}
		case <-pingTicker.C:
			if err := c.write(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		}
	}
}

// serverWs handles websocket requests from the peer.
func serveWs(w http.ResponseWriter, r *http.Request) {
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
	c.readPump()
}
