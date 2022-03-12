package storage

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"github.com/fd239/go_url_shortener/internal/app/common"
	"github.com/jackc/pgx/v4"
	"log"
	"os"
	"path/filepath"
)

type UserItem struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type Database struct {
	Items       map[string]string
	UserItems   map[string][]*UserItem //map[userID][]UserItem
	PGConn      *pgx.Conn
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

func (db *Database) Insert(item string, userID string) (string, error) {
	data := []byte(item)
	hashString := fmt.Sprintf("%x", md5.Sum(data))

	db.Items[hashString] = item

	db.UserItems[userID] = append(db.UserItems[userID], &UserItem{
		ShortURL:    hashString,
		OriginalURL: item,
	})

	if db.StoreInFile {
		err := db.SaveItems()
		if err != nil {
			log.Println("DB Save items error: ", err.Error())
			return "", err
		}
	}

	return hashString, nil

}

func (db *Database) Get(id string) (string, error) {
	if result, ok := db.Items[id]; ok {
		return result, nil
	}

	return "", common.ErrNoURLInMap
}

func (db *Database) GetUserURL(userID string) []*UserItem {
	return db.UserItems[userID]
}

func (db *Database) RestoreItems() error {
	err := db.Consumer.decoder.Decode(&db.Items)

	return err

}

func (db *Database) Ping() error {
	return db.PGConn.Ping(context.Background())
}

func InitDB() (*Database, error) {

	conn, err := pgx.Connect(context.Background(), common.Cfg.DatabaseDSN)
	if err != nil {
		log.Printf("Unable to connect to database: %v\n", err)
	}

	storeInFile := len(common.Cfg.FileStoragePath) > 0

	DB := Database{
		StoreInFile: storeInFile,
		Items:       make(map[string]string),
		UserItems:   make(map[string][]*UserItem),
		Filename:    common.Cfg.FileStoragePath,
		PGConn:      conn,
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
			log.Println("Error db file decode: ", err)
		}

	}

	return &DB, nil

}
