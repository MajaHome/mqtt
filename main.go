package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/MajaSuite/mqtt/broker"
	"github.com/MajaSuite/mqtt/db"
)

var (
	cert  = flag.String("cert", "broker.crt", "path to broker certificate")
	key   = flag.String("key", "broker.key", "path to broker private key")
	debug = flag.Bool("debug", false, "print debuging hex dumps")
)

func main() {
	flag.Parse()
	log.Println("starting broker")

	log.Println("initialize database")
	if err := db.Open("mqtt.db"); err != nil {
		panic(err)
	}

	listener := broker.NewListener(*debug)
	if listener == nil {
		log.Panic("error start listener")
	}

	broker.NewMqtt(*debug, listener)

	finish := make(chan os.Signal, 1)
	signal.Notify(finish, syscall.SIGINT, syscall.SIGTERM)
	<-finish
	log.Println("finished")
}
