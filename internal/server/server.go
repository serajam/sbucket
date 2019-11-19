package server

import (
	"errors"
	"fmt"
	"github.com/serajam/sbucket/internal"
	"github.com/serajam/sbucket/internal/codec"
	"github.com/serajam/sbucket/internal/storage"
	"math/rand"
	"net"
	"os"
	"os/signal"
	"strconv"
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

// ConnDeadline option defines connection activity deadline
// 0 - disables deadline
func ConnDeadline(d int) Option {
	return func(s *server) {
		s.connectionDeadline = d
	}
}

// MaxConnNum option defines max clients number allowed
// 0 - disables connection limit
func MaxConnNum(d int) Option {
	return func(s *server) {
		s.maxClientsNum = d
	}
}

// MaxFailures option defines max failures allowed per connection
// 0 - disables failures limit
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
type handler func(enc codec.Encoder, m *codec.Message)

type server struct {
	handler      *actionsHandler
	logger       SBucketLogger
	connListener net.Listener

	address            string
	codecType          int // codec which will be used to encode and decode messages
	connectTimeout     int // seconds
	connectionDeadline int // seconds
	maxClientsNum      int // seconds
	maxFailures        int

	handlers map[int]handler

	clientSem    chan struct{}
	clientsCount int32

	middleware []Middleware

	mu sync.RWMutex
	// contains server state
	running bool
	clients map[string]*wrappedConn
}

// New creates new storage
func New(storage storage.SBucketStorage, options ...Option) SBucketServer {
	s := server{clients: map[string]*wrappedConn{}}

	for _, o := range options {
		o(&s)
	}

	if len(s.address) == 0 {
		s.address = defaultAddress
	}

	if s.logger == nil {
		s.logger = newDefaultLogger(os.Stdout, os.Stderr)
	}

	if s.maxClientsNum > 0 {
		s.clientSem = make(chan struct{}, s.maxClientsNum)
	}

	if s.codecType == 0 {
		s.codecType = 2
	}

	s.handler = &actionsHandler{storage: storage, logger: s.logger}

	s.handlers = map[int]handler{
		internal.CreateBucketCmd: s.handler.handleCreateBucket,
		internal.DelBucketCmd:    s.handler.handleDeleteBucket,
		internal.AddCmd:          s.handler.handleAdd,
		internal.GetCmd:          s.handler.handleGet,
		internal.PingCmd:         s.handler.handlePing,
	}

	return &s
}

// Start starts storage server
func (s *server) Start() {
	// TODO add http server for exposing stats
	var (
		err error
		c   net.Conn
	)

	s.connListener, err = net.Listen("tcp", s.address)
	if err != nil {
		s.logger.Errorf("Failed to start storage server: %s:", err)
		os.Exit(1)
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs)
	go func() {
		defer signal.Stop(sigs)
		si := <-sigs
		fmt.Println("Received:", si)
		s.Shutdown()
	}()

	s.running = true
	s.logger.Infof("Server listening on %s", s.address)

	for s.isRunning() {
		c, err = s.connListener.Accept()

		if !s.isRunning() {
			break
		}

		if err != nil {
			s.logger.Errorf("Failed to accept connection: %s:", err)
			continue
		}

		conn := &wrappedConn{Conn: c, active: false}
		s.logger.Debugf("Accepted new connection: %s", conn.LocalAddr())
		go s.handleConn(conn)
	}
}

func (s *server) handleConn(conn *wrappedConn) {
	key, err := s.addClientConn(conn)
	if err != nil {
		s.logger.Info(err)
		return
	}

	defer func(key string, conn *wrappedConn) {
		if conn.isPendingClosure() {
			return
		}
		s.logger.Debug("Conn close: " + conn.RemoteAddr().String())
		err = conn.Close()
		if err != nil {
			s.logger.Debug(err)
		}
		s.removeClientConn(key)
	}(key, conn)

	connCodec, err := codec.New(s.codecType, conn)
	if err != nil {
		s.logger.Error(err)
		return
	}

	if !s.acquireConnectionSlot(connCodec) {
		return
	}
	defer func() { <-s.clientSem }()

	atomic.AddInt32(&s.clientsCount, 1)
	defer func() { atomic.AddInt32(&s.clientsCount, -1) }()

	for _, m := range s.middleware {
		err = m.Run(connCodec)
		if err != nil {
			s.logger.Error(err)
			return
		}
	}

	message := codec.Message{}
	failures := s.maxFailures

	conn.setActive(true)
	defer conn.setActive(false)

	for s.isRunning() && !conn.isPendingClosure() {
		if s.maxFailures > 0 && failures == 0 {
			return
		}

		if !s.applyDeadline(conn, connCodec) {
			return
		}

		err = connCodec.Decode(&message)
		if err != nil {

			if strings.Contains(err.Error(), "EOF") {
				s.logger.Debug(err)
				return
			}
			if strings.Contains(err.Error(), "timeout") {
				s.logger.Debug(err)
				return
			}

			if strings.Contains(err.Error(), "closed pipe") {
				s.logger.Debug(err)
				return
			}

			if strings.Contains(err.Error(), "use of closed network connection") {
				s.logger.Debug(err)
				return
			}

			failures--
			s.logger.Errorf("Failed to read command: %s\n", err)
			continue
		}

		if message.Command == internal.CloseCmd {
			return
		}

		h, ok := s.handlers[int(message.Command)]
		if !ok {
			s.handler.writeMessage(connCodec, &codec.Message{Result: false, Data: "Unknown command"})
			continue
		}

		h(connCodec, &message)
	}
}

func (s *server) addClientConn(c *wrappedConn) (string, error) {
	if !s.isRunning() {
		return "", errors.New("server is shutting down")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	key := strconv.Itoa(rand.Int())
	s.clients[key] = c
	return key, nil
}

func (s *server) removeClientConn(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.clients, key)
}

func (s *server) isRunning() bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.running
}

// Shutdown change flag which is checked to verify if server is running
func (s *server) Shutdown() {
	s.logger.Debug("Shutdown started")
	if s.running == false {
		s.logger.Debug("Already stopped")
		return
	}
	s.running = false

	err := s.connListener.Close()
	if err != nil {
		s.logger.Error(err)
	}

	if len(s.clients) > 0 {
		s.logger.Debug("All connections closed")
		return
	}

	s.mu.Lock()
	toClose := make([]*wrappedConn, 0)
	for _, c := range s.clients {
		c.pendingClosure = true
		connCodec, _ := codec.New(s.codecType, c)
		s.handler.writeMessage(connCodec, &codec.Message{Command: internal.CloseCmd})
		toClose = append(toClose, c)
	}

	ticker := time.NewTicker(100 * time.Millisecond)
	timeout := time.NewTimer(10 * time.Second)

	select {
	case <-ticker.C:
		if len(toClose) == 0 {
			break
		}
		i := len(toClose) - 1
		for i > 0 {
			c := toClose[i]
			toClose = append(toClose[:i], toClose[i+1:]...)
			if c.isActive() {
				i--
				continue
			}

			c.Close()
			i--
		}
	case <-timeout.C:
		s.logger.Error("Connections close timeout. Forcing closure")
		for _, c := range toClose {
			c.Close()
		}
	}

	s.mu.Unlock()
	s.logger.Debug("Shutdown ended")
	s.logger.Debug("Stats: " + s.handler.storage.Stats())
}

func (s *server) applyDeadline(c net.Conn, enc codec.Encoder) bool {
	// unlimited
	if s.connectionDeadline == 0 {
		return true
	}

	err := c.SetDeadline(time.Now().Add(time.Duration(s.connectionDeadline) * time.Second))
	if err != nil {
		s.handler.writeMessage(enc, &codec.Message{Result: false, Data: "Internal error"})
		s.logger.Errorf("Failed to set connection deadline: %s:", err)
		return false
	}

	return true
}

func (s *server) acquireConnectionSlot(enc codec.Encoder) bool {
	// unlimited
	if s.maxClientsNum == 0 {
		return true
	}

	timeout := time.NewTimer(time.Duration(s.connectTimeout) * time.Second)
	select {
	case <-timeout.C:
		s.handler.writeMessage(enc, &codec.Message{Result: false, Data: "Connections limit reached"})
		return false
	case s.clientSem <- struct{}{}:
	}

	return true
}
