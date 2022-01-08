package broker

import (
	"github.com/MajaSuite/mqtt/db"
	"github.com/MajaSuite/mqtt/packet"
	"github.com/MajaSuite/mqtt/transport"
	"log"
)

func (e *Engine) clientThread() {
	for event := range e.channel {
		log.Println("engine receive message: " + event.String())

		switch event.PacketType {
		case packet.DISCONNECT:
			client := e.clients[event.ClientId]
			client.Stop()

			// todo send will message

			delete(e.clients, event.ClientId)
		case packet.SUBSCRIBE:
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
		case packet.PUBLISH: // in
			client := e.clients[event.ClientId]

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
			case packet.AtLeastOnce: // PUBLISH -> PUBACK
				res := transport.Event{PacketType: packet.PUBACK, MessageId: event.MessageId, ClientId: event.ClientId}
				client.clientChan <- res

				// save message
				e.delivery[event.MessageId] = res

				// send published message to all subscribed clients
				e.publishMessage(event)
			case packet.ExactlyOnce: // PUBLISH ->PUBREC, PUBREL - PUBCOMP
				client.clientChan <- transport.Event{PacketType: packet.PUBREC, MessageId: event.MessageId,
					ClientId: event.ClientId}

				// record packet in delivery queue
				e.delivery[event.MessageId] = event
			}
		case packet.PUBACK: // out
			// we receive answer on publish command,
			_, ok := e.send[event.MessageId]
			if ok {
				// remove from sent queue
				delete(e.send, event.MessageId)
			} else {
				log.Printf("error process PUBACK, message %d for %s was not found\n", event.MessageId, event.ClientId)
			}
		case packet.PUBREC: // out
			// we receive answer on PUBLISH with qos2,
			answer, ok := e.send[event.MessageId]
			if ok && event.ClientId == answer.ClientId {
				// send PUBREL
				e.clients[event.ClientId].clientChan <- transport.Event{PacketType: packet.PUBREL,
					ClientId: event.ClientId, MessageId: event.MessageId}
			} else {
				log.Printf("error process PUBREC, message %d for %s was not found\n", event.MessageId, event.ClientId)
			}
		case packet.PUBREL: // in
			event, ok := e.delivery[event.MessageId]
			if ok {
				// send answer PUBCOMP
				e.clients[event.ClientId].clientChan <- transport.Event{PacketType: packet.PUBCOMP,
					MessageId: event.MessageId, ClientId: event.ClientId}

				// send publish
				e.publishMessage(event)

				// remove from delivery queue
				delete(e.delivery, event.MessageId)
			} else {
				log.Printf("error process PUBREL, message %d for %s was not found\n", event.MessageId, event.ClientId)
			}
		case packet.PUBCOMP: // out
			// we receive answer on PUBREL
			_, ok := e.send[event.MessageId]
			if ok {
				// remove from sent queue with qos2
				delete(e.send, event.MessageId)
			} else {
				log.Printf("error process PUBCOMP, message %d for %s was not found\n", event.MessageId, event.ClientId)
			}
		default:
			log.Println("engine: unexpected disconnect")

			client := e.clients[event.ClientId]
			if client != nil {
				// if will message is set for client - send this message to all clients
				if client.will != nil {
					will := transport.Event{
						Topic:   transport.EventTopic{Name: client.will.Topic},
						Payload: client.will.Payload,
						Qos:     client.will.QoS.Int(),
						Retain:  client.will.Retain,
					}
					for _, c := range e.clients {
						if c != nil {
							client.clientChan <- will
						}
					}
				}

				// delete clean session
				if !client.session {
					delete(e.clients, event.ClientId)
				}
			}
		}
	}
}
