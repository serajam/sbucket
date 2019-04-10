package server

import "encoding/gob"

// SBucketServer wraps all logic for accepting and handling connections
type SBucketServer interface {
	Start()
	Shutdown()
}

// SBucketLogger interface an be used for supplying any logger for server
type SBucketLogger interface {
	Error(interface{})
	Info(interface{})
	Debug(interface{})

	Errorf(string, ...interface{})
	Infof(string, ...interface{})
	Debugf(string, ...interface{})
}

// Middleware interface provides posibility to execure any kind of work before starting handling connection packets
type Middleware interface {
	Run(enc *gob.Encoder, dec *gob.Decoder) error
}
