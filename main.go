package main

import (
	"aphoteka_scraper/telegram"
	"flag"
	"log"
	"os"
	"runtime"
	"runtime/pprof"
)

var memprofile = flag.String("memprofile", "", "write memory profile to `file`")

func main() {
	flag.Parse()

	err := telegram.RunServer()
	if err != nil {
		log.Print(err)
	}
	log.Print("Server shut down")

	if *memprofile != "" {
		f, err := os.Create(*memprofile)
		if err != nil {
			log.Fatal("could not create memory profile: ", err)
		}
		defer f.Close()
		runtime.GC() // get up-to-date statistics
		if err := pprof.WriteHeapProfile(f); err != nil {
			log.Fatal("could not write memory profile: ", err)
		}
	}
}
