package transport

import (
	"encoding/hex"
	"github.com/MajaSuite/mqtt/packet"
	"io"
	"log"
	"net"
)

func ReadPacket(conn net.Conn) (packet.Packet, error) {
	header := make([]byte, 2)
	n, err := conn.Read(header)
	if n < 2 || err != nil {
		return nil, io.ErrUnexpectedEOF
	}

	log.Println("read header:\n", hex.Dump(header))

	pkt, err := packet.Create(header)
	if err != nil {
		log.Println("read: error create packet")
		return nil, packet.ErrUnknownPacket
	}

	packetLength := header[1]
	if packetLength != 0 {
		payload := make([]byte, packetLength)
		n, err = conn.Read(payload)
		if n < int(packetLength) || err != nil {
			return nil, io.ErrUnexpectedEOF
		}

		log.Println("read packet payload:\n", hex.Dump(payload))

		pkt.Unpack(payload)
	}

	log.Println("read packet:", pkt.String())

	return pkt, nil
}

func WritePacket(conn net.Conn, pkt packet.Packet) error {
	packed := pkt.Pack()

	log.Printf("write:\n%s\n", hex.Dump(packed))

	_, err := conn.Write(packed)
	if err != nil {
		log.Println("write: ", err)
	}

	return err
}
