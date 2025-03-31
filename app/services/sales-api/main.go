package main

import (
	"log"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	_ "go.uber.org/automaxprocs"
	"go.uber.org/automaxprocs/maxprocs"
)

var build = "develop"

func main() {
	// set the correct amount of threads for the service
	// based on CPU availability or quota
	if _, err := maxprocs.Set(); err != nil {
		log.Fatal("maxprocs failed to set GOMAXPROCS", err)
	}

	g := runtime.GOMAXPROCS(0)
	log.Printf("starting service build[%s] CPU[%d]", build, g)
	defer log.Println("service stoped")

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGTERM, syscall.SIGINT)

	<-shutdown
	log.Println("stopping service....")
}
