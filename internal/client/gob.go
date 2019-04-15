package client

import (
	"context"
	"encoding/gob"
	"errors"
	"fmt"
	"github.com/serajam/sbucket/internal"
	"net"
	"strings"
	"sync"
	"time"
)

type gobClient struct {
	conn  net.Conn
	inUse bool

	enc *gob.Encoder
	dec *gob.Decoder
}

type db struct {
	mu sync.RWMutex

	address     string
	dialTimeout time.Duration

	conn   []*gobClient
	active int

	awaitingConn int
	requiredConn chan *gobClient

	pingInterval int
}

// NewGobClient creates 	a pool of connections
func NewGobClient(address string, connNum int, dialTimeout time.Duration, login, pass string) (Client, error) {
	dbConn := &db{
		conn:         make([]*gobClient, 0),
		requiredConn: make(chan *gobClient, connNum),
		address:      address,
		dialTimeout:  dialTimeout,
		pingInterval: 0,
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

	go dbConn.ping()

	return dbConn, nil
}

func (d db) newConn() (*gobClient, error) {
	c, err := net.DialTimeout("tcp", d.address, d.dialTimeout*time.Second)
	if err != nil {
		return nil, fmt.Errorf("failed to open storage connection: %s", err)
	}

	return &gobClient{conn: c, enc: gob.NewEncoder(c), dec: gob.NewDecoder(c)}, nil
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
				err := conn.enc.Encode(&internal.Message{Command: internal.PingCommand})
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

func (d *db) acquireConn(ctx context.Context) *gobClient {
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

func (d *db) releaseConn(conn *gobClient) {
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
func (c *gobClient) authenticate(login, pass string) error {
	err := c.enc.Encode(&internal.AuthMessage{Login: login, Password: pass})
	if err != nil {
		return fmt.Errorf("error while writing command: %s", err)
	}

	resp := internal.Message{}
	err = c.dec.Decode(&resp)
	if err != nil {
		return fmt.Errorf("error while reading response: %s", err)
	}

	if resp.Result == false {
		return errors.New(resp.Data)
	}

	return nil
}

func (d *db) CreateBucket(name string) error {
	if len(name) == 0 {
		return errors.New("invalid name")
	}

	c := d.acquireConn(context.Background())
	defer d.releaseConn(c)

	request := internal.Message{Command: internal.CreateBucketCommand, Value: name}

	err := c.enc.Encode(request)
	if err != nil {
		return fmt.Errorf("error while writing command: %s", err)
	}

	resp := internal.Message{}
	err = c.dec.Decode(&resp)
	if err != nil {
		return fmt.Errorf("error while reading response: %s", err)
	}

	return nil
}

func (d *db) DeleteBucket(name string) error {
	if len(name) == 0 {
		return errors.New("invalid name")
	}

	c := d.acquireConn(context.Background())
	defer d.releaseConn(c)

	request := internal.Message{Command: internal.DeleteBucketCommand, Value: name}

	err := c.enc.Encode(request)
	if err != nil {
		return fmt.Errorf("error while writing command: %s", err)
	}

	resp := internal.Message{}
	err = c.dec.Decode(&resp)
	if err != nil {
		return fmt.Errorf("error while reading response: %s", err)
	}

	return nil
}

func (d *db) Add(bucket, key, val string) error {
	if len(bucket) == 0 || len(key) == 0 || len(val) == 0 {
		return errors.New("invalid params")
	}

	c := d.acquireConn(context.Background())
	if c == nil {
		return errors.New("failed to acquire conn")
	}
	defer d.releaseConn(c)

	msg := internal.Message{Command: internal.AddCommand, Bucket: bucket, Key: key, Value: val}

	err := c.enc.Encode(msg)
	if err != nil {
		return fmt.Errorf("error while writing command: %s", err)
	}

	err = c.dec.Decode(&msg)

	if d.isCOnnectionBroken(err) {
		c.conn.Close()
	}

	if err != nil {
		return fmt.Errorf("error while reading response: %s", err)
	}

	if msg.Result != true {
		return fmt.Errorf("add failed. %s", msg.Data)
	}

	return nil
}
func (d *db) Get(bucket, key string) (string, error) {
	if len(bucket) == 0 || len(key) == 0 {
		return "", errors.New("invalid params")
	}

	c := d.acquireConn(context.Background())
	defer d.releaseConn(c)

	request := internal.Message{Command: internal.GetCommand, Bucket: bucket, Key: key}

	err := c.enc.Encode(request)
	if err != nil {
		return "", fmt.Errorf("error while writing command: %s", err)
	}

	resp := internal.Message{}
	err = c.dec.Decode(&resp)
	if err != nil {
		return "", fmt.Errorf("error while reading response: %s", err)
	}

	return resp.Value, nil
}

func (d *db) Update(bucket, key, val string) error {
	panic("implement me")
}

func (d *db) Delete(bucket, key string) error {
	panic("implement me")
}

func (d *db) Ping() error {
	panic("implement me")
}

func (d *db) Wait() {

	fmt.Printf("\n Await %d Active %d", d.awaitingConn, d.active)

	for d.awaitingConn > 0 {
	}
	for d.active > 0 {
	}
}

func (d *db) Close() error {
	for d.awaitingConn > 0 {
	}
	for d.active > 0 {
	}

	request := internal.Message{Command: internal.CloseCommand}

	for _, c := range d.conn {
		err := c.enc.Encode(request)
		if err != nil {
			return fmt.Errorf("error while writing command: %s", err)
		}
	}

	return nil
}

func (d *db) isCOnnectionBroken(e error) bool {
	if e == nil {
		return false
	}
	if strings.Contains("broken pipe", e.Error()) {
		return true
	}

	return false
}
