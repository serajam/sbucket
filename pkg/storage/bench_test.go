package storage

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"testing"
)

var err error
var data map[int]string

func randUUID(uuid []byte) {
	b := make([]byte, 16)

	_, err = rand.Read(b)
	if err != nil {
		log.Fatal(err)
	}
	copy(uuid, b)
}

func TestMain(m *testing.M) {
	data = make(map[int]string)
	uuid := make([]byte, 16)
	rand.Seed(345)
	for n := 0; n < 10000000; n++ {
		randUUID(uuid)
		data[n] = fmt.Sprintf("%x-%d", uuid, n)
	}
	retCode := m.Run()
	os.Exit(retCode)
}

func BenchmarkSBucketMapStorage_Add(b *testing.B) {
	s := newSBucketMapStorage()

	err = s.NewBucket("TEST")
	if err != nil {
		b.Error(err)
		return
	}

	var v string

	for n := 0; n < b.N; n++ {
		v = data[n]
		s.Add("TEST", v, v)
	}
}

func BenchmarkSBucketMapStorage_Get(b *testing.B) {
	s := newSBucketMapStorage()

	err = s.NewBucket("TEST")
	err = s.Add("TEST", data[0], data[0])
	if err != nil {
		b.Error(err)
		return
	}

	for n := 0; n < b.N; n++ {
		_, err = s.Get("TEST", data[0])
		if err != nil {
			b.Error(err)
			return
		}
	}
}

func BenchmarkSBucketSyncMapStorage_Add(b *testing.B) {
	s := newSBucketSyncMapStorage()

	err = s.NewBucket("TEST")
	if err != nil {
		b.Error(err)
		return
	}

	var v string

	for n := 0; n < b.N; n++ {
		v = data[n]
		s.Add("TEST", v, v)
	}
}

func BenchmarkSBucketSyncMapStorage_Get(b *testing.B) {
	s := newSBucketSyncMapStorage()

	err = s.NewBucket("TEST")
	err = s.Add("TEST", data[0], data[0])
	if err != nil {
		b.Error(err)
		return
	}

	for n := 0; n < b.N; n++ {
		_, err = s.Get("TEST", data[0])
		if err != nil {
			b.Error(err)
			return
		}
	}
}
