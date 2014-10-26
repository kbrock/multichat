package main

import (
	"flag"
  "go/build"
	"net/http"
 	"log"
  "github.com/kbrock/hub"
  "github.com/kbrock/ezprof"
)

var (
	addr    = flag.String("addr", ":8080", "http service address")
  webroot = flag.String("root", defaultRoot(), "path to webroot")
  cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")
  memprofile = flag.String("memprofile", "", "write memory profile to this file")
)

// thanks gary.burd.info/go-websocket-chat
func defaultRoot() string {
  if p, err := build.Default.Import("github.com/kbrock/multichat", "", build.FindOnly) ; err == nil {
    return p.Dir+"/webroot"
   } else {
     return "./webroot"
   }
}

func main() {
	flag.Parse()
  ezprof.StartProfiler(*cpuprofile, *memprofile)
	hub.RunHub()
	http.HandleFunc("/ws", hub.ServeWs)
  http.Handle("/", http.FileServer(http.Dir(*webroot)))
  log.Println("listening on", *addr)
	err := http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
  ezprof.CleanupProfiler(*cpuprofile, *memprofile)
}
