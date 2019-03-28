package server

import (
	"encoding/gob"
	"net"
	"os"
	"reflect"
	"testing"

	"github.com/serajam/sbucket/pkg"
	"github.com/serajam/sbucket/pkg/storage"
)

type ValMock struct{}

func (ValMock) Value() string {
	panic("implement me")
}

func (ValMock) Set(string) {
	panic("implement me")
}

type StorageMock struct{}

func (StorageMock) NewBucket(name string) error {
	panic("implement me")
}

func (StorageMock) DelBucket(name string) error {
	panic("implement me")
}

func (StorageMock) Add(bucket, key, val string) error {
	panic("implement me")
}

func (StorageMock) Get(bucket, key string) (storage.SBucketValue, error) {
	panic("implement me")
}

func (StorageMock) Del(bucket, key string) error {
	panic("implement me")
}

func (StorageMock) Update(bucket, key, val string) error {
	panic("implement me")
}

func (StorageMock) Stats() string {
	panic("implement me")
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
		{"should set deadline to 1", args{d: 1}, func(s *server) {},},
		{"should not set deadline", args{d: 0}, func(s *server) {},},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Deadline(tt.args.d)
			if reflect.TypeOf(got) != reflect.TypeOf(tt.want) {
				t.Errorf("Deadline() = %v, want %v", got, tt.want)
			}

			s := &server{}
			got(s)

			if s.connectionDeadline != tt.args.d {
				t.Errorf("Deadline() = %v, want %v", s.connectionDeadline, tt.args.d)
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
		{"should set max conn num to 1", args{d: 1}, func(s *server) {},},
		{"should not set max conn num", args{d: 0}, func(s *server) {},},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Deadline(tt.args.d)
			if reflect.TypeOf(got) != reflect.TypeOf(tt.want) {
				t.Errorf("MaxConnNum() = %v, want %v", got, tt.want)
			}

			s := &server{}
			got(s)

			if s.connectionDeadline != tt.args.d {
				t.Errorf("MaxConnNum() = %v, want %v", s.maxConnNum, tt.args.d)
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
		{"should set connect timeout to 1", args{d: 1}, func(s *server) {},},
		{"should not set connect timeout", args{d: 0}, func(s *server) {},},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Deadline(tt.args.d)
			if reflect.TypeOf(got) != reflect.TypeOf(tt.want) {
				t.Errorf("ConnectTimeout() = %v, want %v", got, tt.want)
			}

			s := &server{}
			got(s)

			if s.connectionDeadline != tt.args.d {
				t.Errorf("ConnectTimeout() = %v, want %v", s.maxConnNum, tt.args.d)
			}
		})
	}
}

func TestNew(t *testing.T) {
	s := &StorageMock{}
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
			&server{storage: s,},
		},
		{
			"should create new server with options",
			args{storage: s, options: []Option{Deadline(1), ConnectTimeout(1), MaxConnNum(0)}},
			&server{storage: s, connectionDeadline: 1, connectTimeout: 1, maxConnNum: 1},
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
		storage            storage.SBucketStorage
		logger             SBucketLogger
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
		e   *gob.Encoder
		msg *pkg.Message
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &server{
				storage:            tt.fields.storage,
				logger:             tt.fields.logger,
				address:            tt.fields.address,
				connectTimeout:     tt.fields.connectTimeout,
				connectionDeadline: tt.fields.connectionDeadline,
				maxConnNum:         tt.fields.maxConnNum,
				handlers:           tt.fields.handlers,
				clientSem:          tt.fields.clientSem,
				clientsCount:       tt.fields.clientsCount,
				middleware:         tt.fields.middleware,
			}
			s.writeMessage(tt.args.e, tt.args.msg)
		})
	}
}

func Test_server_Run(t *testing.T) {
	type fields struct {
		storage            storage.SBucketStorage
		logger             SBucketLogger
		address            string
		connectTimeout     int
		connectionDeadline int
		maxConnNum         int
		handlers           map[string]handler
		clientSem          chan struct{}
		clientsCount       int32
		middleware         []Middleware
	}
	tests := []struct {
		name   string
		fields fields
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &server{
				storage:            tt.fields.storage,
				logger:             tt.fields.logger,
				address:            tt.fields.address,
				connectTimeout:     tt.fields.connectTimeout,
				connectionDeadline: tt.fields.connectionDeadline,
				maxConnNum:         tt.fields.maxConnNum,
				handlers:           tt.fields.handlers,
				clientSem:          tt.fields.clientSem,
				clientsCount:       tt.fields.clientsCount,
				middleware:         tt.fields.middleware,
			}
			s.Run()
		})
	}
}

func Test_server_handleConn(t *testing.T) {
	type fields struct {
		storage            storage.SBucketStorage
		logger             SBucketLogger
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
		conn net.Conn
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &server{
				storage:            tt.fields.storage,
				logger:             tt.fields.logger,
				address:            tt.fields.address,
				connectTimeout:     tt.fields.connectTimeout,
				connectionDeadline: tt.fields.connectionDeadline,
				maxConnNum:         tt.fields.maxConnNum,
				handlers:           tt.fields.handlers,
				clientSem:          tt.fields.clientSem,
				clientsCount:       tt.fields.clientsCount,
				middleware:         tt.fields.middleware,
			}
			s.handleConn(tt.args.conn)
		})
	}
}

func Test_server_applyDeadline(t *testing.T) {
	type fields struct {
		storage            storage.SBucketStorage
		logger             SBucketLogger
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
		c   net.Conn
		enc *gob.Encoder
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &server{
				storage:            tt.fields.storage,
				logger:             tt.fields.logger,
				address:            tt.fields.address,
				connectTimeout:     tt.fields.connectTimeout,
				connectionDeadline: tt.fields.connectionDeadline,
				maxConnNum:         tt.fields.maxConnNum,
				handlers:           tt.fields.handlers,
				clientSem:          tt.fields.clientSem,
				clientsCount:       tt.fields.clientsCount,
				middleware:         tt.fields.middleware,
			}
			if got := s.applyDeadline(tt.args.c, tt.args.enc); got != tt.want {
				t.Errorf("server.applyDeadline() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_server_acquireConnectionSlot(t *testing.T) {
	type fields struct {
		storage            storage.SBucketStorage
		logger             SBucketLogger
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
		enc *gob.Encoder
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &server{
				storage:            tt.fields.storage,
				logger:             tt.fields.logger,
				address:            tt.fields.address,
				connectTimeout:     tt.fields.connectTimeout,
				connectionDeadline: tt.fields.connectionDeadline,
				maxConnNum:         tt.fields.maxConnNum,
				handlers:           tt.fields.handlers,
				clientSem:          tt.fields.clientSem,
				clientsCount:       tt.fields.clientsCount,
				middleware:         tt.fields.middleware,
			}
			if got := s.acquireConnectionSlot(tt.args.enc); got != tt.want {
				t.Errorf("server.acquireConnectionSlot() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_server_handleCreateBucket(t *testing.T) {
	type fields struct {
		storage            storage.SBucketStorage
		logger             SBucketLogger
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
		enc *gob.Encoder
		m   pkg.Message
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &server{
				storage:            tt.fields.storage,
				logger:             tt.fields.logger,
				address:            tt.fields.address,
				connectTimeout:     tt.fields.connectTimeout,
				connectionDeadline: tt.fields.connectionDeadline,
				maxConnNum:         tt.fields.maxConnNum,
				handlers:           tt.fields.handlers,
				clientSem:          tt.fields.clientSem,
				clientsCount:       tt.fields.clientsCount,
				middleware:         tt.fields.middleware,
			}
			s.handleCreateBucket(tt.args.enc, tt.args.m)
		})
	}
}

func Test_server_handleDeleteBucket(t *testing.T) {
	type fields struct {
		storage            storage.SBucketStorage
		logger             SBucketLogger
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
		enc *gob.Encoder
		m   pkg.Message
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &server{
				storage:            tt.fields.storage,
				logger:             tt.fields.logger,
				address:            tt.fields.address,
				connectTimeout:     tt.fields.connectTimeout,
				connectionDeadline: tt.fields.connectionDeadline,
				maxConnNum:         tt.fields.maxConnNum,
				handlers:           tt.fields.handlers,
				clientSem:          tt.fields.clientSem,
				clientsCount:       tt.fields.clientsCount,
				middleware:         tt.fields.middleware,
			}
			s.handleDeleteBucket(tt.args.enc, tt.args.m)
		})
	}
}

func Test_server_handleAdd(t *testing.T) {
	type fields struct {
		storage            storage.SBucketStorage
		logger             SBucketLogger
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
		enc *gob.Encoder
		m   pkg.Message
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &server{
				storage:            tt.fields.storage,
				logger:             tt.fields.logger,
				address:            tt.fields.address,
				connectTimeout:     tt.fields.connectTimeout,
				connectionDeadline: tt.fields.connectionDeadline,
				maxConnNum:         tt.fields.maxConnNum,
				handlers:           tt.fields.handlers,
				clientSem:          tt.fields.clientSem,
				clientsCount:       tt.fields.clientsCount,
				middleware:         tt.fields.middleware,
			}
			s.handleAdd(tt.args.enc, tt.args.m)
		})
	}
}

func Test_server_handleGet(t *testing.T) {
	type fields struct {
		storage            storage.SBucketStorage
		logger             SBucketLogger
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
		enc *gob.Encoder
		m   pkg.Message
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &server{
				storage:            tt.fields.storage,
				logger:             tt.fields.logger,
				address:            tt.fields.address,
				connectTimeout:     tt.fields.connectTimeout,
				connectionDeadline: tt.fields.connectionDeadline,
				maxConnNum:         tt.fields.maxConnNum,
				handlers:           tt.fields.handlers,
				clientSem:          tt.fields.clientSem,
				clientsCount:       tt.fields.clientsCount,
				middleware:         tt.fields.middleware,
			}
			s.handleGet(tt.args.enc, tt.args.m)
		})
	}
}

func Test_server_handlePing(t *testing.T) {
	type fields struct {
		storage            storage.SBucketStorage
		logger             SBucketLogger
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
		enc *gob.Encoder
		m   pkg.Message
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &server{
				storage:            tt.fields.storage,
				logger:             tt.fields.logger,
				address:            tt.fields.address,
				connectTimeout:     tt.fields.connectTimeout,
				connectionDeadline: tt.fields.connectionDeadline,
				maxConnNum:         tt.fields.maxConnNum,
				handlers:           tt.fields.handlers,
				clientSem:          tt.fields.clientSem,
				clientsCount:       tt.fields.clientsCount,
				middleware:         tt.fields.middleware,
			}
			s.handlePing(tt.args.enc, tt.args.m)
		})
	}
}
