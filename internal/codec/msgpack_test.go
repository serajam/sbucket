package codec

import (
	ms "github.com/vmihailenco/msgpack"
	"net"
	"reflect"
	"syreclabs.com/go/faker"
	"testing"
)

func Test_msgpack_Decode(t *testing.T) {
	type args struct {
		msg *Message
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			"Test decoding",
			args{&Message{
				Command: "DEL",
				Bucket:  "test1",
				Key:     "key1",
				Value:   "val1",
				Data:    faker.Lorem().Characters(1000),
			}},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			server, client := net.Pipe()
			enc := ms.NewEncoder(client)
			l := msgpack{enc: ms.NewEncoder(server), dec: ms.NewDecoder(server)}
			go func() {
				enc.Encode(tt.args.msg)
			}()

			exptectedMsg := Message{}
			if err := l.Decode(&exptectedMsg); (err != nil) != tt.wantErr {
				t.Errorf("Decode() error = %v, wantErr %v", err, tt.wantErr)
				server.Close()
				client.Close()
				return
			}

			if reflect.DeepEqual(exptectedMsg, tt.args.msg) {
				t.Errorf("Decode() Want = %v, got %v", tt.args.msg, exptectedMsg)
			}

			server.Close()
			client.Close()
		})
	}
}

func Test_msgpack_Encode(t *testing.T) {
	type fields struct {
		enc *ms.Encoder
		dec *ms.Decoder
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
			m := msgpack{
				enc: tt.fields.enc,
				dec: tt.fields.dec,
			}
			if err := m.Encode(tt.args.msg); (err != nil) != tt.wantErr {
				t.Errorf("Encode() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_newMsgPackCodec(t *testing.T) {
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
			if got := newMsgPackCodec(tt.args.c); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newMsgPackCodec() = %v, want %v", got, tt.want)
			}
		})
	}
}
