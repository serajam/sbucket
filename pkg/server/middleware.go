package server

import (
	"encoding/gob"
	"errors"
	"fmt"
	"github.com/serajam/sbucket/pkg"
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

// Run execute middleware
func (r AuthMiddleware) Run(enc *gob.Encoder, dec *gob.Decoder) error {
	authMsg := pkg.AuthMessage{}
	err := dec.Decode(&authMsg)
	if err != nil {
		errMsg := errors.New("auth failed: unable to decode message")
		err := enc.Encode(&pkg.Message{Result: false, Data: errMsg.Error()})
		if err != nil {
			return fmt.Errorf("message encode failed: %s", err)
		}
		return errMsg
	}

	err = r.authenticate(authMsg.Login, authMsg.Password)
	if err != nil {
		errEncode := enc.Encode(&pkg.Message{Result: false, Data: err.Error()})
		if errEncode != nil {
			return fmt.Errorf("message `%s` encode failed: %s", err, errEncode)
		}
		return err
	}

	err = enc.Encode(&pkg.Message{Result: true})
	if err != nil {
		return fmt.Errorf("message encode failed: %s", err)
	}

	return nil
}
