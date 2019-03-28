package main

import (
	"fmt"
	"github.com/serajam/sbucket/pkg/client"
	"log"
	"math/rand"
	"os"
	"strconv"
	"testing"
	"time"
)

var data = map[int]string{}
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
	fmt.Println("init")
}

func TestMain(m *testing.M) {
	fmt.Println("main")
	uuid := make([]byte, 16)
	rand.Seed(345)
	for n := 0; n < 1000000; n++ {
		randUUID(uuid)
		data[n] = fmt.Sprintf("%x-%d", uuid, n)
	}

	clnt, err = client.NewGobClient(":3456", 4, 5, "", "")
	if err != nil {
		log.Println(err)
		return
	}

	defer clnt.Close()

	clnt.DeleteBucket("TEST1")
	err = clnt.CreateBucket("TEST1")
	if err != nil {
		log.Println(err)
		return
	}

	retCode := m.Run()

	os.Exit(retCode)
}

func BenchmarkTime(b *testing.B) {
	time.Sleep(100 * time.Second)
}

func BenchmarkClient_Add(b *testing.B) {
	var dataVal string
	b.RunParallel(func(pb *testing.PB) {
		fails := 0
		for pb.Next() {
			dataVal = data[rand.Intn(1000000)]
			err = clnt.Add("TEST1", dataVal, dataVal)
			if err != nil {
				fails++
			}
		}
		fmt.Println(fails)
	})

	clnt.Wait()
}

func BenchmarkClient2_Add(b *testing.B) {
	clnt.DeleteBucket("TEST2")
	err := clnt.CreateBucket("TEST2")
	if err != nil {
		b.Error(err)
	}

	for n := 0; n < b.N; n++ {
		go func() {
			err = clnt.Add("TEST2", data[n], data[n])
			if err != nil {
			}
		}()
	}
}

func BenchmarkClient4_Add(b *testing.B) {
	clnt.DeleteBucket("TEST4")
	err := clnt.CreateBucket("TEST4")
	if err != nil {
		b.Error(err)
	}

	for n := 0; n < b.N; n++ {
		go func() {
			err = clnt.Add("TEST4", data[n], data[n])
			if err != nil {
			}
		}()
	}
}

func BenchmarkClient8_Add(b *testing.B) {
	clnt.DeleteBucket("TEST8")
	err := clnt.CreateBucket("TEST8")
	if err != nil {
		b.Error(err)
	}
	parr := 8
	sem := make(chan struct{}, parr)
	rand.Seed(555)

	for n := 0; n < b.N; n++ {
		sem <- struct{}{}
		go func() {
			err = clnt.Add("TEST8", data[n], data[n])
			<-sem
			if err != nil {
			}
		}()
	}

}

func BenchmarkClient16_Add(b *testing.B) {
	clnt.DeleteBucket("TEST16")
	err := clnt.CreateBucket("TEST16")
	if err != nil {
		b.Error(err)
	}
	parr := 16
	sem := make(chan struct{}, parr)
	rand.Seed(555)

	for n := 0; n < b.N; n++ {
		sem <- struct{}{}
		go func() {
			err = clnt.Add("TEST16", data[n], data[n])
			<-sem
			if err != nil {
			}
		}()
	}

}

func BenchmarkClient_Get(b *testing.B) {
	err := clnt.CreateBucket("TEST")
	if err != nil {
		b.Error(err)
	}

	rand.Seed(555)
	v := strconv.Itoa(rand.Intn(100000000))
	err = clnt.Add("TEST", v, v)
	if err != nil {
		b.Error(err)
	}

	for n := 0; n < b.N; n++ {
		if err != nil {
			b.Error(err)
		}
	}
}
