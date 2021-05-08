package server

import (
	"io"
	"log"
	"mqtt/packet"
	"net"
	"time"
)

type Client struct {
	conn    net.Conn
	engine  chan Event
	channel chan string

	id     string
	Topics []string // subscribed topics

	willQoS    packet.QoS
	willRerail bool
}

func NewClient(id string, conn net.Conn, engine chan Event) *Client {
	channel := make(chan string)

	return &Client{
		conn:    conn,
		engine:  engine,
		channel: channel,
		id:      id,
	}
}

func (c *Client) Start(server *Server) {
	for {
		c.conn.SetReadDeadline(time.Now().Add(time.Millisecond * 10000))
		pkt, _ := server.ReadPacket(c.conn)

		select {
		case msg := <-c.channel:
			log.Println("client: " + msg)
		default:
			break
		}

		if pkt == nil {
			continue
		}

		var err error
		switch pkt.Type() {
		case packet.DISCONNECT:
			res := packet.NewDisconnect()

			event := &Event{
				clientId:   c.id,
				packetType: pkt.Type(),
			}
			c.engine <- *event

			err = server.WritePacket(c.conn, res)
		case packet.PING:
			res := packet.NewPong()

			err = server.WritePacket(c.conn, res)
		case packet.SUBSCRIBE:
			req := pkt.(*packet.SubscribePacket)
			res := packet.NewSubAck()

			res.Id = req.Id
			var qos []packet.QoS
			for _, topic := range req.Topics {
				event := Event{
					clientId:   c.id,
					packetType: pkt.Type(),
					messageId:  req.Id,
					topic:      topic.Topic,
					qos:        topic.QoS.Int(),
				}
				c.engine <- event
				qos = append(qos, topic.QoS)
			}
			res.ReturnCodes = qos

			err = server.WritePacket(c.conn, res)
		case packet.UNSUBSCRIBE:
			//req := pkt.(*packet.UnSubscriibePacket)
			res := packet.NewUnSubAck()

			event := &Event{
				clientId:   c.id,
				packetType: pkt.Type(),
			}
			c.engine <- *event

			err = server.WritePacket(c.conn, res)
		case packet.PUBLISH:
			req := pkt.(*packet.PublishPacket)

			/*
				RETAIN – при публикации данных с установленным флагом retain, брокер сохранит его. При
				следующей подписке на этот топик брокер незамедлительно отправит сообщение с этим флагом.

				DUP – флаг дубликата устанавливается, когда клиент или MQTT брокер совершает повторную
				отправку пакета (используется в типах PUBLISH, SUBSCRIBE, UNSUBSCRIBE, PUBREL). При
				установленном флаге переменный заголовок должен содержать Message ID (идентификатор
				сообщения)

				QoS 0 At most once. На этом уровне издатель один раз отправляет сообщение брокеру и не ждет
				подтверждения от него, то есть отправил и забыл.

				QoS 1 At least once. Этот уровень гарантирует, что сообщение точно будет доставлено брокеру,
				но есть вероятность дублирования сообщений от издателя. После получения дубликата сообщения,
				брокер снова рассылает это сообщение подписчикам, а издателю снова отправляет подтверждение
				о получении сообщения. Если издатель не получил PUBACK сообщения от брокера, он повторно
				отправляет этот пакет, при этом в DUP устанавливается «1».
				PUBLISH -> PUBACK

				QoS 2 Exactly once. На этом уровне гарантируется доставка сообщений подписчику и исключается
				возможное дублирование отправленных сообщений.
				PUBLISH ->PUBREC, PUBREL - PUBCOMP

				Издатель отправляет сообщение брокеру. В этом сообщении указывается уникальный Packet ID,
				QoS=2 и DUP=0. Издатель хранит сообщение неподтвержденным пока не получит от брокера ответ
				PUBREC. Брокер отвечает сообщением PUBREC в котором содержится тот же Packet ID. После его
				получения издатель отправляет PUBREL с тем же Packet ID. До того, как брокер получит PUBREL
				он должен хранить копию сообщения у себя. После получения PUBREL он удаляет копию сообщения
				и отправляет издателю сообщение PUBCOMP о том, что транзакция завершена.
			*/

			// if payload is empty - unsubscribe to the topic

			event := &Event{
				clientId:   c.id,
				packetType: pkt.Type(),
				messageId:  req.Id,
				topic:      req.Topic,
				payload:    req.Payload,
				qos:        req.QoS.Int(),
				retain:     req.Retain,
				dublicate:  req.DUP,
			}
			c.engine <- *event

			switch req.QoS {
			case packet.AtLeastOnce:
				res := packet.NewPubAck()
				res.Id = req.Id
				err = server.WritePacket(c.conn, res)
			case packet.ExactlyOnce:
				res := packet.NewPubRec()
				res.Id = req.Id
				// read response from engine - if packet with id is registered
				err = server.WritePacket(c.conn, res)
			}
		case packet.PUBREL:
			req := pkt.(*packet.PubRelPacket)
			res := packet.NewPubComp()
			res.Id = req.Id
			// read response from engine - if packet with id is registered

			event := &Event{
				clientId:   c.id,
				packetType: pkt.Type(),
				messageId:  req.Id,
			}
			c.engine <- *event

			err = server.WritePacket(c.conn, res)
		default:
			err = packet.ErrUnknownPacket
		}

		if err != nil {
			if err == io.EOF {
				// normal disconnect
				/*
					Will Flag - при установленном флаге, после того, как клиент отключится от брокера без отправки команды DISCONNECT
					(в случаях непредсказуемого обрыва связи и т.д.), брокер оповестит об этом всех подключенных к нему клиентов через
					так называемый Will Message.
				*/
			} else {
				log.Println("err serve: ", err.Error())
			}

			log.Println("client disconnect")
			c.conn.Close()
			break
		}
	}
}

func (c *Client) saveWill(qos packet.QoS, retain bool) {
	c.willQoS = qos
	c.willRerail = retain
}
