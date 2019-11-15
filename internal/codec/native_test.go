package codec

import (
	"github.com/bxcodec/faker/v3"
	"github.com/serajam/sbucket/internal"
	"io/ioutil"
	"net"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"
)

func BenchmarkNative_EncodeDecode(b *testing.B) {
	fa := faker.Lorem{}
	data, _ := fa.Sentence(reflect.Value{})

	msg := &Message{
		Command: internal.AddCommand,
		Bucket:  faker.Name(),
		Key:     faker.Name(),
		Value:   strings.Repeat(data.(string), 1),
	}

	pr, pw := net.Pipe()
	nativeCodec := newNativeCodec(pw, pr)

	b.ResetTimer()
	b.ReportAllocs()

	go func() {
		for {
			if err := nativeCodec.Decode(&Message{}); err != nil {
				b.Errorf("Encode: %s", err)
			}
		}
	}()

	for n := 0; n < b.N; n++ {
		if err := nativeCodec.Encode(msg); err != nil {
			b.Errorf("Encode: %s", err)
		}
	}
}

func BenchmarkNative_Encode(b *testing.B) {
	fa := faker.Lorem{}
	data, _ := fa.Paragraph(reflect.Value{})

	msg := &Message{
		Command: internal.AddCommand,
		Bucket:  faker.Name(),
		Key:     faker.Name(),
		Value:   strings.Repeat(data.(string), 100),
	}

	nativeCodec := newNativeCodec(ioutil.Discard, nil)

	b.ResetTimer()
	b.ReportAllocs()

	for n := 0; n < b.N; n++ {
		if err := nativeCodec.Encode(msg); err != nil {
			b.Errorf("Encode: %s", err)
		}
	}
}

func TestReading(t *testing.T) {
	fa := faker.Lorem{}
	data, _ := fa.Paragraph(reflect.Value{})

	msgSend := &Message{
		Command: internal.AddCommand,
		Bucket:  faker.Name(),
		Key:     faker.Name(),
		Value:   data.(string),
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

	nativeCodec.Encode(msgSend)
	time.Sleep(100 * time.Millisecond)
	pw.Close()
}

func Test_native_Decode(t *testing.T) {
	type args struct {
		msg  *Message
		cmd  uint8
		len  string
		data string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			"Test encode type long cmd",
			args{msg: &Message{}, cmd: internal.DeleteCommand, len: "19", data: "and some other data"},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			server, client := net.Pipe()
			l := native{w: server}
			go func() {
				client.Write([]byte(strconv.Itoa(int(tt.args.cmd)) + " " + tt.args.len + " " + tt.args.data))
			}()

			time.Sleep(1 * time.Second)
			if err := l.Decode(tt.args.msg); (err != nil) != tt.wantErr {
				t.Errorf("Decode() error = %v, wantErr %v", err, tt.wantErr)
				server.Close()
				client.Close()

				return
			}

			if tt.args.msg.Command != tt.args.cmd {
				t.Errorf("Exptected %d but got %d", tt.args.cmd, tt.args.msg.Command)
			}

			if tt.args.msg.Data != tt.args.data {
				t.Errorf("Exptected `%s` but got `%s`", tt.args.data, tt.args.msg.Data)
			}

			server.Close()
			client.Close()

		})
	}
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

func Test_newNativeCodec(t *testing.T) {
	type args struct {
		c net.Conn
	}
	tests := []struct {
		name string
		args args
		want Codec
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := newNativeCodec(tt.args.c, tt.args.c); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newNativeCodec() = %v, want %v", got, tt.want)
			}
		})
	}
}
