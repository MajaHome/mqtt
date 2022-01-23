package broker

import (
	"github.com/MajaSuite/mqtt/db"
	"github.com/MajaSuite/mqtt/packet"
	"github.com/MajaSuite/mqtt/transport"
	"log"
)

func (e *Mqtt) broker() {
	for event := range e.brokerChannel {
		if e.debug {
			log.Println("engine receive message:", event)
		}

		switch event.PacketType {
		case packet.DISCONNECT:
			client := e.clients[event.ClientId]
			client.Stop()

			if client.will != nil {
				if e.debug {
					log.Println("send will message", client.will)
				}

				e.sendWill(e.clients[event.ClientId])
			}

			delete(e.clients, event.ClientId)
		case packet.SUBSCRIBE:
			// fetch messages for subscribed topic
			if retains, err := db.FetchRetain(); err == nil {
				for topic, retain := range retains {
					if packet.MatchTopic(event.Topic.Name, topic) {
						retain.ClientId = event.ClientId
						e.publishMessage(retain)
					}
				}
			}

			// if not clean session - save subscription
			if e.clients[event.ClientId].session && !event.Restore {
				db.SaveSubscription(event.ClientId, event.Topic.Name, event.Topic.Qos)
			}
		case packet.UNSUBSCRIBE:
			if e.clients[event.ClientId].session {
				db.DeleteSubscription(event.ClientId, event.Topic.Name)
			}
		case packet.PUBLISH:
			client := e.clients[event.ClientId] // who send message

			if event.Retain {
				if event.Payload != "" {
					db.SaveRetain(event.Topic.Name, event.Payload, event.Qos)
				} else {
					db.DeleteRetain(event.Topic.Name, event.Qos)
				}
			}

			switch packet.QoS(event.Qos) {
			case packet.AtMostOnce:
				e.publishMessage(event)
			case packet.AtLeastOnce:
				client.toSendOut <- transport.Event{PacketType: packet.PUBACK, MessageId: event.MessageId,
					ClientId: event.ClientId}

				// publish to all subscribed clients
				e.publishMessage(event)
			case packet.ExactlyOnce:
				client.toSendOut <- transport.Event{PacketType: packet.PUBREC, MessageId: event.MessageId,
					ClientId: event.ClientId}

				// record packet in delivery queue
				e.sentWithQos[event.MessageId] = event
			}
		case packet.PUBACK:
			// we receive answer on our publish command with QOS(1)
			_, ok := e.sent[event.MessageId]
			if ok {
				// remove from sent queue
				delete(e.sent, event.MessageId)
			} else {
				log.Printf("error process PUBACK, message %d for %s was not found", event.MessageId, event.ClientId)
			}
		case packet.PUBREC:
			// we receive answer on our PUBLISH with qos2,
			answer, ok := e.sent[event.MessageId] // todo different client may have same message-id
			if ok && event.ClientId == answer.ClientId {
				// send PUBREL
				e.clients[event.ClientId].toSendOut <- transport.Event{PacketType: packet.PUBREL,
					ClientId: event.ClientId, MessageId: event.MessageId}
			} else {
				// todo
				log.Printf("error process PUBREC, message %d for %s was not found\n", event.MessageId, event.ClientId)
			}
		case packet.PUBREL:
			event, ok := e.sentWithQos[event.MessageId]
			if ok {
				// send answer PUBCOMP
				e.clients[event.ClientId].toSendOut <- transport.Event{PacketType: packet.PUBCOMP,
					MessageId: event.MessageId, ClientId: event.ClientId}

				// send publish
				e.publishMessage(event)

				// remove from delivery queue
				delete(e.sentWithQos, event.MessageId)
			} else {
				log.Printf("error process PUBREL, message %d for %s was not found\n", event.MessageId, event.ClientId)
			}
		case packet.PUBCOMP:
			// we receive answer on PUBREL
			_, ok := e.sent[event.MessageId]
			if ok {
				delete(e.sent, event.MessageId)
			} else {
				log.Printf("error process PUBCOMP, message %d for %s was not found\n", event.MessageId, event.ClientId)
			}
		default:
			log.Println("engine unexpected disconnect")
			e.sendWill(e.clients[event.ClientId])
		}
	}
}
