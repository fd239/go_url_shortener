package common

import "errors"

var (
	ErrUnableToFindURL     = errors.New("unable to find original url")
	ErrEmptyBody           = errors.New("empty body")
	ErrBodyReadError       = errors.New("body read error")
	ErrUserCookie          = errors.New("no user cookie")
	ErrNoUserURLs          = errors.New("no user URLs")
	ErrPing                = errors.New("database ping error")
	ErrOriginalURLConflict = errors.New("original url postgresql save conflict")
	ErrResponseEncode      = errors.New("response encode error")
	ErrGzipRead            = errors.New("gzip read error")
	ErrURLDeleted          = errors.New("url deleted")
)
