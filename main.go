package main

import (
	"flag"
	"log"
	"mqtt/server"
	"os"
	"os/signal"
	"syscall"
)

var (
	cert = flag.String("cert", "server.crt", "path to server certificate")
	key  = flag.String("key", "server.key", "path to server private key")
)

func main() {
	flag.Parse()
	log.Println("Starting broker ...")

	// init
	engine := server.NewEngine()

	mqtt, err := server.RunMqtt()
	if err != nil {
		log.Panic("error running mqtt server", err)
	}

	go engine.Process(mqtt)

	finish := make(chan os.Signal, 1)
	signal.Notify(finish, syscall.SIGINT, syscall.SIGTERM)
	<-finish
}
