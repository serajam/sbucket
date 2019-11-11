package codec

import (
	"bufio"
	"fmt"
	"net"
	"reflect"
	"strconv"
	"strings"
)

type native struct {
	c net.Conn
	r *bufio.Reader
}

func newNativeCodec(c net.Conn) Codec {
	return &native{c: c, r: bufio.NewReader(c)}
}

func (l native) Encode(msg interface{}) error {
	return nil
}

func (l native) Decode(msg interface{}) error {
	message, ok := msg.(*Message)
	if !ok {
		return fmt.Errorf("failed to define a msg type for %v", reflect.TypeOf(msg))
	}

	op, err := l.r.ReadString(' ')
	if err != nil {
		return err
	}

	op = strings.TrimRight(op, " ")
	if len(op) != 3 {
		return fmt.Errorf("error occured while trying to read operation type: received %d bytes", len(op))
	}

	lenStr, err := l.r.ReadString(' ')
	lenStr = strings.TrimRight(lenStr, " ")
	if err != nil {
		return err
	}
	len, err := strconv.Atoi(lenStr)
	if len == 0 {
		return fmt.Errorf("error occured while trying to read content len: received %s bytes", lenStr)
	}
	message.Command = op

	content := make([]byte, len)
	_, err = l.r.Read(content)
	if err != nil {
		return err
	}

	message.Data = string(content)

	return nil
}
