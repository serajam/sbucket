package codec

// Codec is the interface that unites encoding and decoding
type Codec interface {
	Encoder
	Decoder
}

// Encoder is the interface that wraps encoding and writing message to connection
type Encoder interface {
	Encode(interface{}) error
}

// Decoder is the interface that wraps reading and decoding message from connection
type Decoder interface {
	Decode(interface{}) error
}
