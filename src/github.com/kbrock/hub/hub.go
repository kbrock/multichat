package hub

// hub maintains the set of active connections and broadcasts messages to the
// connections.
type hub struct {
	// Registered connections.
	connections map[*connection]bool

	// Inbound messages from the connections.
	broadcast chan *message

	// Register requests from the connections.
	register chan *connection

	// Unregister requests from connections.
	unregister chan *connection

	quit chan int
}

func RunHub() {
	go h.run()
}

func QuitHub() {
	h.quit = make(chan int)
	h.quit <- 1
}

var h = hub{
	broadcast:   make(chan *message),
	register:    make(chan *connection),
	unregister:  make(chan *connection),
	connections: make(map[*connection]bool),
}

func (h *hub) run() {
	for {
		select {
		case c := <-h.register:
			h.connections[c] = true
		case c := <-h.unregister:
			// if we still have this connection, remove it
			if _, ok := h.connections[c]; ok {
				delete(h.connections, c)
				close(c.send)
			}
		case m := <-h.broadcast:
			for c := range h.connections {
				// NOTE: blocking on send
				c.send <- m
			}
		case <-h.quit:
			return
		}
	}
}
