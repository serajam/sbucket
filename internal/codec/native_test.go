package codec

import (
	"bufio"
	"net"
	"reflect"
	"testing"
	"time"
)

func Test_native_Decode(t *testing.T) {
	type args struct {
		msg  *Message
		cmd  string
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
			args{msg: &Message{}, cmd: "DEL", len: "19", data: "and some other data"},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			server, client := net.Pipe()
			l := native{c: server, r: bufio.NewReader(server)}
			go func() {
				client.Write([]byte(tt.args.cmd + " " + tt.args.len + " " + tt.args.data))
			}()

			time.Sleep(1 * time.Second)
			if err := l.Decode(tt.args.msg); (err != nil) != tt.wantErr {
				t.Errorf("Decode() error = %v, wantErr %v", err, tt.wantErr)
				server.Close()
				client.Close()

				return
			}

			if tt.args.msg.Command != tt.args.cmd {
				t.Errorf("Exptected %s but got %s", tt.args.cmd, tt.args.msg.Command)
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
				c: tt.fields.c,
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
			if got := newNativeCodec(tt.args.c); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newNativeCodec() = %v, want %v", got, tt.want)
			}
		})
	}
}
