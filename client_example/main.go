package main

import (
	"flag"
	"log"
	"strconv"
	"time"

	"github.com/MajaSuite/mqtt/packet"
	"github.com/MajaSuite/mqtt/transport"
)

var (
	debug = flag.Bool("debug", false, "print debuging hex dumps")
)

func main() {
	client := transport.Connect("127.0.0.1:1883", "go-mqttclient", 30, "", "", *debug)
	if client == nil {
		panic("can't connect to mqtt server")
	}

	go func() {
		for {
			for pkt := range client.Receive {
				log.Println("NEW PACKET FROM SERVER: ", pkt)
			}
		}
	}()

	s := packet.NewSubscribe()
	s.Id = 123
	s.Topics = []packet.SubscribePayload{
		{Topic: "home/topic", QoS: 0},
		{Topic: "home/another", QoS: 1},
	}
	log.Println("send", s)
	client.Send <- s

	log.Println("sleep")
	time.Sleep(time.Second * 2)

	u := packet.NewUnSub()
	u.Id = 321
	u.Topics = []packet.SubscribePayload{
		{Topic: "home/another", QoS: 0},
	}
	log.Println("send", u)
	client.Send <- u

	p := packet.NewPublish()
	p.Id = uint16(783)
	p.Topic = "home/topic"
	p.QoS = 1
	p.Payload = "{\"name\":\"uiwyfencbo47ryo34cnoeirwcfuoegiruoiertwupoiqwucbveprugpt485ugboewugboeirueow\",\"model\":\"yeelink.light.mono5\", \"token\":\"cfcb32279b1a033a72aa69601ff15f01\",\"mac\":\"5C:E5:0C:CC:6B:27\"}, qos: 1, retain: false, dup:false}"
	log.Println("send", p)
	client.Send <- p

	for i := 0; i < 5; i++ {
		p := packet.NewPublish()
		p.Id = uint16(i)
		p.Topic = "home/topic"
		p.Payload = "{\"message\":\"" + strconv.Itoa(i) + "\"}"
		p.QoS = 2
		log.Println("send", p)
		client.Send <- p
	}

	log.Println("sleep 5")
	time.Sleep(time.Second * 5)

	log.Println("disconnect", packet.NewDisconnect())
	client.Send <- packet.NewDisconnect()
	time.Sleep(time.Second * 2)
}
