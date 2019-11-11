package codec

import (
	"fmt"
	"net"
)

// Type of codecs implemented
const (
	Native  = 1
	Gob     = 2
	MsgPack = 3
)

// New creates new codec depending on type
func New(t int, c net.Conn) (Codec, error) {
	switch t {
	case Native:
		panic("Unimpleneted")
	case Gob:
		return newGobCodec(c), nil
	case MsgPack:
		return newMsgPackCodec(c), nil
	default:
		return newMsgPackCodec(c), nil
	}

	return nil, fmt.Errorf("codec creation failed. Unknown type %d specified", t)
}
