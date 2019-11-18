package codec

import (
	"encoding/binary"
	"fmt"
	"io"
	"reflect"
)

const (
	msgSizeLen    = 4
	msgCmdLen     = 1
	bucketSizeLen = 1
	keySizeLen    = 1
	valSizeLen    = 4
)

type native struct {
	w io.Writer
	r io.Reader
}

func newNativeCodec(w io.Writer, r io.Reader) Codec {
	return &native{w: w, r: r}
}

// Encode encodes message using byte slice
func (l native) Encode(m interface{}) error {
	msg, ok := m.(*Message)
	if !ok {
		return fmt.Errorf("failed to define a msg type for %v", reflect.TypeOf(msg))
	}

	// define sizes
	bucketLen := len(msg.Bucket)
	keyLen := len(msg.Key)
	valLen := len(msg.Value)
	totalMsgSize := msgSizeLen + msgCmdLen + bucketSizeLen + bucketLen + keySizeLen + keyLen + valSizeLen + valLen

	// allocate memory for message
	data := make([]byte, totalMsgSize)

	// put msg size
	binary.BigEndian.PutUint32(data[0:4], uint32(totalMsgSize-4))

	// put command bytes and bucket name
	data[4] = msg.Command
	pointer := msgSizeLen + msgCmdLen + bucketSizeLen - 1
	data[pointer] = byte(bucketLen)
	pointer++
	copy(data[pointer:pointer+bucketLen], []byte(msg.Bucket))
	pointer += bucketLen

	// put key and val bytes
	data[pointer] = byte(keyLen)
	pointer++
	copy(data[pointer:pointer+keyLen], []byte(msg.Key))
	pointer += keyLen
	binary.BigEndian.PutUint32(data[pointer:pointer+valSizeLen], uint32(valLen))
	pointer += valSizeLen
	copy(data[pointer:pointer+valLen], []byte(msg.Value))

	_, err := l.w.Write(data)
	if err != nil {
		return err
	}

	return nil
}

// Decode decodes message using byte slice
// Read first 4 bytes to determine msg len
func (l native) Decode(m interface{}) error {
	msg, ok := m.(*Message)
	if !ok {
		return fmt.Errorf("failed to define a msg type for %v", reflect.TypeOf(msg))
	}

	// read msg size
	data := make([]byte, msgSizeLen)
	n, err := io.ReadAtLeast(l.r, data, 4)
	if err != nil {
		return err
	}

	if n < 4 {
		return fmt.Errorf("failed to read message size")
	}

	currMsgSize := int(binary.BigEndian.Uint32(data))
	data = make([]byte, currMsgSize)
	n, err = io.ReadFull(l.r, data)
	if err != nil {
		return err
	}

	// read cmd
	pointer := 0
	cmd := data[pointer]
	msg.Command = cmd

	// read bucket name
	pointer++
	bucketLen := int(data[pointer])
	pointer++
	msg.Bucket = string(data[pointer : pointer+bucketLen])
	pointer += bucketLen

	// read key
	keyLen := int(data[pointer])
	pointer++
	msg.Key = string(data[pointer : pointer+keyLen])
	pointer += keyLen

	// read val
	valLen := int(binary.BigEndian.Uint32(data[pointer : pointer+valSizeLen]))
	pointer += valSizeLen
	msg.Value = string(data[pointer : pointer+valLen])

	return nil
}
