// https://gist.github.com/Xeoncross/89b5324d6c0c04699f5d
package main

import (
	"flag"
	"fmt"
	"github.com/gorilla/websocket"
	"io"
	"net/http"
	"runtime"
	"strconv"
	"time"
	"bytes"
)

var (
	message = flag.String("message", "Hello world", "message to send")
	service = flag.String("service", "localhost:8080/ws", "websocket address to access")
	numClients = flag.Int("clients", 20, "number of clients") //3000
	duration = flag.Duration("time", 20 * time.Second, "number of seconds to run everything")
)

var (
	// counters
	globalMessageSentChan chan int64
	globalMessageReadChan chan int64
	globalMessageCountChan chan int64
	globalMessageErrChan chan int64
	// shutting down
	globalStopSendChan chan int64
	globalStopReadChan chan int64
)

func createCounter() chan int64 {
	ch := make(chan int64)

	go func(c chan int64) {
		var counter int64 = 0
		for {
			c <- counter
			counter += 1
		}
	}(ch)

	return ch;
}

func main() {
	flag.Parse()
	runtime.GOMAXPROCS(1)
	//rand.Seed(time.Now().Unix())

	globalMessageSentChan = createCounter()
	globalMessageReadChan = createCounter()
	globalMessageCountChan = createCounter()
	globalMessageErrChan  = createCounter()
	globalStopSendChan    = make(chan int64)
	globalStopReadChan    = make(chan int64)

	for i := 0; i < *numClients; i++ {
		go createWebSocketClient(i)
	}

	fmt.Println("duration ", *duration)
	fmt.Println("clients ", *numClients)

	time.Sleep(*duration)
	fmt.Println("STOP SEND")
	for i :=0; i < *numClients; i++ {
		globalStopSendChan<-1
	}
	// give the sockets time to receive messages
	time.Sleep(500*time.Millisecond)

	fmt.Println("STOP RECEIVE")
	for i :=0; i < *numClients; i++ {
		globalStopReadChan<-1
	}
	time.Sleep(100*time.Millisecond)

	fmt.Println(<-globalMessageSentChan, " messages sent")
	fmt.Println(<-globalMessageReadChan, " messages read")
	fmt.Println(<-globalMessageCountChan, " messages accounted for")
	fmt.Println(<-globalMessageErrChan, " message read errors")
	fmt.Println()
	fmt.Println(runtime.NumCPU(), " cpus")
	fmt.Println(runtime.NumCgoCall(), " go calls")
	fmt.Println(runtime.NumGoroutine(), "go routines")
}

func readPump(ws *websocket.Conn, eCh chan<- error) { //, ch chan[]byte) {
	for {
		if _, data, err := ws.ReadMessage(); err != nil {
			eCh <- err
			return
		} else {
			numBytes := bytes.IndexByte(data,':')
			var xi int
			if numBytes >= 0 {
				xi, _ = strconv.Atoi(string(data[0:numBytes]))
			} else {
				xi = 1
			}
			<-globalMessageReadChan
			for i := 0 ; i < xi ; i++ {
				<-globalMessageCountChan
			}
			// ch <- slows things down. for metrics remove, for info keep
			// send data we received
			//ch <- data
		}
	}
}

func createWebSocketClient(id int) {
	url := "ws://" + *service
	msg := []byte(*message + "[" + strconv.Itoa(id) + "]")

	d := &websocket.Dialer{HandshakeTimeout: time.Duration(10 * time.Second)}
	ws, resp, err := d.Dial(url, http.Header{"Origin": {"http://localhost:8080"}})

	if err != nil {
		<-globalMessageErrChan
		fmt.Printf("Dial failed: %s\n (number: %v)\n%v", err.Error(), id, resp)
		return
	}

	defer ws.Close()
	//ch := make(chan []byte)
	eCh := make(chan error)
	// read from socket
	go readPump(ws, eCh) //, ch)

	// we're going to send 20 times / second
	sendMessageTimer := time.NewTicker(50 * time.Millisecond) //time.Duration(rand.Int31n(500)) * time.Millisecond)
	defer sendMessageTimer.Stop()
	for {
		select {
		// recieved data on the connection
		// case count := <-ch:
		// 	fmt.Println("<- ", count)
		case <- globalStopSendChan:
			sendMessageTimer.Stop()
		case <- globalStopReadChan:
			//fmt.Println("BYE")
			return
		// error received
		case err := <-eCh:
			// handle our error then exit for loop
			<-globalMessageErrChan

			if err != io.EOF { // simple closed connection
				fmt.Println(err, "receive", id)
			}
			return
		// This will timeout on the read.
		case <-sendMessageTimer.C:
			if err := ws.WriteMessage(websocket.TextMessage, msg); err != nil {
				fmt.Println(err, "send", id)
				<-globalMessageErrChan
				return
			} else {
				<-globalMessageSentChan
			}
		}
	}
}
