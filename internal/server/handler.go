package server

import (
	"github.com/serajam/sbucket/internal"
	"github.com/serajam/sbucket/internal/codec"
	"github.com/serajam/sbucket/internal/storage"
)

type actionsHandler struct {
	storage storage.SBucketStorage
	logger  SBucketLogger
}

func (s *actionsHandler) writeMessage(e codec.Encoder, msg *codec.Message) {
	err := e.Encode(msg)
	if err != nil {
		s.logger.Errorf("Failed to write response: %s:\n", err)
		return
	}
}

func (s *actionsHandler) handleCreateBucket(enc codec.Encoder, m *codec.Message) {
	msg := &codec.Message{Command: internal.ResultCommand, Result: true}

	err := s.storage.NewBucket(m.Value)
	if err != nil {
		msg.Result = false
		msg.Data = err.Error()
		s.writeMessage(enc, msg)
		return
	}

	s.writeMessage(enc, msg)
}

func (s *actionsHandler) handleDeleteBucket(enc codec.Encoder, m *codec.Message) {
	msg := &codec.Message{Command: internal.ResultCommand, Result: true}

	err := s.storage.DelBucket(m.Value)
	if err != nil {
		msg.Result = false
		msg.Data = err.Error()
		s.writeMessage(enc, msg)
		return
	}

	s.writeMessage(enc, msg)
}

func (s *actionsHandler) handleAdd(enc codec.Encoder, m *codec.Message) {
	msg := &codec.Message{Command: internal.ResultCommand, Result: true}

	err := s.storage.Add(m.Bucket, m.Key, m.Value)
	if err != nil {
		msg.Result = false
		msg.Data = err.Error()
		go s.writeMessage(enc, msg)
		return
	}

	go s.writeMessage(enc, msg)
}

func (s *actionsHandler) handleGet(enc codec.Encoder, m *codec.Message) {
	msg := &codec.Message{Command: internal.ResultCommand, Result: true}

	v, err := s.storage.Get(m.Bucket, m.Key)
	if err != nil {
		msg.Result = false
		msg.Data = err.Error()
		s.writeMessage(enc, msg)
		return
	}

	msg.Value = v.Value()
	s.writeMessage(enc, msg)
}

func (s *actionsHandler) handlePing(enc codec.Encoder, m *codec.Message) {
	s.logger.Debug("Received Ping")
}
