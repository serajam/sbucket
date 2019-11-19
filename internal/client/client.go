package client

import (
	"context"
	"errors"
	"fmt"
	"github.com/serajam/sbucket/internal"
	"github.com/serajam/sbucket/internal/codec"
	"net"
	"strings"
	"sync"
	"time"
)

type client struct {
	conn  net.Conn
	inUse bool

	enc codec.Encoder
	dec codec.Decoder
}

type db struct {
	mu sync.RWMutex

	address     string
	dialTimeout time.Duration

	conn   []*client
	active int

	awaitingConn int
	requiredConn chan *client

	pingInterval int
	codecType    int
}

// NewClient creates 	a pool of connections
func NewClient(address string, connNum int, dialTimeout time.Duration, login, pass string, codecType int) (Client, error) {
	dbConn := &db{
		conn:         make([]*client, 0),
		requiredConn: make(chan *client, connNum),
		address:      address,
		dialTimeout:  dialTimeout,
		pingInterval: 0,
		codecType:    codecType,
	}

	for i := connNum; i > 0; i-- {
		c, err := dbConn.newConn()
		if err != nil {
			return nil, err
		}

		if len(login) > 0 && len(pass) > 0 {
			err = c.authenticate(login, pass)
			if err != nil {
				return nil, err
			}
		}

		dbConn.conn = append(dbConn.conn, c)
	}

	//go dbConn.ping()

	return dbConn, nil
}

func (d db) newConn() (*client, error) {
	c, err := net.DialTimeout("tcp", d.address, d.dialTimeout*time.Second)
	if err != nil {
		return nil, fmt.Errorf("failed to open storage connection: %s", err)
	}

	encdec, _ := codec.New(d.codecType, c)

	return &client{conn: c, enc: encdec, dec: encdec}, nil
}

func (d *db) ping() {
	if d.pingInterval == 0 {
		return
	}
	ticker := time.NewTicker(time.Duration(d.pingInterval) * time.Second)
	for {
		select {
		case <-ticker.C:
			d.mu.Lock()
			if len(d.conn) == 0 {
				d.mu.Unlock()
				continue
			}

			for i, conn := range d.conn {
				err := conn.enc.Encode(&codec.Message{Command: internal.PingCmd})
				if err != nil {
					fmt.Println("ERR", err)
					conn.conn.Close()
					if i == len(d.conn)-1 {
						d.conn = d.conn[:len(d.conn)-1]
						continue
					}
					copy(d.conn[i:], d.conn[i+1:])
					d.conn = d.conn[:len(d.conn)-1]
				}
			}
			d.mu.Unlock()
		}
	}
}

func (d *db) acquireConn(ctx context.Context) *client {
	d.mu.Lock()

	free := len(d.conn)

	if free > 0 {
		conn := d.conn[0]
		copy(d.conn, d.conn[1:])
		d.conn = d.conn[:free-1]
		conn.inUse = true
		d.active++
		d.mu.Unlock()

		return conn
	}

	d.awaitingConn++
	d.mu.Unlock()
	conn, ok := <-d.requiredConn

	if !ok {
		return nil
	}

	return conn
}

func (d *db) releaseConn(conn *client) {
	d.mu.Lock()

	if d.awaitingConn > 0 {
		d.requiredConn <- conn
		d.awaitingConn--
		d.mu.Unlock()
		return
	}

	conn.inUse = false
	d.active--
	d.conn = append(d.conn, conn)
	d.mu.Unlock()
}

// Authenticate authenticates for using storage
func (c *client) authenticate(login, pass string) error {
	err := c.enc.Encode(&codec.AuthMessage{Login: login, Password: pass})
	if err != nil {
		return fmt.Errorf("error while writing command: %s", err)
	}

	resp := codec.Message{}
	err = c.dec.Decode(&resp)
	if err != nil {
		return fmt.Errorf("error while reading response: %s", err)
	}

	if resp.Result == false {
		return errors.New(resp.Data)
	}

	return nil
}

// CreateBucket creates new bucket
func (d *db) CreateBucket(name string) error {
	if len(name) == 0 {
		return errors.New("invalid name")
	}

	c := d.acquireConn(context.Background())
	defer d.releaseConn(c)

	request := codec.Message{Command: internal.CreateBucketCmd, Value: name}

	err := c.enc.Encode(request)
	if err != nil {
		return fmt.Errorf("error while writing command: %s", err)
	}

	resp := codec.Message{}
	err = c.dec.Decode(&resp)
	if err != nil {
		return fmt.Errorf("error while reading response: %s", err)
	}

	return nil
}

// DeleteBucket deletes bucket
func (d *db) DeleteBucket(name string) error {
	if len(name) == 0 {
		return errors.New("invalid name")
	}

	c := d.acquireConn(context.Background())
	defer d.releaseConn(c)

	request := codec.Message{Command: internal.DelBucketCmd, Value: name}

	err := c.enc.Encode(request)
	if err != nil {
		return fmt.Errorf("error while writing command: %s", err)
	}

	resp := codec.Message{}
	err = c.dec.Decode(&resp)
	if err != nil {
		return fmt.Errorf("error while reading response: %s", err)
	}

	return nil
}

// Add new values to bucket
func (d *db) Add(bucket, key, val string) error {
	if len(bucket) == 0 {
		return errors.New("bucket name required")
	}

	if len(key) == 0 {
		return errors.New("key name required")
	}

	c := d.acquireConn(context.Background())
	if c == nil {
		return errors.New("failed to acquire conn")
	}
	defer d.releaseConn(c)

	msg := codec.Message{Command: internal.AddCmd, Bucket: bucket, Key: key, Value: val}

	err := c.enc.Encode(msg)
	if err != nil {
		return fmt.Errorf("error while writing command: %s", err)
	}

	err = c.dec.Decode(&msg)

	if err != nil {
		return fmt.Errorf("error while reading response: %s", err)
	}

	if msg.Result != true {
		return fmt.Errorf("add failed. %s", msg.Data)
	}

	return nil
}

// Get value from bucket
func (d *db) Get(bucket, key string) (string, error) {
	if len(bucket) == 0 || len(key) == 0 {
		return "", errors.New("invalid params")
	}

	c := d.acquireConn(context.Background())
	defer d.releaseConn(c)

	request := codec.Message{Command: internal.GetCmd, Bucket: bucket, Key: key}

	err := c.enc.Encode(request)
	if err != nil {
		return "", fmt.Errorf("error while writing command: %s", err)
	}

	resp := codec.Message{}
	err = c.dec.Decode(&resp)
	if err != nil {
		return "", fmt.Errorf("error while reading response: %s", err)
	}

	return resp.Value, nil
}

// Update value in bucket by key
func (d *db) Update(bucket, key, val string) error {
	panic("implement me")
}

// Delete value in bucket by key
func (d *db) Delete(bucket, key string) error {
	panic("implement me")
}

// Ping server
func (d *db) Ping() error {
	panic("implement me")
}

// Wait for all conn to finish
func (d *db) Wait() {
	fmt.Printf("\n Await %d Active %d\n", d.awaitingConn, d.active)

	for d.awaitingConn > 0 {
	}
	for d.active > 0 {
	}
}

// Close all conn
func (d *db) Close() error {
	request := codec.Message{Command: internal.CloseCmd}

	for _, c := range d.conn {
		err := c.enc.Encode(request)
		if err != nil {
			return fmt.Errorf("error while writing command: %s", err)
		}
	}

	return nil
}

func (d *db) isConnectionBroken(e error) bool {
	if e == nil {
		return false
	}
	if strings.Contains("broken pipe", e.Error()) {
		return true
	}

	return false
}
