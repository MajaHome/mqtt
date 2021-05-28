package main

import (
	"flag"
	"log"
	"os"
	"syscall"
	"os/signal"
	"github.com/MajaSuite/mqtt/broker"
)

var (
	cert = flag.String("cert", "broker.crt", "path to broker certificate")
	key  = flag.String("key", "broker.key", "path to broker private key")
)

func main() {
	flag.Parse()
	log.Println("Starting broker ...")

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
