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

func (s *actionsHandler) writeMessage(e codec.Encoder, msg *internal.Message) {
	err := e.Encode(msg)
	if err != nil {
		s.logger.Errorf("Failed to write response: %s:\n", err)
		return
	}
}

func (s *actionsHandler) handleCreateBucket(enc codec.Encoder, m *internal.Message) {
	err := s.storage.NewBucket(m.Value)
	if err != nil {
		s.writeMessage(enc, &internal.Message{Result: false, Data: err.Error()})
		return
	}
	s.writeMessage(enc, &internal.Message{Result: true})
}

func (s *actionsHandler) handleDeleteBucket(enc codec.Encoder, m *internal.Message) {
	err := s.storage.DelBucket(m.Value)
	if err != nil {
		s.writeMessage(enc, &internal.Message{Result: false, Data: err.Error()})
		return
	}
	s.writeMessage(enc, &internal.Message{Result: true})
}

func (s *actionsHandler) handleAdd(enc codec.Encoder, m *internal.Message) {
	err := s.storage.Add(m.Bucket, m.Key, m.Value)
	if err != nil {
		go s.writeMessage(enc, &internal.Message{Result: false, Data: err.Error()})
		return
	}
	go s.writeMessage(enc, &internal.Message{Result: true})
}

func (s *actionsHandler) handleGet(enc codec.Encoder, m *internal.Message) {
	v, err := s.storage.Get(m.Bucket, m.Key)
	if err != nil {
		s.writeMessage(enc, &internal.Message{Result: false, Data: err.Error()})
		return
	}
	s.writeMessage(enc, &internal.Message{Value: v.Value(), Result: true})
}

func (s *actionsHandler) handlePing(enc codec.Encoder, m *internal.Message) {
	s.logger.Debug("Received Ping")
}