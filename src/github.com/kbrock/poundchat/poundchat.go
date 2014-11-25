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
	"github.com/kbrock/ezprof"
)

var (
	message = flag.String("message", "Hello world", "message to send")
	service = flag.String("service", "localhost:8080/ws", "websocket address to access")
	numClients = flag.Int("clients", 20, "number of clients") //3000
	duration = flag.Duration("time", 20 * time.Second, "number of seconds to run everything")
	cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")
	memprofile = flag.String("memprofile", "", "write memory profile to this file")
	sendFrequency = flag.Duration("sendfrequency", 50 * time.Millisecond, "how often to send messages")
	slowClient = flag.Int("slow", 0, "every x client is slow (0 = disable)")
	slowMultiplier = flag.Int("multiplier", 2, "How many frequencies does it sleep?")
	delay time.Duration
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

func createAllCounters() {
	globalMessageSentChan = createCounter()
	globalMessageReadChan = createCounter()
	globalMessageCountChan = createCounter()
	globalMessageErrChan  = createCounter()
	globalStopSendChan    = make(chan int64)
	globalStopReadChan    = make(chan int64)
}

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

func createClients() {
	fmt.Println()
	fmt.Println(*numClients," clients ", *duration)	

	for i := 0; i < *numClients; i++ {
		go createWebSocketClient(i)
	}
}

func removeClients() {
	fmt.Println(*numClients, " clients exiting 0.6s")
	go sendAll(*numClients, globalStopSendChan)
	time.Sleep(500 * time.Millisecond + delay)
	go sendAll(*numClients, globalStopReadChan)
	time.Sleep(100 * time.Millisecond + delay)
}

// send a message to all clients (to say stop sending / stop reading)
func sendAll(count int, c chan int64) {
	for i :=0; i < count; i++ {
		c <- 1
	}
}

func displayCounters() {
	fmt.Println()
	sent   := <-globalMessageSentChan
	totMsg := <-globalMessageCountChan
	expectedCount := int64(*numClients) * sent
	fmt.Println(sent, " msgs sent")
	fmt.Println(<-globalMessageReadChan, " packets received")
	if (totMsg == expectedCount) {
		fmt.Println(totMsg, " msgs received (match)")
	} else {
		fmt.Println(totMsg, " msgs received (MISMATCH)")
		fmt.Println(expectedCount, " msgs expected")
  }

	numErrors := <-globalMessageErrChan
	if (numErrors != 0) {
		fmt.Println()
		fmt.Println(numErrors, " errors")
	}
}

func displayStats() {
	fmt.Println()
	fmt.Println(runtime.NumCPU(), " cpus")
	fmt.Println(runtime.NumCgoCall(), " go calls")
	fmt.Println(runtime.NumGoroutine(), "go routines")
}

func main() {
	flag.Parse()
	runtime.GOMAXPROCS(1)
	//rand.Seed(time.Now().Unix())

	delay = *sendFrequency * time.Duration(*slowMultiplier)
	createAllCounters()
	ezprof.StartProfiler(*cpuprofile, *memprofile)
	createClients()
	time.Sleep(*duration)
	ezprof.CleanupProfiler(*cpuprofile, *memprofile)
	removeClients()
	displayCounters()
	displayStats()

}

func readPump(ws *websocket.Conn, eCh chan<- error, id int) { //, ch chan[]byte) {
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
			if (*slowClient != 0 && ((id % *slowClient) == 1)) {
				//blocking for a slow send
				if (id < 2) {
					fmt.Println("sleeping", id, "read", xi)
				}
				time.Sleep(delay)
				if (id < 2) {
					fmt.Println("slept   ", id)
				}
			} else {
				if (id < 2) {
					fmt.Println("reading ", id, "read", xi)
				}
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
	go readPump(ws, eCh, id) //, ch)

	// we're going to send 20 times / second
	sendMessageTimer := time.NewTicker(*sendFrequency) //time.Duration(rand.Int31n(500)) * time.Millisecond)
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
