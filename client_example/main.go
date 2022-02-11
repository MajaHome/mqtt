package main

import (
	"flag"
	"github.com/MajaSuite/mqtt/client"
	"github.com/MajaSuite/mqtt/packet"
	"log"
	"strconv"
	"time"
)

var (
	debug = flag.Bool("debug", false, "print debuging hex dumps")
)

func main() {
	flag.Parse()
	cc, err := client.Connect("127.0.0.1:1883", "go-mqttclient", 30, false, "", "", *debug)
	if err != nil {
		panic("can't connect to mqtt server")
	}

	go func() {
		for pkt := range cc.Receive {
			log.Println("NEW PACKET FROM SERVER: ", pkt)
		}
	}()

	s := packet.NewSubscribe()
	s.Id = 123
	s.Topics = []packet.SubscribePayload{
		{Topic: "home/topic", QoS: 0},
		{Topic: "home/another", QoS: 2},
	}
	log.Println("send", s)
	cc.Send <- s

	time.Sleep(time.Second * 2)

	p := packet.NewPublish()
	p.Id = uint16(1)
	p.Topic = "home/another"
	p.QoS = 0
	p.Payload = "{\"name\":\"uiwyfencbo47ryo34cnoeirwcfuoegiruoiertwupoiqwucbveprugpt485ugboewugbocfcb32279b1a033a72aa69601ff15f027\"}"
	log.Println("send", p)
	cc.Send <- p

	u := packet.NewUnSub()
	u.Id = 321
	u.Topics = []packet.SubscribePayload{
		{Topic: "home/another", QoS: 0},
	}
	log.Println("send", u)
	cc.Send <- u

	for i := 2; i < 50; i++ {
		pp := packet.NewPublish()
		pp.Id = uint16(i)
		pp.Topic = "home/topic"
		pp.Payload = "{\"message\":\"" + strconv.Itoa(i) + "\"}"
		pp.QoS = 1
		log.Println("send", pp)
		cc.Send <- pp
		time.Sleep(time.Second * 2)
	}

	log.Println("disconnect", packet.NewDisconnect())
	cc.Send <- packet.NewDisconnect()
	time.Sleep(time.Second)
}
