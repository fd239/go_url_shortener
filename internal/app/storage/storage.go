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

type BatchItemRequest struct {
	CorrelationID string `json:"correlation_id" db:"correlation_id"`
	ShortURL      string `json:"short_url" db:"short_url"`
	OriginalURL   string `json:"original_url" db:"original_url"`
}

type BatchItemResponse struct {
	CorrelationID string `json:"correlation_id" db:"correlation_id"`
	ShortURL      string `json:"short_url" db:"short_url"`
}

type UserItem struct {
	ShortURL    string `json:"short_url" db:"short_url"`
	OriginalURL string `json:"original_url" db:"original_url"`
}

type Database struct {
	Items       map[string]string
	UserItems   map[string][]*UserItem //map[userID][]UserItem
	PGConn      *pgx.Conn
	Filename    string
	StoreInFile bool
	StoreInPg   bool
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

func (db *Database) getShortItem(item string) string {
	data := []byte(item)
	return fmt.Sprintf("%x", md5.Sum(data))
}

func (db *Database) Insert(item string, userID string) (string, error) {
	hashString := db.getShortItem(item)

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

	if db.StoreInPg {
		if _, err := db.PGConn.Exec(context.Background(), `insert into short_url(original_url, short_url) values ($1, $2)`, item, hashString); err != nil {
			log.Println("PG Save items error: ", err.Error())
			return "", err
		}

		if _, err := db.PGConn.Exec(context.Background(), `insert into user_short_url(original_url, short_url, userID) values ($1, $2, $3)`, item, hashString, userID); err != nil {
			log.Println("PG Save user short url error: ", err.Error())
			return "", err
		}

	}

	return hashString, nil

}

func (db *Database) Get(id string) (string, error) {
	if db.StoreInPg {
		var url string
		err := db.PGConn.QueryRow(context.Background(), "select original_url from short_url where short_url=$1", id).Scan(&url)
		if err != nil {
			log.Println("PG Get short url query error: ", err.Error())
			return "", err
		}
		return url, nil
	}

	if result, ok := db.Items[id]; ok {
		return result, nil
	}

	return "", common.ErrNoURLInMap
}

func (db *Database) GetUserURL(userID string) []*UserItem {
	if db.StoreInPg {
		userURLs := make([]*UserItem, 0)
		rows, err := db.PGConn.Query(context.Background(), "select original_url, short_url from user_short_url where userID=$1", userID)

		if err != nil {
			log.Println("PG Get user urls query error: ", err.Error())
			return nil
		}

		defer rows.Close()

		for rows.Next() {
			userItem := &UserItem{}
			err = rows.Scan(&userItem.OriginalURL, &userItem.ShortURL)
			if err != nil {
				log.Println("PG Get user urls row scan error: ", err.Error())
			}
			userURLs = append(userURLs, userItem)
		}

		err = rows.Err()
		if err != nil {
			log.Println("PG Get user urls rows err error: ", err.Error())
			return nil
		}

		return userURLs

	}

	return db.UserItems[userID]
}

func (db *Database) RestoreItems() error {
	err := db.Consumer.decoder.Decode(&db.Items)

	return err

}

func (db *Database) Ping() error {
	return db.PGConn.Ping(context.Background())
}

func (db *Database) BatchItems(items []BatchItemRequest) ([]BatchItemResponse, error) {
	ctx := context.Background()
	tx, err := db.PGConn.Begin(ctx)
	if err != nil {
		log.Println("PG Context begin error: ", err.Error())
		return nil, err
	}

	// New batch
	batch := &pgx.Batch{}
	var batchItemsResponse []BatchItemResponse

	for _, item := range items {

		batchItemResponse := BatchItemResponse{}

		shortURL := db.getShortItem(item.OriginalURL)

		batch.Queue("INSERT INTO batch_url (correlation_id, short_url, original_url) VALUES ($1, $2, $3);", item.CorrelationID, shortURL, item.OriginalURL)

		batchItemResponse.CorrelationID = item.CorrelationID
		batchItemResponse.ShortURL = fmt.Sprintf("%s/%s", common.Cfg.BaseURL, shortURL)

		batchItemsResponse = append(batchItemsResponse, batchItemResponse)

	}

	batchResults := tx.SendBatch(ctx, batch)

	var qerr error
	var rows pgx.Rows
	for qerr == nil {
		rows, qerr = batchResults.Query()
		rows.Close()
	}

	return batchItemsResponse, tx.Commit(ctx)

}

func InitDB() (*Database, error) {

	storeInPg := len(common.Cfg.DatabaseDSN) > 0
	storeInFile := len(common.Cfg.FileStoragePath) > 0

	DB := Database{
		StoreInFile: storeInFile,
		StoreInPg:   storeInPg,
		Items:       make(map[string]string),
		UserItems:   make(map[string][]*UserItem),
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

		err = DB.RestoreItems()
		if err != nil {
			log.Println("Error db file decode: ", err.Error())
		}

	}

	if storeInPg {
		conn, err := pgx.Connect(context.Background(), common.Cfg.DatabaseDSN)
		if err != nil {
			log.Printf("Unable to connect to database: %v\n", err)
		}

		DB.PGConn = conn

		_, err = DB.PGConn.Exec(context.Background(), `CREATE TABLE IF NOT EXISTS short_url (original_url varchar(150) NOT NULL, short_url varchar(50) NOT NULL)`)
		if err != nil {
			log.Println("short url table creation error: ", err.Error())
		}

		_, err = DB.PGConn.Exec(context.Background(), `CREATE TABLE IF NOT EXISTS user_short_url (original_url varchar(150) NOT NULL, short_url varchar(50) NOT NULL, userID varchar(50))`)
		if err != nil {
			log.Println("user short url table creation error: ", err.Error())
		}

		_, err = DB.PGConn.Exec(context.Background(), `CREATE TABLE IF NOT EXISTS batch_url (original_url varchar(150) NOT NULL, short_url varchar(50) NOT NULL, correlation_id varchar(36) NOT NULL)`)
		if err != nil {
			log.Println("user short url table creation error: ", err.Error())
		}

	}

	return &DB, nil

}
