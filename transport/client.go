package transport

import (
	"github.com/MajaSuite/mqtt/packet"
	"log"
	"net"
)

type MqttClient struct {
	conn	net.Conn
	broker	chan *Event		// messages from broker
	sendout chan *Event		// messages to broker
}

func Connect(addr string, clientId string, login string, pass string)  (*MqttClient, error) {
	// socket connect
	c, err := net.Dial("tcp4", addr)
	if err != nil {
		return nil, err
	}

	// send connect
	cp := packet.NewConnect()
	cp.Version = 4
	cp.ClientID = clientId
	cp.KeepAlive = 60
	if login != "" {
		cp.Username = login
	}
	if pass != "" {
		cp.Password = pass
	}
	if err := WritePacket(c, cp); err != nil {
		return nil, err
	}

	resp, err := ReadPacket(c)
	if err != nil {
		return nil, err
	}

	if resp == nil || resp.Type() != packet.CONNACK {
		return nil, packet.ErrProtocolError
	}

	if resp.(*packet.ConnAckPacket).ReturnCode == 0 {
		mqttClient := &MqttClient{conn: c, broker: make(chan *Event)}
		go mqttClient.manage()
		return mqttClient, nil
	}

	return nil, packet.ErrConnect
}

func (c *MqttClient) Disconnect() error {
	// send to c.sendout
	return nil
}

func (c *MqttClient) Subscribe() {
	// save subscription?
	// send to c.sendout
}

func (c *MqttClient) Unsubscribe() {
	// send to c.sendout
}

func (c *MqttClient) Publish() {
	// send to c.sendout
}

func (c *MqttClient) manage() {
	for event := range c.sendout {
		log.Println("to send: ", event)
	}
}