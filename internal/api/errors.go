package api

import "errors"

var (
	ErrUnauthorized = errors.New("unauthorized")
	ErrLoginAborted = errors.New("login aborted")
)
