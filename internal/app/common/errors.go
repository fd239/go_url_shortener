package common

import "errors"

var (
	ErrNoUrlInMap    = errors.New("no URL in map")
	ErrEmptyBody     = errors.New("empty body")
	ErrBodyReadError = errors.New("body read error")
)
