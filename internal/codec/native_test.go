package codec

import (
	"github.com/bxcodec/faker/v3"
	"github.com/serajam/sbucket/internal"
	"net"
	"reflect"
	"testing"
	"time"
)

func TestResultTrue(t *testing.T) {
	msgSend := &Message{
		Command: internal.ResultCmd,
		Result:  true,
	}

	pr, pw := net.Pipe()

	nativeCodec := newNativeCodec(pw, pr)

	go func() {
		defer pr.Close()
		msgRecv := &Message{}
		err := nativeCodec.Decode(msgRecv)
		if err != nil {
			t.Errorf("Decode: %s", err)
		}

		if !reflect.DeepEqual(msgSend, msgRecv) {
			t.Error("Messages are not equal")
		}
	}()

	time.Sleep(100 * time.Millisecond)
	err := nativeCodec.Encode(msgSend)
	if err != nil {
		t.Error(err)
	}
	pw.Close()
}

func TestResultValue(t *testing.T) {
	msgSend := &Message{
		Command: internal.ResultValueCmd,
		Result:  true,
		Value:   faker.Paragraph(),
	}

	pr, pw := net.Pipe()

	nativeCodec := newNativeCodec(pw, pr)

	go func() {
		defer pr.Close()
		msgRecv := &Message{}
		err := nativeCodec.Decode(msgRecv)
		if err != nil {
			t.Errorf("Decode: %s", err)
		}

		if !reflect.DeepEqual(msgSend, msgRecv) {
			t.Error("Messages are not equal")
		}
	}()

	time.Sleep(100 * time.Millisecond)
	err := nativeCodec.Encode(msgSend)
	if err != nil {
		t.Error(err)
	}
	pw.Close()
}

func TestGetValue(t *testing.T) {
	msgSend := &Message{
		Command: internal.GetCmd,
		Bucket:  faker.Name(),
		Key:     faker.Name(),
	}

	pr, pw := net.Pipe()

	nativeCodec := newNativeCodec(pw, pr)

	go func() {
		defer pr.Close()
		msgRecv := &Message{}
		err := nativeCodec.Decode(msgRecv)
		if err != nil {
			t.Errorf("Decode: %s", err)
		}

		if !reflect.DeepEqual(msgSend, msgRecv) {
			t.Error("Messages are not equal")
		}
	}()

	time.Sleep(100 * time.Millisecond)
	err := nativeCodec.Encode(msgSend)
	if err != nil {
		t.Error(err)
	}
	pw.Close()
}

func TestCreateBucket(t *testing.T) {
	msgSend := &Message{
		Command: internal.DelBucketCmd,
		Bucket:  faker.Name(),
	}

	pr, pw := net.Pipe()

	nativeCodec := newNativeCodec(pw, pr)

	go func() {
		defer pr.Close()
		msgRecv := &Message{}
		err := nativeCodec.Decode(msgRecv)
		if err != nil {
			t.Errorf("Decode: %s", err)
		}

		if !reflect.DeepEqual(msgSend, msgRecv) {
			t.Error("Messages are not equal")
		}
	}()

	time.Sleep(100 * time.Millisecond)
	err := nativeCodec.Encode(msgSend)
	if err != nil {
		t.Error(err)
	}
	pw.Close()
}

func TestResultFalse(t *testing.T) {
	msgSend := &Message{
		Command: internal.ResultCmd,
		Data:    faker.Sentence(),
	}

	pr, pw := net.Pipe()

	nativeCodec := newNativeCodec(pw, pr)

	go func() {
		defer pr.Close()
		msgRecv := &Message{}
		err := nativeCodec.Decode(msgRecv)
		if err != nil {
			t.Errorf("Decode: %s", err)
		}

		if !reflect.DeepEqual(msgSend, msgRecv) {
			t.Error("Messages are not equal")
		}
	}()

	time.Sleep(100 * time.Millisecond)
	err := nativeCodec.Encode(msgSend)
	if err != nil {
		t.Error(err)
	}
	pw.Close()
}

func TestReading(t *testing.T) {
	msgSend := &Message{
		Command: internal.AddCmd,
		Bucket:  faker.Name(),
		Key:     faker.Name(),
		Value:   faker.Paragraph(),
	}

	pr, pw := net.Pipe()

	nativeCodec := newNativeCodec(pw, pr)

	go func() {
		defer pr.Close()
		msgRecv := &Message{}
		err := nativeCodec.Decode(msgRecv)
		if err != nil {
			t.Errorf("Decode: %s", err)
		}

		if !reflect.DeepEqual(msgSend, msgRecv) {
			t.Error("Messages are not equal")
		}
	}()

	time.Sleep(100 * time.Millisecond)
	err := nativeCodec.Encode(msgSend)
	if err != nil {
		t.Error(err)
	}
	pw.Close()
}

func Test_native_Encode(t *testing.T) {
	type fields struct {
		c net.Conn
	}
	type args struct {
		msg interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := native{
				w: tt.fields.c,
			}
			if err := l.Encode(tt.args.msg); (err != nil) != tt.wantErr {
				t.Errorf("Encode() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
