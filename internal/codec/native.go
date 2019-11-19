package codec

import (
	"encoding/binary"
	"fmt"
	"github.com/serajam/sbucket/internal"
	"io"
	"reflect"
)

const (
	msgSizeLen    = 4
	dataSizeLen   = 4
	msgCmdLen     = 1
	bucketSizeLen = 1
	keySizeLen    = 1
	valSizeLen    = 4
	resultSizeLen = 1
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

	if msg.Command == internal.ResultCmd || msg.Command == internal.ResultValueCmd {
		return l.encodeResult(msg)
	}

	// define sizes
	bucketLen := len(msg.Bucket)
	keyLen := len(msg.Key)

	if msg.Command == internal.CreateBucketCmd || msg.Command == internal.DelBucketCmd {
		totalMsgSize := msgSizeLen + msgCmdLen + bucketSizeLen + bucketLen

		// allocate memory for message
		data := make([]byte, totalMsgSize)

		// put msg size
		binary.BigEndian.PutUint32(data[0:4], uint32(totalMsgSize-4))

		// put command bytes and bucket name
		data[4] = msg.Command
		pointer := msgSizeLen + msgCmdLen + bucketSizeLen - 1
		l.putBucket(pointer, bucketLen, data, msg)

		return nil
	}

	if msg.Command == internal.GetCmd {
		totalMsgSize := msgSizeLen + msgCmdLen + bucketSizeLen + bucketLen + keySizeLen + keyLen

		// allocate memory for message
		data := make([]byte, totalMsgSize)

		// put msg size
		binary.BigEndian.PutUint32(data[0:4], uint32(totalMsgSize-4))

		// put command bytes and bucket name
		data[4] = msg.Command
		pointer := msgSizeLen + msgCmdLen + bucketSizeLen - 1
		l.putBucket(pointer, bucketLen, data, msg)

		// put key
		data[pointer] = byte(keyLen)
		pointer++
		copy(data[pointer:pointer+keyLen], []byte(msg.Key))

		return nil
	}

	valLen := len(msg.Value)
	totalMsgSize := msgSizeLen + msgCmdLen + bucketSizeLen + bucketLen + keySizeLen + keyLen + valSizeLen + valLen

	// allocate memory for message
	data := make([]byte, totalMsgSize)

	// put msg size
	binary.BigEndian.PutUint32(data[0:4], uint32(totalMsgSize-4))

	// put command bytes and bucket name
	data[4] = msg.Command
	pointer := msgSizeLen + msgCmdLen + bucketSizeLen - 1

	pointer = l.putBucket(pointer, bucketLen, data, msg)

	// put key and val bytes
	l.putKey(pointer, keyLen, data, msg)
	l.putVal(pointer, valLen, data, msg)

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

	if msg.Command == internal.ResultCmd || msg.Command == internal.ResultValueCmd {
		return l.decodeResult(pointer, data, msg)
	}

	if msg.Command == internal.CreateBucketCmd || msg.Command == internal.DelBucketCmd {
		l.readBucket(pointer, data, msg)
		return nil
	}

	if msg.Command == internal.GetCmd {
		// read bucket name
		pointer = l.readBucket(pointer, data, msg)
		l.readKey(pointer, data, msg)

		return nil
	}

	pointer = l.readBucket(pointer, data, msg)
	pointer = l.readKey(pointer, data, msg)
	l.readVal(pointer, data, msg)

	return nil
}

func (l native) decodeResult(pointer int, data []byte, msg *Message) error {
	// read result
	pointer++
	result := data[pointer]
	if result == 1 {
		msg.Result = true
	}
	pointer++

	if !msg.Result {
		// read val
		dataLen := int(binary.BigEndian.Uint32(data[pointer : pointer+valSizeLen]))
		pointer += valSizeLen
		msg.Data = string(data[pointer : pointer+dataLen])
	}

	if msg.Command == internal.ResultValueCmd {
		l.readVal(pointer, data, msg)
	}

	return nil
}

func (l native) encodeResult(msg *Message) error {
	dataLen := len(msg.Data)

	totalMsgSize := msgSizeLen + msgCmdLen + resultSizeLen
	if dataLen > 0 {
		totalMsgSize += dataSizeLen + dataLen
	}
	if msg.Command == internal.ResultValueCmd {
		totalMsgSize += valSizeLen + len(msg.Value)
	}

	data := make([]byte, totalMsgSize)
	pointer := 0

	// put msg size
	binary.BigEndian.PutUint32(data[pointer:msgSizeLen], uint32(totalMsgSize-4))
	pointer += msgSizeLen

	// put command bytes and bucket name
	data[pointer] = msg.Command
	pointer++

	var result byte = 0
	if msg.Result {
		result = 1
	}
	data[pointer] = result
	pointer++

	if dataLen > 0 {
		binary.BigEndian.PutUint32(data[pointer:pointer+dataSizeLen], uint32(dataLen))
		pointer += dataSizeLen

		copy(data[pointer:pointer+dataLen], []byte(msg.Data))
		pointer += dataLen
	}

	if msg.Command == internal.ResultValueCmd {
		// put key and val bytes
		l.putVal(pointer, len(msg.Value), data, msg)
	}

	_, err := l.w.Write(data)
	if err != nil {
		return err
	}

	return nil
}

func (l native) putBucket(pointer int, bucketLen int, data []byte, msg *Message) int {
	data[pointer] = byte(bucketLen)
	pointer++
	copy(data[pointer:pointer+bucketLen], []byte(msg.Bucket))
	pointer += bucketLen

	return pointer
}

func (l native) readBucket(pointer int, data []byte, msg *Message) int {
	// read bucket name
	pointer++
	bucketLen := int(data[pointer])
	pointer++
	msg.Bucket = string(data[pointer : pointer+bucketLen])
	pointer += bucketLen

	return pointer
}

func (l native) putVal(pointer, valLen int, data []byte, msg *Message) {
	binary.BigEndian.PutUint32(data[pointer:pointer+valSizeLen], uint32(valLen))
	pointer += valSizeLen
	copy(data[pointer:pointer+valLen], []byte(msg.Value))
}

func (l native) readVal(pointer int, data []byte, msg *Message) {
	valLen := int(binary.BigEndian.Uint32(data[pointer : pointer+valSizeLen]))
	pointer += valSizeLen
	msg.Value = string(data[pointer : pointer+valLen])
}

func (l native) putKey(pointer, keyLen int, data []byte, msg *Message) int {
	data[pointer] = byte(keyLen)
	pointer++
	copy(data[pointer:pointer+keyLen], []byte(msg.Key))
	pointer += keyLen

	return pointer
}

func (l native) readKey(pointer int, data []byte, msg *Message) int {
	keyLen := int(data[pointer])
	pointer++
	msg.Key = string(data[pointer : pointer+keyLen])
	pointer += keyLen

	return pointer
}
