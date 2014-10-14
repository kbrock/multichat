// https://gist.github.com/Xeoncross/89b5324d6c0c04699f5d
package main

import (
//  "bytes"
	"flag"
	"fmt"
	"github.com/gorilla/websocket"
	"io"
//	"math/rand"
	"net/http"
	"runtime"
	"strconv"
//	"sync"
	"time"
)

var (
	message = flag.String("message", "Hello world", "message to send")
	service = flag.String("service", "localhost:8080/ws", "websocket address to access")
	numClients = flag.Int("clients", 20, "number of clients") //3000
	duration = flag.Duration("time", 20 * time.Second, "number of seconds to run everything")
)

// Counters using channels
var globalMessageSentChan chan int64
var globalMessageReadChan chan int64
var globalMessageErrChan chan int64

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

// ensure all websockets have received their own messages
//var done sync.WaitGroup

func main() {
	flag.Parse()
	runtime.GOMAXPROCS(1)
	//rand.Seed(time.Now().Unix())

	globalMessageSentChan = createCounter()
	globalMessageReadChan = createCounter()
	globalMessageErrChan = createCounter()
	for i := 0; i < *numClients; i++ {
//		done.Add(1)
		go createWebSocketClient(i)
	}

//	fmt.Println("Waiting for messages to be sent")
//	done.Wait()
	fmt.Println("duration ", *duration)
	fmt.Println("clients ", *numClients)

	// We let the test run for a bit and then kill everything
	time.Sleep(*duration)

	fmt.Println(<-globalMessageSentChan, " messages sent")
	fmt.Println(<-globalMessageReadChan, " messages read")
	fmt.Println(<-globalMessageErrChan, " message read errors")
	fmt.Println()
	fmt.Println(runtime.NumCPU(), " cpus")
	fmt.Println(runtime.NumCgoCall(), " go calls")
	fmt.Println(runtime.NumGoroutine(), "go routines")
}

func createWebSocketClient(id int) {
	url := "ws://" + *service
	msg := []byte("" + *message + " " + strconv.Itoa(id))

	d := &websocket.Dialer{HandshakeTimeout: time.Duration(10 * time.Second)}
	ws, resp, err := d.Dial(url, http.Header{"Origin": {"http://localhost:8080"}})

	if err != nil {
		fmt.Printf("Dial failed: %s\n (number: %v)\n%v", err.Error(), id, resp)
		return
	}

	// var loaded bool = false
	// defer func() {
	// 	if !loaded {
	// 		fmt.Println(id, "failed to work")
	// 		done.Done()
	// 	}
	// }()
	defer ws.Close()

	ch := make(chan []byte)
	eCh := make(chan error)

	// Start a goroutine to read from our net connection
	go func(ch chan []byte, eCh chan error) {
		for {
			<-globalMessageReadChan
			if _, data, err := ws.ReadMessage(); err != nil {
				eCh <- err
				return
			} else {
				// send data we received
				ch <- data
			}
		}
	}(ch, eCh)

	for {
		timer := time.NewTimer(50 * time.Millisecond) //time.Duration(rand.Int31n(500)) * time.Millisecond)

		select {
		// recieved data on the connection
		case _ = <-ch:
			//if ( bytes.Equal(msg, newMsg)){
			// if !loaded {
			// 	loaded = true
			// 	done.Done()
			// }

		// Do something with the data
		// This case means we got an error and the goroutine has finished
		case err := <-eCh:
			// handle our error then exit for loop
			<-globalMessageErrChan

			if err != io.EOF { // simple closed connection
				fmt.Println(err, "receive", id)
			}

			return
		// This will timeout on the read.
		case <-timer.C:
			// do nothing? this is just so we can time out if we need to.
			// you probably don't even need to have this here unless you want
			// do something specifically on the timeout.

			<-globalMessageSentChan

			//if _, err := ws.Write([]byte(msg)); err != nil {
			if err := ws.WriteMessage(websocket.TextMessage, msg); err != nil {
  			fmt.Println(err, "send", id)
				return
			}

			// case <-endtimer.C:
			// 	fmt.Printf("ending %v", id)
			// 	return
		}
	}
}
