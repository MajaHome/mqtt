package packet

type PacketImpl struct {
	ClientId string
}

func (pi *PacketImpl) Source() string {
	return pi.ClientId
}

func (pi *PacketImpl) SetSource(clientId string) {
	pi.ClientId = clientId
}

func (pi *PacketImpl) Type() Type {
	return RESERVED
}

func (pi *PacketImpl) Length() int {
	return 0
}

func (pi *PacketImpl) Unpack(buf []byte) error {
	return nil
}

func (pi *PacketImpl) Pack() []byte {
	return []byte{}
}

func (pi *PacketImpl) String() string {
	return ""
}
