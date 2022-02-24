package app

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"github.com/fd239/go_url_shortener/internal/app/common"
	"log"
	"os"
	"path/filepath"
)

type UrlEntity struct {
	HashString string
	URL        string
}

type Database struct {
	LocalStruct  map[string]string
	Filename     string
	LocalStorage bool
	Producer     *producer
	Consumer     *consumer
}

type consumer struct {
	file    *os.File
	decoder *json.Decoder
}

func NewConsumer(fileName string) (*consumer, error) {
	err := os.MkdirAll(filepath.Dir(fileName), os.ModePerm)
	file, err := os.OpenFile(fileName, os.O_RDONLY|os.O_CREATE, 0777)
	if err != nil {
		return nil, err
	}
	return &consumer{
		file:    file,
		decoder: json.NewDecoder(file),
	}, nil
}

func (c *consumer) HasNext() bool {
	return c.decoder.More()
}

func (c *consumer) Close() error {
	return c.file.Close()
}

type producer struct {
	file    *os.File
	encoder *json.Encoder
}

func NewProducer(fileName string) (*producer, error) {
	err := os.MkdirAll(filepath.Dir(fileName), os.ModePerm)
	file, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0777)
	if err != nil {
		return nil, err
	}
	return &producer{
		file:    file,
		encoder: json.NewEncoder(file),
	}, nil
}

func (p *producer) WriteURL(UrlEntity *UrlEntity) error {
	return p.encoder.Encode(UrlEntity)
}
func (p *producer) Close() error {
	return p.file.Close()
}

var DB Database

func (db *Database) SaveShortRoute(url string) (string, error) {
	var err error
	data := []byte(url)
	hashString := fmt.Sprintf("%x", md5.Sum(data))

	db.LocalStruct[hashString] = url

	if !db.LocalStorage {
		err := db.Producer.WriteURL(&UrlEntity{
			URL:        url,
			HashString: hashString,
		})
		if err != nil {
			log.Println("Write url to file error")
		}
	}

	return hashString, err

}

func (db *Database) GetShortRoute(routeId string) (string, error) {
	if result, ok := db.LocalStruct[routeId]; ok {
		return result, nil
	}

	return "", common.ErrNoUrlInMap
}

func (db *Database) RestoreURLs() {
	for db.Consumer.HasNext() {
		entity := UrlEntity{}
		err := db.Consumer.decoder.Decode(&entity)
		if err != nil {
			log.Println("Error db file decode")
			continue
		}
		db.LocalStruct[entity.HashString] = entity.URL
	}

}

func InitDB() {

	producer, _ := NewProducer(common.Cfg.FileStoragePath)
	consumer, _ := NewConsumer(common.Cfg.FileStoragePath)

	DB = Database{
		LocalStorage: len(common.Cfg.FileStoragePath) == 0,
		LocalStruct:  make(map[string]string),
		Filename:     common.Cfg.FileStoragePath,
		Producer:     producer,
		Consumer:     consumer,
	}

	if !DB.LocalStorage {
		DB.RestoreURLs()
	}

}
