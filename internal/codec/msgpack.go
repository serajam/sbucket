package codec

import (
	ms "github.com/vmihailenco/msgpack"
	"net"
)

type msgpack struct {
	enc *ms.Encoder
	dec *ms.Decoder
}

func newMsgPackCodec(c net.Conn) Codec {
	return &msgpack{enc: ms.NewEncoder(c), dec: ms.NewDecoder(c)}
}

// Encode encodes msg
func (m msgpack) Encode(msg interface{}) error {
	return m.enc.Encode(msg)
}

// Decode decodes msg
func (m msgpack) Decode(msg interface{}) error {
	return m.dec.Decode(msg)
}
