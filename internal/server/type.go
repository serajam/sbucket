package server

import (
	"github.com/serajam/sbucket/internal/codec"
)

// SBucketServer wraps all logic for accepting and handling connections
type SBucketServer interface {
	Start()
	Shutdown()
}

// SBucketLogger interface can be used for supplying any logger for server
type SBucketLogger interface {
	Error(interface{})
	Info(interface{})
	Debug(interface{})

	Errorf(string, ...interface{})
	Infof(string, ...interface{})
	Debugf(string, ...interface{})
}

// Middleware interface provides possibility to execute any kind of work before starting handling connection packets
type Middleware interface {
	Run(codec codec.Codec) error
}
