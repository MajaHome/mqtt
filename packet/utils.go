package packet

import (
	"encoding/binary"
	"encoding/hex"
	"io"
	"log"
	"net"
	"strings"
)

func ReadInt8(buf []byte, offset int) (uint8, int, error) {
	var value uint8

	if len(buf) >= (offset + 1) {
		value = buf[offset]
		return value, offset + 1, nil
	}

	return 0, offset, ErrReadFromBuf
}

func WriteInt8(buf []byte, offset int, value uint8) int {
	buf[offset] = value
	offset++
	return offset
}

func ReadInt16(buf []byte, offset int) (uint16, int, error) {
	var value uint16

	if len(buf) >= (offset + 2) {
		value = uint16(buf[offset+1]) | uint16(buf[offset])<<8
		return value, offset + 2, nil
	}

	return 0, offset, ErrReadFromBuf
}

func WriteInt16(buf []byte, offset int, value uint16) int {
	binary.BigEndian.PutUint16(buf[offset:], value)
	offset += 2
	return offset
}

func ReadBytes(buf []byte, offset int, length int) ([]byte, int, error) {
	var value []byte

	if len(buf) >= (offset + length) {
		value = buf[offset : offset+length]
		return value, offset + length, nil
	}

	return nil, offset, ErrReadFromBuf
}

// Write bytes array length and array itself into buffer
// return offset after write
func WriteBytes(buf []byte, offset int, bytes []byte) int {
	copy(buf[offset:], bytes)
	return offset + len(bytes)
}

func ReadString(buf []byte, offset int, length int) (string, int, error) {
	s, i, e := ReadBytes(buf, offset, length)
	return string(s), i, e
}

// Write string len and string itself into buffer
// return offset after write
func WriteString(buf []byte, offset int, value string) int {
	offset = WriteInt16(buf, offset, uint16(len(value)))
	copy(buf[offset:], value)
	return offset + len(value)
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

			packetLength += int(add[0]) - 1
			if add[0] <= 127 {
				break
			}
		}
	}

	log.Printf("read header: 0x%x, %d bytes\n", header[0], packetLength)

	pkt := Create(header[0])
	if pkt == nil {
		log.Println("read: error create packet")
		return nil, ErrUnknownPacket
	}

	if packetLength != 0 {
		payload := make([]byte, packetLength)
		if n, err := conn.Read(payload); n < packetLength || err != nil {
			log.Printf("read only %d bytes, try to read last one\n", n)
			last, err := conn.Read(payload[n:])
			if err != nil {
				return nil, io.ErrUnexpectedEOF
			}
			log.Printf("read last %d bytes, seems fine now\n", last)
		}

		if debug {
			log.Printf("read packet payload:\n%s", hex.Dump(payload))
		}

		pkt.Unpack(payload)
	}

	log.Println("read packet:", pkt.String())

	return pkt, nil
}

func WritePacket(conn net.Conn, pkt Packet, debug bool) error {
	packed := pkt.Pack()

	if debug {
		log.Printf("write:\n%s", hex.Dump(packed))
	}

	_, err := conn.Write(packed)
	if err != nil {
		log.Println("write: ", err)
	}

	return err
}

func WriteLength(len int) []byte {
	var n int
	if len < 128 {
		n = 1
	} else if n < 16384 {
		n = 2
	} else if n < 2097152 {
		n = 3
	} else if n <= 268435455 {
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
