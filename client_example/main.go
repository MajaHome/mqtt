package main

import (
	"github.com/MajaSuite/mqtt/transport"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	_, err := transport.Connect("127.0.0.1:1883", "mqttclient", "", "")
	if err != nil {
		panic(err)
	}


	finish := make(chan os.Signal, 1)
	signal.Notify(finish, syscall.SIGINT, syscall.SIGTERM)
	<-finish
}

