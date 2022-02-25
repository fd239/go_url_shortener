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
	Items       map[string]string
	Filename    string
	StoreInFile bool
	Producer    *producer
	Consumer    *consumer
}

type consumer struct {
	file    *os.File
	decoder *json.Decoder
}

func NewConsumer(fileName string) *consumer {
	err := os.MkdirAll(filepath.Dir(fileName), os.ModePerm)

	if err != nil {
		return nil
	}

	file, err := os.OpenFile(fileName, os.O_RDONLY|os.O_CREATE, 0777)

	if err != nil {
		return nil
	}

	return &consumer{
		file:    file,
		decoder: json.NewDecoder(file),
	}
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

func NewProducer(fileName string) *producer {
	err := os.MkdirAll(filepath.Dir(fileName), os.ModePerm)

	if err != nil {
		return nil
	}

	file, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0777)

	if err != nil {
		return nil
	}

	return &producer{
		file:    file,
		encoder: json.NewEncoder(file),
	}
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

	db.Items[hashString] = url

	if db.StoreInFile {
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
	if result, ok := db.Items[routeId]; ok {
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
		db.Items[entity.HashString] = entity.URL
	}

}

func InitDB() {
	DB = Database{
		StoreInFile: len(common.Cfg.FileStoragePath) > 0,
		Items:       make(map[string]string),
		Filename:    common.Cfg.FileStoragePath,
		Producer:    NewProducer(common.Cfg.FileStoragePath),
		Consumer:    NewConsumer(common.Cfg.FileStoragePath),
	}

	if DB.StoreInFile {
		DB.RestoreURLs()
	}

}
