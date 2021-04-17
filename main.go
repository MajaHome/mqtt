package main

import (
	"flag"
	"fmt"
	"os"
	"syscall"
	"os/signal"
	"mqtt/server"
)

var (
	cert = flag.String("cert", "server.crt", "path to server certificate")
	key = flag.String("key", "server.key", "path to server private key")
)

func main() {
	flag.Parse()
	fmt.Println("Starting broker ...")

	// init
	backend := server.GetBackend()
	engine := server.GetEngine(backend)

	mqtt, err := server.Run()
	if err != nil {
		panic(err)
	}
	go engine.Manage(mqtt)

	//var mqtts server.Server
	//if cert != nil && key != nil {
	//	c, err := tls.LoadX509KeyPair(*cert, *key)
	//	if err != nil {
	//		panic(err)
	//	}
	//
	//	tls := &tls.Config{Certificates: []tls.Certificate{c}}
	//	mqtts, err = server.RunSecured(tls)
	//	if err != nil {
	//		panic(err)
	//	}
	//
	//	go engine.Accept(mqtts)
	//}

	finish := make(chan os.Signal, 1)
	signal.Notify(finish, syscall.SIGINT, syscall.SIGTERM)
	<-finish

	// shutdown
	//engine.Close()
}