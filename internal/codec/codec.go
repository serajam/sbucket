package codec

import (
	"fmt"
	"net"
)

// Type of codecs implemented
const (
	Native = 1
	Gob    = 2
)

// New creates new codec depending on type
func New(t int, c net.Conn) (Codec, error) {
	switch t {
	case Native:
	case Gob:
		return newGobCodec(c), nil
	}

	return nil, fmt.Errorf("codec creation failed. Unknown type %d specified", t)
}
