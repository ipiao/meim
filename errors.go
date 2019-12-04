package meim

import "errors"

var (
	ErrRingEmpty   = errors.New("ring buffer empty")
	ErrRingFull    = errors.New("ring buffer full")
	ErrRoomDropped = errors.New("room dropped")
)
