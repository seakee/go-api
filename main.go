// Copyright 2024 Seakee.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package main

import (
	"github.com/seakee/go-api/app/config"
	"log"
	"os"
	"os/signal"
	"runtime"

	"github.com/seakee/go-api/bootstrap"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal("Loading cfg error: ", err)
	}

	a, err := bootstrap.NewApp(cfg)
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
