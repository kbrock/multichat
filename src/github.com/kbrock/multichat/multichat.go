// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
  "go/build"
	"net/http"
 	"log"
)

var (
	addr    = flag.String("addr", ":8080", "http service address")
  webroot = flag.String("root", defaultRoot(), "path to webroot")
)
//var homeTempl = template.Must(template.ParseFiles("home.html"))

// thanks gary.burd.info/go-websocket-chat
func defaultRoot() string {
  p, err := build.Default.Import("github.com/kbrock/multichat", "", build.FindOnly)
  if err == nil {
    return p.Dir+"/webroot"
   } else {
     return "./webroot"
   }
}

func main() {
	flag.Parse()
	go h.run()
	http.HandleFunc("/ws", serveWs)
  http.Handle("/", http.FileServer(http.Dir(*webroot)))
  log.Println("listening on", *addr)
	err := http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
