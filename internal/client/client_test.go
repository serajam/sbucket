package client

import (
	"github.com/serajam/sbucket/internal/codec"
	"testing"
	"time"
)

func TestClient_CreateBucket(t *testing.T) {
	time.Sleep(1 * time.Second)

	client, err := NewClient(":3456", 4, 1, "", "", codec.MsgPack)
	if err != nil {
		t.Error(err)
		return
	}
	defer client.Close()

	client.DeleteBucket("TEST")
	err = client.CreateBucket("TEST")
	if err != nil {
		t.Error(err)
		return
	}

	err = client.Add("TEST", "1", "USER 456")
	if err != nil {
		t.Error(err)
		return
	}

	_, err = client.Get("TEST", "1")
	if err != nil {
		t.Error(err)
		return
	}
}
