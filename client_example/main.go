package main

import (
	"log"
	"strconv"
	"time"

	"github.com/MajaSuite/mqtt/packet"
	"github.com/MajaSuite/mqtt/transport"
)

func main() {
	client, err := transport.Connect("127.0.0.1:1883", "go-mqttclient", 30, "", "", false)
	if err != nil {
		panic(err)
	}

	go func() {
		for {
			for pkt := range client.Broker {
				log.Println("NEW PACKET FROM SERVER: ", pkt)
			}
		}
	}()

	go client.Start()

	s := packet.NewSubscribe()
	s.Id = 123
	s.Topics = []packet.SubscribePayload{
		{Topic: "home/topic", QoS: 0},
		{Topic: "home/another", QoS: 1},
	}
	client.Sendout <- s

	time.Sleep(time.Second * 2)

	u := packet.NewUnSub()
	u.Id = 321
	u.Topics = []packet.SubscribePayload{
		{Topic: "home/another", QoS: 0},
	}
	client.Sendout <- u

	p := packet.NewPublish()
	p.Id = uint16(783)
	p.Topic = "home/topic"
	p.QoS = 1
	p.Payload = "{\"name\":\"uiwyfencbo47ryo34cnoeirwcfuoegiruoiertwupoiqwucbveprugpt485ugboewugboeirueow\",\"model\":\"yeelink.light.mono5\", \"token\":\"cfcb32279b1a033a72aa69601ff15f01\",\"mac\":\"5C:E5:0C:CC:6B:27\"}, qos: 1, retain: false, dup:false}"
	client.Sendout <- p

	for i := 0; i < 5; i++ {
		p := packet.NewPublish()
		p.Id = uint16(i)
		p.Topic = "home/topic"
		p.Payload = "{\"message\":\"" + strconv.Itoa(i) + "\"}"
		p.QoS = 2
		client.Sendout <- p
	}
	time.Sleep(time.Second * 5)

	client.Sendout <- packet.NewDisconnect()
	time.Sleep(time.Second * 2)
}
