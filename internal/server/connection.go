package server

import (
	"net"
	"sync"
)

type wrappedConn struct {
	net.Conn

	mu             sync.RWMutex
	active         bool
	pendingClosure bool
}

func (c *wrappedConn) setActive(s bool) {
	c.mu.Lock()
	c.active = s
	c.mu.Unlock()
}

func (c *wrappedConn) isActive() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.active
}

func (c *wrappedConn) isPendingClosure() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.pendingClosure
}
