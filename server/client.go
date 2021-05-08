package server

import (
	"io"
	"log"
	"mqtt/packet"
	"net"
	"time"
)

type Client struct {
	conn     net.Conn
	engine   chan Event
	incoming chan string

	id       string
	Topics   []string // subscribed topics

	willQoS    packet.QoS
	willRerail bool
}

func NewClient(id string, conn net.Conn, engine chan Event) *Client {
	inChannel := make(chan string)

	return &Client{
		conn:     conn,
		engine:   engine,
		incoming: inChannel,
		id:       id,
	}
}

func (c *Client) Start(server *Server) {
	for {
		c.conn.SetReadDeadline(time.Now().Add(time.Second * 1))
		pkt, _ := server.ReadPacket(c.conn)
		if pkt == nil {
			continue
		}

		var err error
		switch pkt.Type() {
		case packet.DISCONNECT:
			res := packet.NewDisconnect()

			event := &Event{
				clientId: c.id,
				packetType: pkt.Type(),
			}
			c.engine <- *event

			err = server.WritePacket(c.conn, res)
		case packet.PING:
			res := packet.NewPong()

			err = server.WritePacket(c.conn, res)
		case packet.SUBSCRIBE:
			res := packet.NewSubAck()

			r := pkt.(*packet.SubscribePacket)
			res.Id = r.Id
			var qos []packet.QoS
			var topics []EventTopic
			for _, q := range r.Topics {
				topics = append(topics, EventTopic { qos: q.QoS.Int(),	topic: q.Topic })
				qos = append(qos, q.QoS)
			}
			res.ReturnCodes = qos

			event := Event{
				clientId: c.id,
				packetType: pkt.Type(),
				messageId: r.Id,
				topics: topics,
			}
			c.engine <- event

			err = server.WritePacket(c.conn, res)
		case packet.UNSUBSCRIBE:
			res := packet.NewUnSubAck()

			//r := pkt.(*packet.UnSubscriibePacket)

			event := &Event{
				clientId: c.id,
				packetType: pkt.Type(),
			}
			c.engine <- *event

			err = server.WritePacket(c.conn, res)
		case packet.PUBLISH:
			res := packet.NewPubAck()
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
				clientId: c.id,
				packetType: pkt.Type(),
			}
			c.engine <- *event

			err = server.WritePacket(c.conn, res)
		case packet.PUBCOMP:
			res := packet.NewPubAck()

			event := &Event{
				clientId: c.id,
				packetType: pkt.Type(),
			}
			c.engine <- *event

			err = server.WritePacket(c.conn, res)
		case packet.PUBREC:
			res := packet.NewPubAck()

			event := &Event{
				clientId: c.id,
				packetType: pkt.Type(),
			}
			c.engine <- *event

			err = server.WritePacket(c.conn, res)
		case packet.PUBREL:
			res := packet.NewPubAck()

			event := &Event{
				clientId: c.id,
				packetType: pkt.Type(),
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
