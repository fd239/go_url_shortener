package app

import (
	"crypto/md5"
	"fmt"
)

func GetShortLink(s string) string {
	data := []byte(s)
	return fmt.Sprintf("%x", md5.Sum(data))
}
