package main

import (
	"github.com/seakee/go-api/app"
	"github.com/seakee/go-api/bootstrap"
	"log"
	"os"
	"os/signal"
	"runtime"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	config, err := app.LoadConfig()
	if err != nil {
		log.Fatal("Loading config error: ", err)
	}

	a, err := bootstrap.NewApp(config)
	if err != nil {
		log.Fatal("New App error: ", err)
	}

	a.Start()

	s := waitForSignal()
	log.Println("Signal received, app closed.", s)
}

func waitForSignal() os.Signal {
	signalChan := make(chan os.Signal, 1)
	defer close(signalChan)
	signal.Notify(signalChan, os.Kill, os.Interrupt)
	s := <-signalChan
	signal.Stop(signalChan)
	return s
}
