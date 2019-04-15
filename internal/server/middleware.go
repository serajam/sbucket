package server

import (
	"errors"
	"fmt"
	"github.com/serajam/sbucket/internal"
	"github.com/serajam/sbucket/internal/codec"
)

// AuthMiddleware can be to proved secure access to server
type AuthMiddleware struct {
	Login, Pass string
}

func (r AuthMiddleware) authenticate(login, pass string) error {
	if r.Login != login {
		return errors.New("invalid credentials")
	}

	if r.Pass != pass {
		return errors.New("invalid credentials")
	}

	return nil
}

func (r AuthMiddleware) write(c codec.Encoder, msg *internal.Message) error {
	err := c.Encode(msg)
	if err != nil {
		return fmt.Errorf("message %v encode failed: %s", *msg, err)
	}
	return nil
}

// Run execute middleware
func (r AuthMiddleware) Run(c codec.Codec) error {
	authMsg := internal.AuthMessage{}
	err := c.Decode(&authMsg)
	if err != nil {
		errMsg := errors.New("auth failed: unable to decode message")
		return r.write(c, &internal.Message{Result: false, Data: errMsg.Error()})
	}

	err = r.authenticate(authMsg.Login, authMsg.Password)
	if err != nil {
		return r.write(c, &internal.Message{Result: false, Data: err.Error()})
	}

	return r.write(c, &internal.Message{Result: true})
}
