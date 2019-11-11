package server

import (
	"fmt"
	"github.com/serajam/sbucket/internal/codec"
	"github.com/serajam/sbucket/internal/storage"
	"math/rand"
	"net"
	"testing"
	"time"
)

func BenchmarkServer(b *testing.B) {
	strg, err := storage.New(storage.SyncMapType)
	if err != nil {
		b.Error(err)
		return
	}

	rand.Seed(time.Now().Unix())
	port := 8000 + rand.Intn(1000)

	s := New(strg,
		Address(fmt.Sprintf(":%d", port)),
		MaxConnNum(10000),
		ConnectTimeout(5),
		MaxFailures(5),
	)

	go func() {
		s.Start()
	}()
//	defer s.Shutdown()
	time.Sleep(500 * time.Millisecond)

	conn, err := net.Dial("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		b.Errorf("failed to dial server, err: %s", err)
		return
	}

	msg := codec.Message{
		Command: "DEL",
		Bucket:  "test",
		Key:     "test",
		Value:   "test",
		Data:    "some test data",
	}
	cod, _ := codec.New(codec.MsgPack, conn)

	for n := 0; n < b.N; n++ {
		cod.Encode(msg)
	}
}
