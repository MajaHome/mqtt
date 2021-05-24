package main

import (
	"flag"
	"log"
	"github.com/MajaSuite/mqtt/server"
	"github.com/MajaSuite/mqtt/transport"
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

	mqtt, err := transport.RunMqtt()
	if err != nil {
		log.Panic("error running mqtt server", err)
	}

	engine := server.NewEngine()
	engine.Process(mqtt)

	finish := make(chan os.Signal, 1)
	signal.Notify(finish, syscall.SIGINT, syscall.SIGTERM)
	<-finish
}
