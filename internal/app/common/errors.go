package common

import "errors"

var (
	ErrNoURLInMap    = errors.New("no URL in map")
	ErrEmptyBody     = errors.New("empty body")
	ErrBodyReadError = errors.New("body read error")
	ErrUserCookie    = errors.New("no user cookie")
	ErrNoUserURLs    = errors.New("no user URLs")
	ErrPing          = errors.New("database ping error")
)
