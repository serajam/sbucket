package codec

import (
	"encoding/gob"
	"net"
)

type gobCodec struct {
	dec *gob.Decoder
	enc *gob.Encoder
}

func newGobCodec(c net.Conn) Codec {
	return &gobCodec{dec: gob.NewDecoder(c), enc: gob.NewEncoder(c)}
}

// Encode encodes and writes message
func (c *gobCodec) Encode(msg interface{}) error {
	return c.enc.Encode(msg)
}

// Decode reads and decodes message
func (c *gobCodec) Decode(msg interface{}) error {
	return c.dec.Decode(msg)
}
