package broker

import (
	"fmt"
	"log"
	"time"

	"github.com/MajaSuite/mqtt/db"
	"github.com/MajaSuite/mqtt/packet"
)

type Broker struct {
	debug   bool
	channel chan packet.Packet // channel to mqtt broker engine (to push packet, received from client)
	clients map[string]*Client // hashmap of all connected clients
}

func NewBroker(debug bool) *Broker {
	broker := &Broker{
		debug:   debug,
		channel: make(chan packet.Packet),
		clients: make(map[string]*Client),
	}

	go broker.broker()
	go broker.rescan()

	return broker
}

// send to all subscribed clients
func (e *Broker) publishMessage(pkt *packet.PublishPacket) {
	for _, client := range e.clients {
		if client != nil && client.conn != nil {
			if client.Contains(pkt.Topic) {
				client.messageId++
				pkt.Id = client.messageId
				client.channel <- pkt

				if pkt.QoS > 0 {
					client.ack[fmt.Sprintf("s%d", pkt.Id)] = pkt
				}
			}
		}
	}
}

// send will message (on client disconnect)
func (e *Broker) sendWill(client *Client) {
	if client != nil {
		if client.will != nil {
			publish := packet.NewPublish()
			//client.will.Flag
			publish.Topic = client.will.Topic
			publish.Payload = client.will.Payload
			publish.QoS = client.will.QoS
			//client.will.Retain

			for _, c := range e.clients {
				if c != nil && c.conn != nil {
					client.channel <- publish
				}
			}
		}
	}
}

func (e *Broker) rescan() {
	for {
		time.Sleep(time.Second * 10)

		if e.debug {
			//log.Println("rescan queue for ack")
		}

		//for id, event := range e.sent {
		//	// TODO try to send messages without ack again
		//}
	}
}

func (e *Broker) broker() {
	for pkt := range e.channel {
		if e.debug {
			log.Printf("broker receive message %s from %s", pkt, pkt.Source())
		}
		client := e.clients[pkt.Source()]

		switch pkt.Type() {
		case packet.DISCONNECT:
			e.sendWill(client)
			if !client.session {
				delete(e.clients, pkt.Source())
			}
		case packet.SUBSCRIBE:
			retains, _ := db.FetchRetain()
			for _, payload := range pkt.(*packet.SubscribePacket).Topics {
				client.addSubscription(payload)
				if retains[payload.Topic] != nil {
					// have retain message for this client and current topic. push it to client
					client.messageId++
					publish := packet.NewPublish()
					publish.Id = client.messageId
					publish.Retain = true
					publish.Topic = payload.Topic
					publish.QoS = retains[payload.Topic].QoS
					publish.Payload = retains[payload.Topic].Payload
					client.channel <- publish
				}

				// if not clean session - save subscription
				if client.session {
					db.SaveSubscription(pkt.Source(), payload.Topic, payload.QoS.Int())
				}
			}
		case packet.UNSUBSCRIBE:
			if e.clients[pkt.Source()].session {
				for _, payload := range pkt.(*packet.UnSubscribePacket).Topics {
					db.DeleteSubscription(pkt.Source(), payload.Topic)
				}
			}
		case packet.PUBLISH:
			if pkt.(*packet.PublishPacket).Retain {
				if pkt.(*packet.PublishPacket).Payload != "" {
					db.SaveRetain(pkt.(*packet.PublishPacket).Topic, pkt.(*packet.PublishPacket).Payload, pkt.(*packet.PublishPacket).QoS.Int())
				} else {
					db.DeleteRetain(pkt.(*packet.PublishPacket).Topic, pkt.(*packet.PublishPacket).QoS.Int())
				}
			}

			if pkt.(*packet.PublishPacket).DUP {
				dup := client.ack[fmt.Sprintf("r%d", pkt.(*packet.PublishPacket).Id)]
				if dup != nil {
					//if dup.QoS() == ...
					// ...
				}
			} else {
				switch pkt.(*packet.PublishPacket).QoS {
				case packet.AtMostOnce:
					e.publishMessage(pkt.(*packet.PublishPacket))
				case packet.AtLeastOnce:
					puback := packet.NewPubAck()
					puback.Id = pkt.(*packet.PublishPacket).Id
					client.channel <- puback
					e.publishMessage(pkt.(*packet.PublishPacket))
				case packet.ExactlyOnce:
					pubrec := packet.NewPubRec()
					pubrec.Id = pkt.(*packet.PublishPacket).Id
					client.channel <- pubrec
					client.ack[fmt.Sprintf("r%d", pubrec.Id)] = pkt
				}
			}
		case packet.PUBACK:
			// we receive answer on client publish command with QOS(1); prefix to rescan is "s"
			p := client.ack[fmt.Sprintf("s%d", pkt.(*packet.PubAckPacket).Id)]
			if p != nil {
				log.Printf("%s confirmed", p)
				delete(client.ack, fmt.Sprintf("s%d", pkt.(*packet.PubAckPacket).Id))
			} else {
				log.Printf("packet to ack %d from %s not found", pkt.(*packet.PubAckPacket).Id, pkt.Source())
			}
		case packet.PUBREC: // we
			// we receive answer on our PUBLISH with qos2,
			p := client.ack[fmt.Sprintf("s%d", pkt.(*packet.PubRecPacket).Id)]
			if p != nil {
				delete(client.ack, fmt.Sprintf("s%d", pkt.(*packet.PubRecPacket).Id))
				log.Printf("%s confirmed", p)

				pubrel := packet.NewPubRel()
				pubrel.Id = pkt.(*packet.PubRecPacket).Id
				client.channel <- pubrel

				client.ack[fmt.Sprintf("l%d", pkt.(*packet.PubRecPacket).Id)] = p
			} else {
				log.Printf("packet to rec %d from %s not found", pkt.(*packet.PubRecPacket).Id, pkt.Source())
			}
		case packet.PUBREL:
			// we receive answer on client send publish and client answer to pubrec
			p := client.ack[fmt.Sprintf("r%d", pkt.(*packet.PubRelPacket).Id)]
			if p != nil {
				delete(client.ack, fmt.Sprintf("r%d", pkt.(*packet.PubRelPacket).Id))
				log.Printf("%s confirmed", p)

				e.publishMessage(p.(*packet.PublishPacket))

				pubcomp := packet.NewPubComp()
				pubcomp.Id = pkt.(*packet.PubRelPacket).Id
				client.channel <- pubcomp
			} else {
				log.Printf("packet to rel %d from %s not found", pkt.(*packet.PubRelPacket).Id, pkt.Source())

				// keep client silents
				pubcomp := packet.NewPubComp()
				pubcomp.Id = pkt.(*packet.PubRelPacket).Id
				client.channel <- pubcomp
			}
		case packet.PUBCOMP: // we
			// we receive answer on our PUBREL
			p := client.ack[fmt.Sprintf("l%d", pkt.(*packet.PubCompPacket).Id)]
			if p != nil {
				delete(client.ack, fmt.Sprintf("l%d", pkt.(*packet.PubCompPacket).Id))

				for _, oc := range e.clients {
					if oc != nil && oc.conn != nil && oc.clientId != client.clientId {
						if oc.Contains(p.(*packet.PublishPacket).Topic) {
							oc.messageId++
							p.(*packet.PublishPacket).Id = oc.messageId
							oc.channel <- p
							oc.ack[fmt.Sprintf("s%d", oc.messageId)] = p
						}
					}
				}
			} else {
				log.Printf("packet to comp %d from %s not found", pkt.(*packet.PubCompPacket).Id, pkt.Source())
			}
		default:
			log.Println("engine unexpected disconnect")
			if client != nil {
				e.sendWill(client)
				if !client.session {
					delete(e.clients, pkt.Source())
				}
			}
		}
	}
}
