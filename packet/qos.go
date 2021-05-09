package packet

import (
	"errors"
	"strconv"
)

var ErrInvalidQos = errors.New("invalid qos")

const (
	AtMostOnce  QoS = 0x00
	AtLeastOnce QoS = 0x01
	ExactlyOnce QoS = 0x02
)

type QoS byte

func (q QoS) Valid() bool {
	return q == 0x0 || q == 0x1 || q == 0x2
}

func (q QoS) Int() int {
	return int(q)
}
func (q QoS) String() string {
	return strconv.Itoa(int(q))
}
