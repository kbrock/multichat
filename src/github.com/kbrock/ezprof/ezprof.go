package ezprof

import (
  "log"
  "os"
  "os/signal"
  "runtime/pprof"
)

func CleanupProfiler(cpufilename, memfilename string) {
  if memfilename != "" {
    log.Printf("writing memory profiling")
    f, err := os.Create(memfilename)
    if err != nil {
      log.Fatal(err)
    }
    pprof.WriteHeapProfile(f)
    f.Close()
  }

  if cpufilename != "" {
    log.Printf("writing cpu profiling")
    pprof.StopCPUProfile()
  }
}

func StartProfiler(cpufilename, memfilename string) {
  if cpufilename != "" {
    log.Printf("starting cpu profiling")
      f, err := os.Create(cpufilename)
      if err != nil {
          log.Fatal(err)
      }
      pprof.StartCPUProfile(f)
  }

  c := make(chan os.Signal, 1)
  signal.Notify(c, os.Interrupt)
  go func(c chan os.Signal, cpufilename, memfilename string) {
    for sig := range c {
      log.Printf("captured %v, stopping profiler and exiting..", sig)
      CleanupProfiler(cpufilename, memfilename)
      os.Exit(1)
    }
  }(c, cpufilename, memfilename)
}
