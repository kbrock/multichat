package main

import (
	//"code.google.com/p/go.net/websocket"
	"fmt"
	"github.com/gorilla/websocket"
	"io"
	"math/rand"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"sync"
	"time"
)

const message = "Hello world"

var service string

// I want to know how much communication happened
var globalMessageSentChan chan int64
var globalMessageReadChan chan int64
var globalMessageErrChan chan int64

// Atomic counters using channels
func makeCounterChannels() {

	globalMessageSentChan = make(chan int64)
	go func() {
		var counter int64 = 0
		for {
			globalMessageSentChan <- counter
			counter += 1
		}
	}()

	globalMessageReadChan = make(chan int64)
	go func() {
		var counter int64 = 0
		for {
			globalMessageReadChan <- counter
			counter += 1
		}
	}()

	globalMessageErrChan = make(chan int64)
	go func() {
		var counter int64 = 0
		for {
			globalMessageErrChan <- counter
			counter += 1
		}
	}()
}

var done sync.WaitGroup

func main() {

	runtime.GOMAXPROCS(1)

	service = "localhost:8080/ws"

	if len(os.Args) == 2 {
		service = os.Args[1]
	} else {
		fmt.Println("Usage: ", os.Args[0], "localhost:8080")
	}

	defer func() {
		fmt.Printf("NumCPU:", runtime.NumCPU(),
			"NumCgoCall:", runtime.NumCgoCall(),
			"NumGoRoutine:", runtime.NumGoroutine(),
		)
	}()

	makeCounterChannels()
	rand.Seed(time.Now().Unix())

	// Create X clients and randomly assign them a "meetingId"
	// meetingID = chat room/channel name
	// This doesn't matter for normal demo servers that only have 1 room/meeting

	//endChan := make(chan bool)

	for i := 0; i < 3000; i++ {

		done.Add(1)

		func(id, meetingId string) {
			go createWebSocketClient(id, meetingId)
		}(strconv.Itoa(i), strconv.Itoa(i%10))

	}

	fmt.Println("Waiting on done.Wait()")
	done.Wait()
	fmt.Println("run another 20 seconds...")

	//close(endChan)

	// We let the test run for a bit and then kill everything
	time.Sleep(10 * time.Second)

	fmt.Println(<-globalMessageSentChan, " messages sent")
	fmt.Println(<-globalMessageReadChan, " messages read")
	fmt.Println(<-globalMessageErrChan, " message read errors")

	os.Exit(1)
}

func createWebSocketClient(id, meetingId string) {

	var loaded bool = false

	//defer done.Done()

	defer func() {
		//fmt.Println(meetingId, id, "closing")
		if !loaded {
			fmt.Println(id, "failed to work")
			done.Done()
		}
	}()

	int_id, _ := strconv.Atoi(id)

	if (int_id % 1000) == 0 {
		fmt.Println(id)
	}

	//url := service + "?meetingId=" + meetingId
	url := "ws://" + service

	// go.net
	//ws, err := websocket.Dial("ws://"+url, "", "http://"+url)

	// gorilla
	req := http.Header{"Origin": {"http://localhost:8080"}}
	d := &websocket.Dialer{HandshakeTimeout: time.Duration(10 * time.Second)}
	ws, resp, err := d.Dial(url, req)

	if err != nil {
		fmt.Printf("Dial failed: %s\n (number: %v)\n%v", err.Error(), id, resp)
		//os.Exit(1)
		return
	}

	defer ws.Close()

	ch := make(chan []byte)
	eCh := make(chan error)

	// Start a goroutine to read from our net connection
	go func(ch chan []byte, eCh chan error) {
		for {

			<-globalMessageReadChan

			// try to read the data
			//data := make([]byte, 512)
			//_, err := ws.Read(data)

			_, data, err := ws.ReadMessage()

			if err != nil {
				// send an error if it's encountered
				eCh <- err
				return
			}

			// send data if we read some.
			ch <- data
		}
	}(ch, eCh)

	//endtimer := time.NewTimer(time.Duration(60) * time.Second)

	for {

		timer := time.NewTimer(time.Duration(rand.Int31n(500)) * time.Millisecond)

		select {
		// This case means we recieved data on the connection
		case _ = <-ch:

			if !loaded {
				loaded = true
				done.Done()
			}

			// Do something with the data
			// This case means we got an error and the goroutine has finished
		case err := <-eCh:
			// handle our error then exit for loop
			<-globalMessageErrChan

			if err != io.EOF { // simple closed connection
				fmt.Println(err, meetingId, id)
			}

			return
		// This will timeout on the read.
		case <-timer.C:
			// do nothing? this is just so we can time out if we need to.
			// you probably don't even need to have this here unless you want
			// do something specifically on the timeout.

			<-globalMessageSentChan // +1 for the writes

			msg := message + " " + meetingId + " " + id //+  " " + strconv.Itoa(i)

			//if _, err := ws.Write([]byte(msg)); err != nil {
			if err := ws.WriteMessage(websocket.TextMessage, []byte(msg)); err != nil {
				fmt.Println(err, "meeting", meetingId, id)
				return
			}

			// case <-endtimer.C:
			// 	fmt.Printf("ending %v %v", meetingId, id)
			// 	return
		}
	}
}
