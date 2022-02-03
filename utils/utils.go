package utils

import (
	"encoding/binary"
	"errors"
)

var (
	ErrReadFromBuf = errors.New("error read data from buffer")
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
