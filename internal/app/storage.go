package app

import (
	"crypto/md5"
	"errors"
	"fmt"
	"github.com/fd239/go_url_shortener/internal/app/_const"
)

var urlMap = map[string]string{}

func SaveShortRoute(url string) string {
	data := []byte(url)
	hashString := fmt.Sprintf("%x", md5.Sum(data))

	urlMap[hashString] = url

	return hashString
}

func GetShortRoute(routeId string) (string, error) {
	if result, ok := urlMap[routeId]; ok {
		return result, nil
	}

	return "", errors.New(_const.ErrMsg_NoUrlInMap)

}
