package packet

import "strconv"

const (
	AtMostOnce QoS = 0x00
	AtLeastOnce QoS = 0x01
	ExactlyOnce QoS = 0x02
)

type QoS byte

func (q QoS) Valid() bool {
	return q == 0x00 || q == 0x01 || q == 0x02
}

func (q QoS) ToString() string {
	return strconv.Itoa(int(q))
}