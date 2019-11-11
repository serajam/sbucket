package server

import (
	"bytes"
	"encoding/gob"
	"errors"
	"io"
	"io/ioutil"
	"net"
	"os"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/serajam/sbucket/internal"
	"github.com/serajam/sbucket/internal/codec"
	"github.com/serajam/sbucket/internal/storage"
)

type mockValue struct{}

func (mockValue) Value() string {
	return "test"
}

func (mockValue) Set(string) {

}

type mockMiddleware struct {
}

func (mockMiddleware) Run(enc codec.Codec) error {
	return errors.New("error")
}

type storageMockOk struct{}

func (storageMockOk) NewBucket(name string) error {
	return nil
}

func (storageMockOk) DelBucket(name string) error {
	return nil
}

func (storageMockOk) Add(bucket, key, val string) error {
	return nil
}

func (storageMockOk) Get(bucket, key string) (storage.SBucketValue, error) {
	return &mockValue{}, nil
}

func (storageMockOk) Del(bucket, key string) error {
	return nil
}

func (storageMockOk) Update(bucket, key, val string) error {
	return nil
}

func (storageMockOk) Stats() string {
	return "ok"
}

type storageMockError struct{}

func (storageMockError) NewBucket(name string) error {
	return errors.New("error")
}

func (storageMockError) DelBucket(name string) error {
	return errors.New("error")
}

func (storageMockError) Add(bucket, key, val string) error {
	return errors.New("error")
}

func (storageMockError) Get(bucket, key string) (storage.SBucketValue, error) {
	return nil, errors.New("error")
}

func (storageMockError) Del(bucket, key string) error {
	return errors.New("error")
}

func (storageMockError) Update(bucket, key, val string) error {
	return errors.New("error")
}

func (storageMockError) Stats() string {
	return ""
}

func TestWithMiddleware(t *testing.T) {
	type args struct {
		middleware []Middleware
	}
	tests := []struct {
		name string
		args args
		want Option
	}{
		{
			"should return option and set middleware len 1",
			args{middleware: []Middleware{AuthMiddleware{}}},
			func(s *server) {},
		},
		{
			"should return option and do not set middleware",
			args{middleware: []Middleware{}},
			func(s *server) {},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := WithMiddleware(tt.args.middleware...)
			if reflect.TypeOf(got) != reflect.TypeOf(tt.want) {
				t.Errorf("WithMiddleware() = %v, want %v", got, tt.want)
			}
			s := &server{}
			got(s)

			if len(s.middleware) != len(tt.args.middleware) {
				t.Errorf("WithMiddleware() = %v, want %v", len(s.middleware), len(tt.args.middleware))
			}
		})
	}
}

func TestAddress(t *testing.T) {
	type args struct {
		address string
	}
	tests := []struct {
		name string
		args args
		want Option
	}{
		{
			"should set address",
			args{address: "address"},
			func(s *server) {},
		},
		{
			"should not set address",
			args{address: ""},
			func(s *server) {},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Address(tt.args.address)
			if reflect.TypeOf(got) != reflect.TypeOf(tt.want) {
				t.Errorf("Address() = %v, want %v", got, tt.want)
			}

			s := &server{}
			got(s)

			if len(s.address) != len(tt.args.address) {
				t.Errorf("Address() = %v, want %v", len(s.middleware), len(tt.args.address))
			}
		})
	}
}

func TestLogger(t *testing.T) {
	type args struct {
		loggerType string
		info       string
		error      string
	}
	tests := []struct {
		name string
		args args
		want Option
	}{
		{
			"should set default logger",
			args{"", os.DevNull, os.DevNull},
			func(s *server) {},
		},
		{
			"should set logrus logger",
			args{logrusType, os.DevNull, os.DevNull},
			func(s *server) {},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Logger(tt.args.loggerType, tt.args.info, tt.args.error)
			if reflect.TypeOf(got) != reflect.TypeOf(tt.want) {
				t.Errorf("Logger() = %v, want %v", got, tt.want)
			}

			s := &server{}
			got(s)

			logger := newLogger(tt.args.loggerType, os.Stdout, os.Stdout)
			if reflect.TypeOf(s.logger) != reflect.TypeOf(logger) {
				t.Errorf("Logger() = %v, want %v", s.logger, logger)
			}
		})
	}
}

func TestDeadline(t *testing.T) {
	type args struct {
		d int
	}
	tests := []struct {
		name string
		args args
		want Option
	}{
		{"should set deadline to 1", args{d: 1}, func(s *server) {}},
		{"should not set deadline", args{d: 0}, func(s *server) {}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ConnDeadline(tt.args.d)
			if reflect.TypeOf(got) != reflect.TypeOf(tt.want) {
				t.Errorf("ConnDeadline() = %v, want %v", got, tt.want)
			}

			s := &server{}
			got(s)

			if s.connectionDeadline != tt.args.d {
				t.Errorf("ConnDeadline() = %v, want %v", s.connectionDeadline, tt.args.d)
			}
		})
	}
}

func TestMaxConnNum(t *testing.T) {
	type args struct {
		d int
	}
	tests := []struct {
		name string
		args args
		want Option
	}{
		{"should set max clients num to 1", args{d: 1}, func(s *server) {}},
		{"should not set max clients num", args{d: 0}, func(s *server) {}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ConnDeadline(tt.args.d)
			if reflect.TypeOf(got) != reflect.TypeOf(tt.want) {
				t.Errorf("MaxConnNum() = %v, want %v", got, tt.want)
			}

			s := &server{}
			got(s)

			if s.connectionDeadline != tt.args.d {
				t.Errorf("MaxConnNum() = %v, want %v", s.maxClientsNum, tt.args.d)
			}
		})
	}
}

func TestMaxFailures(t *testing.T) {
	type args struct {
		d int
	}
	tests := []struct {
		name string
		args args
		want Option
	}{
		{"should set max failures  to 1", args{d: 1}, func(s *server) {}},
		{"should not set max failures", args{d: 0}, func(s *server) {}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MaxFailures(tt.args.d)
			if reflect.TypeOf(got) != reflect.TypeOf(tt.want) {
				t.Errorf("MaxFailures() = %v, want %v", got, tt.want)
			}

			s := &server{}
			got(s)

			if s.maxFailures != tt.args.d {
				t.Errorf("MaxFailures() = %v, want %v", s.maxClientsNum, tt.args.d)
			}
		})
	}
}

func TestConnectTimeout(t *testing.T) {
	type args struct {
		d int
	}
	tests := []struct {
		name string
		args args
		want Option
	}{
		{"should set connect timeout to 1", args{d: 1}, func(s *server) {}},
		{"should not set connect timeout", args{d: 0}, func(s *server) {}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ConnDeadline(tt.args.d)
			if reflect.TypeOf(got) != reflect.TypeOf(tt.want) {
				t.Errorf("ConnectTimeout() = %v, want %v", got, tt.want)
			}

			s := &server{}
			got(s)

			if s.connectionDeadline != tt.args.d {
				t.Errorf("ConnectTimeout() = %v, want %v", s.maxClientsNum, tt.args.d)
			}
		})
	}
}

func TestNew(t *testing.T) {
	s := &storageMockOk{}
	type args struct {
		storage storage.SBucketStorage
		options []Option
	}
	tests := []struct {
		name string
		args args
		want SBucketServer
	}{
		{
			"should create new server",
			args{storage: s, options: nil},
			&server{handler: &actionsHandler{storage: s}},
		},
		{
			"should create new server with options",
			args{storage: s, options: []Option{ConnDeadline(1), ConnectTimeout(1), MaxConnNum(0)}},
			&server{handler: &actionsHandler{storage: s}, connectionDeadline: 1, connectTimeout: 1, maxClientsNum: 1},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := New(tt.args.storage, tt.args.options...)

			if reflect.TypeOf(got) != reflect.TypeOf(got) {
				t.Errorf("New() = %+v, want %+v", reflect.TypeOf(got), reflect.TypeOf(tt.want))
			}
		})
	}
}

func Test_server_writeMessage(t *testing.T) {
	type fields struct {
		logger SBucketLogger
	}
	type args struct {
		e   *gob.Encoder
		msg *codec.Message
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			"should write message",
			fields{
				nil,
			},
			args{e: gob.NewEncoder(ioutil.Discard), msg: &codec.Message{}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &server{
				logger:  tt.fields.logger,
				handler: &actionsHandler{},
			}
			s.handler.writeMessage(tt.args.e, tt.args.msg)
		})
	}
}

func Test_server_Start(t *testing.T) {
	type fields struct {
		logger  SBucketLogger
		address string
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{
			"should run server",
			fields{
				newDefaultLogger(ioutil.Discard, ioutil.Discard),
				":54444",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &server{
				handler: &actionsHandler{},
				logger:  newDefaultLogger(os.Stdout, os.Stderr),
				address: tt.fields.address,
				clients: map[string]*wrappedConn{},
			}
			wg := &sync.WaitGroup{}
			wg.Add(1)
			go func(wg *sync.WaitGroup) {
				s.Start()
				wg.Done()
			}(wg)
			time.Sleep(1 * time.Second)

			c, err := net.Dial("tcp", ":54444")
			if err != nil {
				t.Error(err)
				return
			}

			s.clients["test"] = &wrappedConn{Conn: c}
			c.Close()

			s.Shutdown()
			wg.Wait()
		})
	}
}

func Test_server_handleConn(t *testing.T) {
	type fields struct {
		storage            storage.SBucketStorage
		errOutput          io.Writer
		address            string
		connectTimeout     int
		connectionDeadline int
		maxConnNum         int
		handlers           map[string]handler
		clientSem          chan struct{}
		clientsCount       int32
		middleware         []Middleware
	}
	type args struct {
		callback func(c net.Conn)
	}
	tests := []struct {
		name            string
		fields          fields
		args            args
		wantLoggerError bool
	}{
		{
			"should handle connection EOF",
			fields{
				&storageMockOk{},
				os.Stderr,
				":54444",
				10,
				10,
				1,
				make(map[string]handler),
				make(chan struct{}, 1),
				0,
				[]Middleware{},
			},
			args{callback: func(c net.Conn) {
				time.Sleep(1000 * time.Millisecond)
				c.Close()
			}},
			false,
		},

		{
			"should fail on read deadline",
			fields{
				&storageMockOk{},
				os.Stderr,
				":54444",
				1,
				1,
				1,
				make(map[string]handler),
				make(chan struct{}, 1),
				0,
				[]Middleware{},
			},
			args{callback: func(c net.Conn) {

			}},
			false,
		},

		{
			"should timeout if no connection slot available",
			fields{
				&storageMockOk{},
				os.Stderr,
				":54444",
				1,
				1,
				1,
				make(map[string]handler),
				make(chan struct{}, 0),
				0,
				[]Middleware{},
			},
			args{callback: func(c net.Conn) {
			}},
			false,
		},

		{
			"should run middleware error",
			fields{
				&storageMockOk{},
				os.Stderr,
				":54444",
				1,
				1,
				1,
				make(map[string]handler),
				make(chan struct{}, 1),
				0,
				[]Middleware{mockMiddleware{}},
			},
			args{callback: func(c net.Conn) {
				time.Sleep(100 * time.Millisecond)
				c.Close()
			}},
			true,
		},

		{
			"should handle invalid command",
			fields{
				&storageMockOk{},
				os.Stderr,
				":54444",
				1,
				1,
				1,
				make(map[string]handler),
				make(chan struct{}, 1),
				0,
				[]Middleware{},
			},
			args{callback: func(c net.Conn) {
				time.Sleep(100 * time.Millisecond)
				enc := gob.NewEncoder(c)
				err := enc.Encode(&codec.Message{Command: "INVALID"})
				if err != nil {
					println(err)
				}
				time.Sleep(100 * time.Millisecond)
				c.Close()
			}},
			false,
		},

		{
			"should handle command",
			fields{
				&storageMockOk{},
				os.Stderr,
				":54444",
				1,
				1,
				1,
				map[string]handler{"TEST": func(c codec.Encoder, m *codec.Message) {}},
				make(chan struct{}, 1),
				0,
				[]Middleware{},
			},
			args{callback: func(c net.Conn) {
				time.Sleep(500 * time.Millisecond)
				enc := gob.NewEncoder(c)
				err := enc.Encode(&codec.Message{Command: "TEST"})
				if err != nil {
					println(err)
				}
				time.Sleep(100 * time.Millisecond)
				c.Close()
			}},
			false,
		},

		{
			"should handle close command",
			fields{
				&storageMockOk{},
				os.Stderr,
				":54444",
				1,
				1,
				1,
				map[string]handler{},
				make(chan struct{}, 1),
				0,
				[]Middleware{},
			},
			args{callback: func(c net.Conn) {
				time.Sleep(500 * time.Millisecond)
				enc := gob.NewEncoder(c)
				err := enc.Encode(&codec.Message{Command: internal.CloseCommand})
				if err != nil {
					println(err)
				}
				time.Sleep(100 * time.Millisecond)
				c.Close()
			}},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buffer := &bytes.Buffer{}

			s := &server{
				handler:            &actionsHandler{storage: tt.fields.storage},
				logger:             newDefaultLogger(ioutil.Discard, buffer),
				address:            tt.fields.address,
				codecType:          codec.Gob,
				connectTimeout:     tt.fields.connectTimeout,
				connectionDeadline: tt.fields.connectionDeadline,
				maxClientsNum:      tt.fields.maxConnNum,
				handlers:           tt.fields.handlers,
				clientSem:          tt.fields.clientSem,
				clientsCount:       tt.fields.clientsCount,
				middleware:         tt.fields.middleware,
			}

			p, _ := net.Pipe()
			c := &wrappedConn{Conn: p}
			go tt.args.callback(c)
			go s.handleConn(c)
			time.Sleep(500 * time.Millisecond)

			logs, _ := ioutil.ReadAll(buffer)
			if !tt.wantLoggerError && len(logs) > 0 {
				t.Errorf("Was not expecting errors in log: %s", logs)
				return
			}
		})
	}
}

func Test_server_applyDeadline(t *testing.T) {
	type fields struct {
		connectionDeadline int
	}
	type args struct {
		callback func(c net.Conn)
	}
	tests := []struct {
		name   string
		args   args
		fields fields
		want   bool
	}{
		{
			"should apply clients deadline",
			args{callback: func(c net.Conn) {
				time.Sleep(500 * time.Millisecond)
				c.Close()
			}},
			fields{1},
			true,
		},
		{
			"should not apply clients deadline",
			args{callback: func(c net.Conn) {
				c.Close()
			}},
			fields{1},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &server{
				handler:            &actionsHandler{logger: newDefaultLogger(ioutil.Discard, ioutil.Discard)},
				logger:             newDefaultLogger(ioutil.Discard, ioutil.Discard),
				connectionDeadline: tt.fields.connectionDeadline,
			}

			c, _ := net.Pipe()
			go tt.args.callback(c)
			cod, _ := codec.New(codec.Gob, c)
			time.Sleep(200 * time.Millisecond)
			if got := s.applyDeadline(c, cod); got != tt.want {
				t.Errorf("server.applyDeadline() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_server_acquireConnectionSlot(t *testing.T) {
	type fields struct {
		connectTimeout int
		maxConnNum     int
		clientSem      chan struct{}
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			"should acquire connection slot",
			fields{
				1,
				1,
				make(chan struct{}, 1),
			},
			true,
		},

		{
			"should not acquire connection slot",
			fields{
				1,
				1,
				make(chan struct{}),
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &server{
				handler:        &actionsHandler{logger: newDefaultLogger(ioutil.Discard, ioutil.Discard)},
				logger:         newDefaultLogger(ioutil.Discard, ioutil.Discard),
				connectTimeout: tt.fields.connectTimeout,
				maxClientsNum:  tt.fields.maxConnNum,
				clientSem:      tt.fields.clientSem,
			}

			c, _ := net.Pipe()
			cod, _ := codec.New(codec.Gob, c)
			c.Close()
			if got := s.acquireConnectionSlot(cod); got != tt.want {
				t.Errorf("server.acquireConnectionSlot() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_server_handleCreateBucket(t *testing.T) {
	type fields struct {
		storage storage.SBucketStorage
	}
	type args struct {
		m *codec.Message
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			"should create new bucket",
			fields{&storageMockOk{}},
			args{m: &codec.Message{Command: internal.CreateBucketCommand, Value: "NEW"}},
			true,
		},

		{
			"should not create new bucket",
			fields{&storageMockError{}},
			args{m: &codec.Message{Command: internal.CreateBucketCommand, Value: ""}},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &server{handler: &actionsHandler{storage: tt.fields.storage}}

			c, r := net.Pipe()
			cod, _ := codec.New(codec.Gob, c)
			codr, _ := codec.New(codec.Gob, r)
			go s.handler.handleCreateBucket(cod, tt.args.m)
			time.Sleep(200 * time.Millisecond)
			got := &codec.Message{}
			codr.Decode(got)
			c.Close()
			r.Close()

			if got.Result != tt.want {
				t.Errorf("server.handleCreateBucket() = %v, want %v", got.Result, tt.want)
			}
		})
	}
}

func Test_server_handleDeleteBucket(t *testing.T) {
	type fields struct {
		storage storage.SBucketStorage
	}
	type args struct {
		m *codec.Message
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			"should delete bucket",
			fields{&storageMockOk{}},
			args{m: &codec.Message{Command: internal.DeleteBucketCommand, Value: "NEW"}},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &server{handler: &actionsHandler{storage: tt.fields.storage}}

			c, r := net.Pipe()
			cod, _ := codec.New(codec.Gob, c)
			codr, _ := codec.New(codec.Gob, r)
			go s.handler.handleDeleteBucket(cod, tt.args.m)
			time.Sleep(200 * time.Millisecond)
			got := &codec.Message{}
			codr.Decode(got)
			c.Close()
			r.Close()

			if got.Result != tt.want {
				t.Errorf("server.handleCreateBucket() = %v, want %v", got.Result, tt.want)
			}
		})
	}
}

func Test_server_handleAdd(t *testing.T) {
	type fields struct {
		storage storage.SBucketStorage
	}
	type args struct {
		m *codec.Message
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			"should create key",
			fields{&storageMockOk{}},
			args{m: &codec.Message{Command: internal.AddCommand, Key: "TEST", Value: "NEW"}},
			true,
		},

		{
			"should not create new key",
			fields{&storageMockError{}},
			args{m: &codec.Message{Command: internal.AddCommand, Key: "TEST", Value: ""}},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &server{handler: &actionsHandler{storage: tt.fields.storage}}

			c, r := net.Pipe()
			cod, _ := codec.New(codec.Gob, c)
			codr, _ := codec.New(codec.Gob, r)
			go s.handler.handleAdd(cod, tt.args.m)
			time.Sleep(200 * time.Millisecond)
			got := &codec.Message{}
			codr.Decode(got)
			c.Close()
			r.Close()

			if got.Result != tt.want {
				t.Errorf("server.handleAdd() = %v, want %v", got.Result, tt.want)
			}
		})
	}
}

func Test_server_handleGet(t *testing.T) {
	type fields struct {
		storage storage.SBucketStorage
	}
	type args struct {
		m *codec.Message
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			"should get key",
			fields{&storageMockOk{}},
			args{m: &codec.Message{Command: internal.GetCommand, Key: "TEST"}},
			true,
		},

		{
			"should not get key",
			fields{&storageMockError{}},
			args{m: &codec.Message{Command: internal.GetCommand, Key: ""}},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &server{handler: &actionsHandler{storage: tt.fields.storage}}

			c, r := net.Pipe()
			cod, _ := codec.New(codec.Gob, c)
			codr, _ := codec.New(codec.Gob, r)
			go s.handler.handleGet(cod, tt.args.m)
			time.Sleep(200 * time.Millisecond)
			got := &codec.Message{}
			codr.Decode(got)
			c.Close()
			r.Close()

			if got.Result != tt.want {
				t.Errorf("server.handleGet() = %v, want %v", got.Result, tt.want)
			}
		})
	}
}

func Test_server_isRunning(t *testing.T) {
	type fields struct {
		running bool
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			"should return running",
			fields{running: true},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &server{
				running: tt.fields.running,
			}
			if got := s.isRunning(); got != tt.want {
				t.Errorf("server.isRunning() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_server_Shutdown(t *testing.T) {
	type fields struct {
		running bool
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{
			"should set running false",
			fields{running: true},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l, _ := net.Listen("tcp", ":54444")
			sto, _ := storage.New(storage.MapType)
			s := &server{
				running:      tt.fields.running,
				logger:       newDefaultLogger(ioutil.Discard, ioutil.Discard),
				connListener: l,
				handler:      &actionsHandler{storage: sto, logger: newDefaultLogger(ioutil.Discard, ioutil.Discard)},
			}
			go func() {

			}()
			s.Shutdown()

			if s.running != false {
				t.Error("server.isRunning() should set running to false")
			}
		})
	}
}
