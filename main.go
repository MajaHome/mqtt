package main

import (
	"flag"
	"github.com/MajaSuite/mqtt/broker"
	"github.com/MajaSuite/mqtt/packet"
	"log"
	"os"
	"os/signal"
	"syscall"
)

var (
	cert  = flag.String("cert", "broker.crt", "path to broker certificate")
	key   = flag.String("key", "broker.key", "path to broker private key")
	debug = flag.Bool("debug", false, "print debuging hex dumps")
)

func main() {
	flag.Parse()
	log.Println("Starting broker ...")

	if *debug {
		log.Println("DEBUG mode ON")
		packet.DEBUG = true
	}

	mqtt, err := broker.RunMqtt()
	if err != nil {
		log.Panic("error running mqtt broker", err)
	}

	engine := broker.NewEngine()
	engine.ManageServer(mqtt)

	finish := make(chan os.Signal, 1)
	signal.Notify(finish, syscall.SIGINT, syscall.SIGTERM)
	<-finish
}
