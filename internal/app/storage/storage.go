package storage

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"github.com/fd239/go_url_shortener/internal/app/common"
	"log"
	"os"
	"path/filepath"
)

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

func NewConsumer(fileName string) (*consumer, error) {
	err := os.MkdirAll(filepath.Dir(fileName), os.ModePerm)

	if err != nil {
		return nil, err
	}

	file, err := os.OpenFile(fileName, os.O_RDONLY|os.O_CREATE, 0777)

	if err != nil {
		return nil, err
	}

	return &consumer{
		file:    file,
		decoder: json.NewDecoder(file),
	}, err
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

	if err != nil {
		return nil, err
	}

	file, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0777)

	if err != nil {
		return nil, err
	}

	return &producer{
		file:    file,
		encoder: json.NewEncoder(file),
	}, err
}

func (db *Database) SaveItems() error {
	return db.Producer.encoder.Encode(db.Items)
}

func (p *producer) Close() error {
	return p.file.Close()
}

func (db *Database) Insert(item string) (string, error) {
	var err error
	data := []byte(item)
	hashString := fmt.Sprintf("%x", md5.Sum(data))

	db.Items[hashString] = item

	if db.StoreInFile {
		db.SaveItems()
	}

	return hashString, err

}

func (db *Database) Get(id string) (string, error) {
	if result, ok := db.Items[id]; ok {
		return result, nil
	}

	return "", common.ErrNoUrlInMap
}

func (db *Database) RestoreItems() error {
	err := db.Consumer.decoder.Decode(&db.Items)

	return err

}

func InitDB() (*Database, error) {
	var err error
	storeInFile := len(common.Cfg.FileStoragePath) > 0

	DB := Database{
		StoreInFile: storeInFile,
		Items:       make(map[string]string),
		Filename:    common.Cfg.FileStoragePath,
	}

	if storeInFile {
		dataBaseProducer, err := NewProducer(common.Cfg.FileStoragePath)

		if err != nil {
			log.Println("Error producer creation: ", err.Error())
		}

		dataBaseConsumer, err := NewConsumer(common.Cfg.FileStoragePath)

		if err != nil {
			log.Println("Error consumer creation: ", err.Error())
		}

		DB.Producer = dataBaseProducer
		DB.Consumer = dataBaseConsumer

	}

	if storeInFile {
		err := DB.RestoreItems()
		if err != nil {
			log.Println("Error db file decode")
		}

	}

	return &DB, err

}
