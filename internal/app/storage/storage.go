package storage

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"github.com/fd239/go_url_shortener/internal/app/common"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"log"
	"os"
	"path/filepath"
)

const PostgreSQLSuccessfull = 100000
const PostgreSQLDuplicate = 100001

type BatchItemRequest struct {
	CorrelationID string `json:"correlation_id" db:"id"`
	ShortURL      string `json:"short_url" db:"short_url"`
	OriginalURL   string `json:"original_url" db:"original_url"`
}

type BatchItemResponse struct {
	CorrelationID string `json:"correlation_id" db:"id"`
	ShortURL      string `json:"short_url" db:"short_url"`
}

type UserItem struct {
	ShortURL    string `json:"short_url" db:"short_url"`
	OriginalURL string `json:"original_url" db:"original_url"`
}

type Database struct {
	Items       map[string]string
	UserItems   map[string][]*UserItem //map[userID][]UserItem
	PGConn      *pgxpool.Pool
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
		rowID := uuid.NewString()
		stmt :=
			`WITH e AS (
			INSERT INTO short_url (original_url, short_url, id, user_id)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (original_url) DO NOTHING
		RETURNING short_url
		)
		SELECT short_url, 100000
		FROM e
		UNION ALL
		SELECT short_url, 100001
		FROM short_url
		WHERE original_url=$1;`
		rows, err := db.PGConn.Query(context.Background(), stmt, item, hashString, rowID, userID)
		if err != nil {
			log.Println("PG Save items error: ", err.Error())
			return "", err
		}

		defer rows.Close()

		if rows.Next() {
			shortURL := ""
			insertResult := 0
			rows.Scan(&shortURL, &insertResult)
			if insertResult == PostgreSQLDuplicate {
				return shortURL, common.ErrOriginalURLConflict
			}
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

	return "", common.ErrUnableToFindURL
}

func (db *Database) GetUserURL(userID string) []*UserItem {
	if db.StoreInPg {
		userURLs := make([]*UserItem, 0)
		rows, err := db.PGConn.Query(context.Background(), "select original_url, short_url from short_url where user_id=$1", userID)

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

func (db *Database) BatchItems(items []BatchItemRequest, userID string) ([]BatchItemResponse, error) {
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

		batch.Queue("INSERT INTO short_url (id, short_url, original_url, user_id) VALUES ($1, $2, $3, $4) ON CONFLICT (id) DO UPDATE SET id = excluded.id RETURNING id;", item.CorrelationID, shortURL, item.OriginalURL, userID)

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

func (db *Database) UpdateItems(itemsIDs []string, userID string) error {
	ctx := context.Background()
	tx, err := db.PGConn.Begin(ctx)
	if err != nil {
		log.Println("PG Context begin error: ", err.Error())
		return err
	}
	defer tx.Rollback(ctx)

	// New batch
	batch := &pgx.Batch{}

	for _, itemID := range itemsIDs {
		batch.Queue(`UPDATE short_url SET deleted = CASE user_id WHEN $1 THEN true else deleted END WHERE id = $2`, userID, itemID)
	}

	batchResults := tx.SendBatch(ctx, batch)

	var qerr error
	var rows pgx.Rows
	for qerr == nil {
		rows, qerr = batchResults.Query()
		rows.Close()
	}

	return tx.Commit(ctx)
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
		conn, err := pgxpool.Connect(context.Background(), common.Cfg.DatabaseDSN)
		if err != nil {
			log.Printf("Unable to connect to database: %v\n", err)
		}

		DB.PGConn = conn

		stmt :=
			`CREATE TABLE IF NOT EXISTS short_url
		(
			id           varchar(36) PRIMARY KEY NOT NULL,
			original_url varchar(150) UNIQUE     NOT NULL,
			short_url    varchar(50)             NOT NULL,
			user_id      varchar(50)
		)`

		_, err = DB.PGConn.Exec(context.Background(), stmt)

		if err != nil {
			log.Println("short url table creation error: ", err.Error())
		}

	}

	return &DB, nil

}
