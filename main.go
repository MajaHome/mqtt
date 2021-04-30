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
	debug = flag.Bool("debug", false, "show debug info")
	cert  = flag.String("cert", "server.crt", "path to server certificate")
	key   = flag.String("key", "server.key", "path to server private key")
)

func main() {
	flag.Parse()
	log.Println("Starting broker ...")

	// init
	backend := server.GetBackend(*debug)
	engine := server.GetEngine(backend)

	mqtt, err := server.Run()
	if err != nil {
		log.Panic("error running server", err)
	}
	go engine.Manage(mqtt)

	finish := make(chan os.Signal, 1)
	signal.Notify(finish, syscall.SIGINT, syscall.SIGTERM)
	<-finish

	engine.Close()
}
