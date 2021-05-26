package main

import (
	"github.com/MajaSuite/mqtt/packet"
	"github.com/MajaSuite/mqtt/transport"
	"log"
	"time"
)

func main() {
	client, err := transport.Connect("127.0.0.1:1883", "go-mqttclient", 30,"", "")
	if err != nil {
		panic(err)
	}

	s := packet.NewSubscribe()
	s.Topics = []packet.SubscribePayload{
		packet.SubscribePayload{Topic: "home/topic", QoS: 0},
		packet.SubscribePayload{Topic: "home/another", QoS: 1},
	}
	client.Subscribe(s)

	u := packet.NewUnSub()
	u.Topics = []packet.SubscribePayload{
		packet.SubscribePayload{Topic: "home/another", QoS: 0},
	}
	client.Unsubscribe(u)

	p := packet.NewPublish()
	p.Topic = "home/topic"
	p.Payload = "{\"message\":\"test\"}"
	p.QoS = 1
	client.Sendout <- p

	for pkt := range client.Broker {
		if pkt.Type() == packet.PUBLISH {
			log.Println("NEW PUBLISH: ", pkt)
		}
	}

	client.Disconnect()
	time.Sleep(time.Second*2)
}
