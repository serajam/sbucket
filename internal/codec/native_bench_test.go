package codec

import (
	"github.com/bxcodec/faker/v3"
	"github.com/serajam/sbucket/internal"
	"io/ioutil"
	"net"
	"reflect"
	"strings"
	"testing"
)

func BenchmarkNative_EncodeDecode(b *testing.B) {
	fa := faker.Lorem{}
	data, _ := fa.Sentence(reflect.Value{})

	msg := &Message{
		Command: internal.AddCmd,
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
		Command: internal.AddCmd,
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
