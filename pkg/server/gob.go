package server

import (
	"encoding/gob"
	"github.com/serajam/sbucket/pkg"
	"github.com/serajam/sbucket/pkg/storage"
	"net"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const defaultAddress = ":3456"

// Option can be used to define parameters for creating new storage
type Option func(s *server)

// WithMiddleware options can be used to specify pre connection handlers
// which will be executed before starting handling incoming messages
func WithMiddleware(middleware ...Middleware) Option {
	return func(s *server) {
		if len(middleware) == 0 {
			return
		}

		s.middleware = middleware
	}
}

// Address for listening
func Address(address string) Option {
	return func(s *server) {
		if len(address) == 0 {
			return
		}

		s.address = address
	}
}

// Logger option for defining info and error output path
func Logger(loggerType, info, error string) Option {
	return func(s *server) {
		if len(info) == 0 || len(error) == 0 {
			return
		}

		infoPath, err := os.OpenFile(info, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0777)
		if err != nil {
			return
		}

		errorPath, err := os.OpenFile(error, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0777)
		if err != nil {
			return
		}

		s.logger = newLogger(loggerType, infoPath, errorPath)
	}
}

// Deadline option defines connection activity deadline
func Deadline(d int) Option {
	return func(s *server) {
		s.connectionDeadline = d
	}
}

// MaxConnNum option defines max conn number allowed
func MaxConnNum(d int) Option {
	return func(s *server) {
		s.maxConnNum = d
	}
}

// MaxFailures option defines max failures allowed per connection
func MaxFailures(d int) Option {
	return func(s *server) {
		s.maxFailures = d
	}
}

// ConnectTimeout option defines max wait time when trying to aquire slot for new connection
func ConnectTimeout(d int) Option {
	return func(s *server) {
		s.connectTimeout = d
	}
}

// message handler
type handler func(enc *gob.Encoder, m pkg.Message)

type server struct {
	storage storage.SBucketStorage
	logger  SBucketLogger

	address            string
	connectTimeout     int // seconds
	connectionDeadline int // seconds
	maxConnNum         int // seconds
	maxFailures        int

	handlers map[string]handler

	clientSem    chan struct{}
	clientsCount int32

	middleware []Middleware

	mu sync.RWMutex
	// contains server state
	running bool
}

// New creates new storage
func New(storage storage.SBucketStorage, options ...Option) SBucketServer {
	s := server{storage: storage}

	for _, o := range options {
		o(&s)
	}

	if len(s.address) == 0 {
		s.address = defaultAddress
	}

	if s.logger == nil {
		s.logger = newDefaultLogger(os.Stdout, os.Stderr)
	}

	if s.clientSem == nil {
		s.clientSem = make(chan struct{}, s.maxConnNum)
	}

	s.handlers = map[string]handler{
		pkg.CreateBucketCommand: s.handleCreateBucket,
		pkg.DeleteBucketCommand: s.handleDeleteBucket,
		pkg.AddCommand:          s.handleAdd,
		pkg.GetCommand:          s.handleGet,
		pkg.PingCommand:         s.handlePing,
	}

	return &s
}

func (s *server) writeMessage(e *gob.Encoder, msg *pkg.Message) {
	err := e.Encode(msg)
	if err != nil {
		s.logger.Errorf("Failed to write request: %s:\n", err)
		return
	}
}

func (s *server) isRunning() bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.running
}

// Shutdown change flag which is checked to verify if server is running
func (s *server) Shutdown() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.running = false
}

// Start starts storage server
func (s *server) Start() {
	s.running = true

	// TODO add http server for exposing stats
	var (
		err      error
		listener net.Listener
		conn     net.Conn
	)

	listener, err = net.Listen("tcp", s.address)
	if err != nil {
		s.logger.Errorf("Failed to start storage server: %s:", err)
		os.Exit(1)
	}

	defer func() {
		listener.Close()
	}()

	s.logger.Infof("Server listening on %s", s.address)

	for s.isRunning() {
		conn, err = listener.Accept()
		if err != nil {
			s.logger.Errorf("Failed to accept connection: %s:", err)
			continue
		}

		go s.handleConn(conn)
	}
}

func (s *server) handleConn(conn net.Conn) {
	defer conn.Close()

	dec := gob.NewDecoder(conn)
	enc := gob.NewEncoder(conn)

	if !s.acquireConnectionSlot(enc) {
		return
	}
	defer func() { <-s.clientSem }()

	atomic.AddInt32(&s.clientsCount, 1)
	defer func() { atomic.AddInt32(&s.clientsCount, -1) }()

	for _, m := range s.middleware {
		err := m.Run(enc, dec)
		if err != nil {
			s.logger.Error(err)
			return
		}
	}

	message := pkg.Message{}

	failures := s.maxFailures
	for {
		if failures > 10 {
			return
		}

		if !s.applyDeadline(conn, enc) {
			return
		}

		err := dec.Decode(&message)
		if err != nil {
			if err.Error() == "EOF" {
				return
			}
			if strings.Contains(err.Error(), "timeout") {
				s.writeMessage(enc, &pkg.Message{Result: false, Data: "Timeout"})
				return
			}

			failures++
			s.logger.Errorf("Failed to read command: %s\n", err)
			continue
		}

		if message.Command == pkg.CloseCommand {
			return
		}

		h, ok := s.handlers[message.Command]
		if !ok {
			s.writeMessage(enc, &pkg.Message{Result: false, Data: "Unknown command"})
			continue
		}

		h(enc, message)
	}
}

func (s *server) applyDeadline(c net.Conn, enc *gob.Encoder) bool {
	// unlimited
	if s.connectionDeadline == 0 {
		return true
	}

	err := c.SetDeadline(time.Now().Add(time.Duration(s.connectTimeout) * time.Second))
	if err != nil {
		s.logger.Error("Failed to set connection deadline")
		s.writeMessage(enc, &pkg.Message{Result: false, Data: "Internal error"})
		s.logger.Errorf("Failed to set connection deadline: %s:", err)
		return false
	}

	return true
}

func (s *server) acquireConnectionSlot(enc *gob.Encoder) bool {
	// unlimited
	if s.maxConnNum == 0 {
		return true
	}

	timeout := time.NewTimer(time.Duration(s.connectTimeout) * time.Second)
	select {
	case <-timeout.C:
		s.writeMessage(enc, &pkg.Message{Result: false, Data: "Connections limit reached"})
		return false
	case s.clientSem <- struct{}{}:
	}

	return true
}

func (s *server) handleCreateBucket(enc *gob.Encoder, m pkg.Message) {
	err := s.storage.NewBucket(m.Value)
	if err != nil {
		s.writeMessage(enc, &pkg.Message{Result: false, Data: err.Error()})
		return
	}
	s.writeMessage(enc, &pkg.Message{Result: true})
}

func (s *server) handleDeleteBucket(enc *gob.Encoder, m pkg.Message) {
	err := s.storage.DelBucket(m.Value)
	if err != nil {
		s.writeMessage(enc, &pkg.Message{Result: false, Data: err.Error()})
		return
	}
	s.writeMessage(enc, &pkg.Message{Result: true})
}

func (s *server) handleAdd(enc *gob.Encoder, m pkg.Message) {
	err := s.storage.Add(m.Bucket, m.Key, m.Value)
	if err != nil {
		go s.writeMessage(enc, &pkg.Message{Result: false, Data: err.Error()})
		return
	}
	go s.writeMessage(enc, &pkg.Message{Result: true})
}

func (s *server) handleGet(enc *gob.Encoder, m pkg.Message) {
	v, err := s.storage.Get(m.Bucket, m.Key)
	if err != nil {
		s.writeMessage(enc, &pkg.Message{Result: false, Data: err.Error()})
		return
	}
	s.writeMessage(enc, &pkg.Message{Value: v.Value()})
}

func (s *server) handlePing(enc *gob.Encoder, m pkg.Message) {
	s.logger.Debug("Received Ping")
}
