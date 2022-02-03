package packet

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"io"
	"log"
	"net"
	"strings"
)

const (
	RESERVED Type = iota
	CONNECT
	CONNACK
	PUBLISH
	PUBACK
	PUBREC
	PUBREL
	PUBCOMP
	SUBSCRIBE
	SUBACK
	UNSUBSCRIBE
	UNSUBACK
	PING
	PONG
	DISCONNECT
)

var (
	ErrInvalidPacketType   = errors.New("invalid packet type")
	ErrProtocolError       = errors.New("protocol error (not supported)")
	ErrInvalidPacketLength = errors.New("invalid packet Len")
	ErrUnknownPacket       = errors.New("unknown packet type")
	ErrUnsupportedVersion  = errors.New("unsupported mqtt version")
	ErrConnect             = errors.New("error connect to broker")
)

type Packet interface {
	Source() string
	SetSource(clientId string)
	Type() Type
	Length() int
	Unpack(buf []byte) error
	Pack() []byte
	String() string
}

type Type byte

func (t Type) String() string {
	switch t {
	case CONNECT:
		return "Connect"
	case CONNACK:
		return "ConnAck"
	case PUBLISH:
		return "Publish"
	case PUBACK:
		return "PubAck"
	case PUBREC:
		return "PubRec"
	case PUBREL:
		return "PubRel"
	case PUBCOMP:
		return "PubComp"
	case SUBSCRIBE:
		return "Subscribe"
	case SUBACK:
		return "SubAck"
	case UNSUBSCRIBE:
		return "Unsubscribe"
	case UNSUBACK:
		return "UnsubAck"
	case PING:
		return "Ping"
	case PONG:
		return "Pong"
	case DISCONNECT:
		return "Disconnect"
	}

	return "Unknown"
}

func Create(buf byte) Packet {
	t := Type(buf >> 4)

	switch t {
	case CONNECT:
		return CreateConnect(buf)
	case CONNACK:
		return CreateConnAck(buf)
	case PUBLISH:
		return CreatePublish(buf)
	case PUBACK:
		return CreatePubAck(buf)
	case PUBREC:
		return CreatePubRec(buf)
	case PUBREL:
		return CreatePubRel(buf)
	case PUBCOMP:
		return CreatePubComp(buf)
	case SUBSCRIBE:
		return CreateSubscribe(buf)
	case SUBACK:
		return CreateSubAck(buf)
	case UNSUBSCRIBE:
		return CreateUnSubscribe(buf)
	case UNSUBACK:
		return CreateUnSubAck(buf)
	case PING:
		return CreatePing(buf)
	case PONG:
		return CreatePong(buf)
	case DISCONNECT:
		return CreateDisconnect(buf)
	}

	return nil
}

func ReadPacket(conn net.Conn, debug bool) (Packet, error) {
	header := make([]byte, 2)
	if n, err := conn.Read(header); n < 2 || err != nil {
		return nil, io.ErrUnexpectedEOF
	}

	// read variable packet length
	packetLength := int(header[1])
	if packetLength > 127 {
		add := make([]byte, 1)

		for i := 2; ; i++ {
			// allowed only 4 bytes for len
			if i > 4 {
				return nil, ErrInvalidPacketLength
			}

			if n, err := conn.Read(add); n < 1 || err != nil {
				return nil, io.ErrUnexpectedEOF
			}
			header = append(header, add[0])
			if add[0] <= 127 {
				break
			}
		}

		// convert
		if x, n := binary.Uvarint(header[1:]); n != len(header[1:]) {
			return nil, ErrInvalidPacketLength
		} else {
			packetLength = int(x)
		}
	}

	if debug {
		log.Printf("read: header: 0x%x, %d bytes\n", header[0], packetLength)
	}

	pkt := Create(header[0])
	if pkt == nil {
		if debug {
			log.Println("read: error create packet")
		}
		return nil, ErrUnknownPacket
	}

	if packetLength != 0 {
		payload := make([]byte, packetLength)
		if n, err := conn.Read(payload); n < packetLength || err != nil {
			if debug {
				log.Printf("read only %d bytes, try to read last one\n", n)
			}
			last, err := conn.Read(payload[n:])
			if err != nil {
				return nil, io.ErrUnexpectedEOF
			}
			if debug {
				log.Printf("read last %d bytes, seems fine now\n", last)
			}
		}

		if debug {
			log.Printf("read packet payload:\n%s", hex.Dump(payload))
		}

		pkt.Unpack(payload)
	}

	if debug {
		log.Println("read packet (decoded):", pkt.String())
	}

	return pkt, nil
}

func WritePacket(conn net.Conn, pkt Packet, debug bool) error {
	packed := pkt.Pack()

	if debug {
		log.Printf("write:\n%s", hex.Dump(packed))
	}

	_, err := conn.Write(packed)

	return err
}

func WriteLength(len int) []byte {
	var n int

	if len < 128 {
		n = 1
	} else if len < 16384 {
		n = 2
	} else if len < 2097152 {
		n = 3
	} else if len <= 268435455 {
		n = 4
	}
	buf := make([]byte, n)
	binary.PutUvarint(buf, uint64(len))
	return buf
}

func MatchTopic(mask string, topic string) bool {
	t := strings.Split(topic, "/")

	var found bool
	var i int = 0 // start from first level
	maskPart := strings.Split(mask, "/")

	for {
		if len(maskPart) <= i || len(t) <= i {
			break
		}

		if maskPart[i] == "#" {
			found = true
			break
		}

		// match at this level
		if maskPart[i] == "*" || maskPart[i] == t[i] {
			if len(t) == i+1 {
				if len(t) == len(maskPart) {
					found = true
					break
				}
				break
			}

			// try next level
			i++
			continue
		}

		// doesn't match
		break
	}

	return found
}
