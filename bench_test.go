package main

import (
	"encoding/hex"
	"fmt"
	"github.com/serajam/sbucket/internal/client"
	"github.com/serajam/sbucket/internal/server"
	"github.com/serajam/sbucket/internal/storage"
	"log"
	"math/rand"
	"os"
	"sync"
	"testing"
	"time"
)

var clnt client.Client
var err error

func randUUID(uuid []byte) {
	b := make([]byte, 16)

	_, err := rand.Read(b)
	if err != nil {
		log.Fatal(err)
	}
	copy(uuid, b)
}

func init() {
}

func TestMain(m *testing.M) {
	db, err := storage.New(storage.SyncMapType)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	s := server.New(db,
		server.Address(":3456"),
		server.MaxConnNum(1000),
		server.ConnDeadline(1),
	)
	go func() {
		s.Start()
	}()
	time.Sleep(500 * time.Millisecond)

	clnt, err = client.NewClient(":3456", 4, 5, "", "", 2)
	if err != nil {
		log.Println(err)
		return
	}

	//defer clnt.Close()

	retCode := m.Run()

	s.Shutdown()
	os.Exit(retCode)
}

func BenchmarkClient_Add(b *testing.B) {
	data := make([]string, 0, 1000000)

	uuid := make([]byte, 16)
	rand.Seed(time.Now().Unix() + int64(rand.Int()))

	for n := 0; n < 1000000; n++ {
		randUUID(uuid)
		data[n] = hex.EncodeToString(uuid)
	}

	var dataVal string
	var errs []error
	i := 0

	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		fails := 0
		for pb.Next() {
			dataVal = data[rand.Intn(i+1000000)]
			err = clnt.Add("TEST1", dataVal, dataVal)
			if err != nil {
				fails++
				errs = append(errs, err)
			}
			i++
		}
	})
	clnt.Wait()
}

func BenchmarkClient2_Add(b *testing.B) {
	elem := b.N
	data := make([]string, 0, elem)
	uuid := make([]byte, 16)
	rand.Seed(time.Now().Unix() + int64(rand.Int()))

	for n := 0; n < elem; n++ {
		randUUID(uuid)
		data = append(data, hex.EncodeToString(uuid))
	}

	e := clnt.DeleteBucket("TEST2")
	if e != nil {
		b.Error(e)
		return
	}

	e = clnt.CreateBucket("TEST2")
	if e != nil {
		b.Error(e)
		return
	}

	b.ReportAllocs()
	b.ResetTimer()

	var wg sync.WaitGroup
	for n := 0; n < elem; n++ {
		wg.Add(1)
		go func(k int) {
			defer wg.Done()
			err = clnt.Add("TEST2", data[k], data[k])
			if err != nil {
				fmt.Println(err)
			}
		}(n)
	}

	wg.Wait()
}

func BenchmarkClient4_Add(b *testing.B) {
	data := make([]string, 0, 1000000)

	uuid := make([]byte, 16)
	rand.Seed(time.Now().Unix() + int64(rand.Int()))

	for n := 0; n < 1000000; n++ {
		randUUID(uuid)
		data[n] = hex.EncodeToString(uuid)
	}

	clnt.DeleteBucket("TEST4")
	err := clnt.CreateBucket("TEST4")
	if err != nil {
		b.Error(err)
	}

	b.ReportAllocs()
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		go func(k int) {
			err = clnt.Add("TEST4", data[k], data[k])
			if err != nil {
			}
		}(n)
	}
}

func BenchmarkClient8_Add(b *testing.B) {
	data := make([]string, 0, 1000000)

	uuid := make([]byte, 16)
	rand.Seed(time.Now().Unix() + int64(rand.Int()))

	for n := 0; n < 1000000; n++ {
		randUUID(uuid)
		data[n] = hex.EncodeToString(uuid)
	}

	clnt.DeleteBucket("TEST8")
	err := clnt.CreateBucket("TEST8")
	if err != nil {
		b.Error(err)
	}
	parr := 8
	sem := make(chan struct{}, parr)
	rand.Seed(555)

	b.ReportAllocs()
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		sem <- struct{}{}
		go func(k int) {
			err = clnt.Add("TEST8", data[k], data[k])
			<-sem
			if err != nil {
			}
		}(n)
	}

}

func BenchmarkClient16_Add(b *testing.B) {
	data := make([]string, 0, 1000000)

	uuid := make([]byte, 16)
	rand.Seed(time.Now().Unix() + int64(rand.Int()))

	for n := 0; n < 1000000; n++ {
		randUUID(uuid)
		data[n] = hex.EncodeToString(uuid)
	}

	clnt.DeleteBucket("TEST16")
	err := clnt.CreateBucket("TEST16")
	if err != nil {
		b.Error(err)
	}
	parr := 16
	sem := make(chan struct{}, parr)
	rand.Seed(555)

	b.ReportAllocs()
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		sem <- struct{}{}
		go func(k int) {
			err = clnt.Add("TEST16", data[k], data[k])
			<-sem
			if err != nil {
			}
		}(n)
	}

}

func BenchmarkClient_Get(b *testing.B) {
	err := clnt.CreateBucket("TEST")
	if err != nil {
		b.Error(err)
	}
	uuid := make([]byte, 16)
	rand.Seed(time.Now().Unix() + int64(rand.Int()))
	randUUID(uuid)
	v := hex.EncodeToString(uuid)

	err = clnt.Add("TEST", v, v)
	if err != nil {
		b.Error(err)
	}

	b.ReportAllocs()
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		_, err = clnt.Get("TEST", v)
		if err != nil {
			b.Error(err)
			return
		}
	}
}
