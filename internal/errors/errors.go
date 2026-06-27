package errors

import "errors"

var (
	ErrNoRoute          = errors.New("no valid route")
	ErrNodeUnavailable  = errors.New("node unavailable")
	ErrPayloadCorrupted = errors.New("payload corrupted")
)
