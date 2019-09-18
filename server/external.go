package server

import (
	"errors"
)

var (
	HandleConnAccepted func(Conn)
	HandleCloseConn    func(Conn)
	HandleConnClosed   func(Conn)
)

//
func CheckExternalHandlers() error {
	if HandleConnAccepted == nil {
		return errors.New("external handler HandleConnAccepted not set")
	}

	if HandleCloseConn == nil {
		return errors.New("external handler HandleCloseConn not set")
	}

	if HandleConnClosed == nil {
		return errors.New("external handler HandleConnClosed not set")
	}
	return nil
}
