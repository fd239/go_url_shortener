package storage

import (
	"context"
	"crypto/md5"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/fd239/go_url_shortener/config"
	"github.com/fd239/go_url_shortener/internal/app/common"
	_ "github.com/jackc/pgx/v4/stdlib"
	"log"
	"os"
	"path/filepath"
	"strings"
)

//PostgresSQLSuccessful insert statement have no duplicate by original url
const PostgresSQLSuccessful = 100000

//PostgresSQLDuplicate insert statement have duplicate by original url
const PostgresSQLDuplicate = 100001

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

type Item struct {
	ShortURL    string
	OriginalURL string
	Deleted     bool
	User        string
}

//Database repo struct with all possible options: im-mem, file and pg
type Database struct {
	Items       map[string]string
	UserItems   map[string][]*UserItem //map[userID][]UserItem
	ArrayItems  []*Item
	PGConn      *sql.DB
	Filename    string
	StoreInFile bool
	StoreInPg   bool
	StoreInArr  bool
	Producer    *producer
	Consumer    *consumer
}

type consumer struct {
	file    *os.File
	decoder *json.Decoder
}

//NewConsumer creating consumer by filename
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

//NewConsumer creating producer by filename
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

// SaveItems save from records from memory to file
func (db *Database) SaveItems() error {
	return db.Producer.encoder.Encode(db.Items)
}

// Close save file
func (p *producer) Close() error {
	return p.file.Close()
}

// getShortItem func for short url make
func (db *Database) getShortItem(item string) string {
	data := []byte(item)
	return fmt.Sprintf("%x", md5.Sum(data))
}

// Insert save short url and user ID to storage
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

	if db.StoreInArr {
		db.ArrayItems = append(db.ArrayItems, &Item{
			ShortURL:    hashString,
			OriginalURL: item,
			Deleted:     false,
			User:        userID,
		})
	}

	if db.StoreInPg {
		rows, err := db.PGConn.Query(insertStmt, item, hashString, userID)
		if err != nil {
			log.Println("PG Save items error: ", err.Error())
			return "", err
		}

		defer func(rows *sql.Rows) {
			err = rows.Close()
			if err != nil {
				log.Println("rows close error: ", err)
			}
		}(rows)

		if rows.Next() {
			shortURL := ""
			insertResult := 0
			err = rows.Scan(&shortURL, &insertResult)
			if err != nil {
				log.Println("rows scan error: ", err)
			}
			if insertResult == PostgresSQLDuplicate {
				return shortURL, common.ErrOriginalURLConflict
			}
		}

		err = rows.Err()
		if err != nil {
			log.Println("PG Insert rows err error: ", err.Error())
			return "", err
		}
	}

	return hashString, nil
}

// Get URL by id from storage
func (db *Database) Get(id string) (string, error) {
	if db.StoreInPg {
		var url string
		var deleted bool
		err := db.PGConn.QueryRow(getOriginalURLStmt, id).Scan(&url, &deleted)
		if err != nil {
			log.Println("PG Get short url query error: ", err.Error())
			return "", err
		}

		if deleted {
			return "", common.ErrURLDeleted
		}

		return url, nil
	}

	if db.StoreInArr {
		for _, item := range db.ArrayItems {
			if item.ShortURL == id {
				if item.Deleted {
					return "", common.ErrURLDeleted
				}
				return item.OriginalURL, nil
			}
		}
		return "", common.ErrUnableToFindURL
	}

	if result, ok := db.Items[id]; ok {
		return result, nil
	}

	return "", common.ErrUnableToFindURL
}

// GetUserURL receive all user urls by userID
func (db *Database) GetUserURL(userID string) ([]*UserItem, error) {
	if db.StoreInPg {
		userURLs := make([]*UserItem, 0)
		rows, err := db.PGConn.Query(getUserURL, userID)

		if err != nil {
			log.Println("PG Get user urls query error: ", err.Error())
			return nil, err
		}

		defer func(rows *sql.Rows) {
			err = rows.Close()
			if err != nil {
				log.Println("rows close error: ", err)
			}
		}(rows)

		for rows.Next() {
			userItem := &UserItem{}
			err = rows.Scan(&userItem.OriginalURL, &userItem.ShortURL)
			if err != nil {
				log.Println("PG Get user urls row scan error: ", err.Error())
				return nil, err
			}
			userURLs = append(userURLs, userItem)
		}

		err = rows.Err()
		if err != nil {
			log.Println("PG Get user urls rows err error: ", err.Error())
			return nil, err
		}

		return userURLs, nil
	}

	if db.StoreInArr {
		userURLs := make([]*UserItem, 0)
		for _, item := range db.ArrayItems {
			if item.User == userID {
				if !item.Deleted {
					userURLs = append(userURLs, &UserItem{
						ShortURL:    item.ShortURL,
						OriginalURL: item.OriginalURL,
					})
				}
				return userURLs, nil
			}
		}
		return nil, nil
	}

	return db.UserItems[userID], nil
}

// RestoreItems restore items from file to in-mem storage
func (db *Database) RestoreItems() error {
	err := db.Consumer.decoder.Decode(&db.Items)
	return err
}

// Ping postgres health check
func (db *Database) Ping() error {
	return db.PGConn.Ping()
}

// CreateItems batch insert items to postgres
func (db *Database) CreateItems(items []BatchItemRequest, userID string) ([]BatchItemResponse, error) {
	ctx := context.Background()
	tx, err := db.PGConn.Begin()
	if err != nil {
		log.Println("PG Context begin error: ", err.Error())
		return nil, err
	}

	defer func(tx *sql.Tx) {
		err = tx.Rollback()
		if err != nil {
			log.Println("transaction rollback error: ", err)
		}
	}(tx)

	stmt, err := tx.PrepareContext(ctx, batchInsert)

	if err != nil {
		log.Println("PG prepare context error: ", err.Error())
		tx.Rollback()
		return nil, err
	}

	defer func(stmt *sql.Stmt) {
		err = stmt.Close()
		if err != nil {
			log.Println("statement close error: ", err)
		}
	}(stmt)

	var batchItemsResponse []BatchItemResponse
	for _, item := range items {
		batchItemResponse := BatchItemResponse{}
		shortURL := db.getShortItem(item.OriginalURL)

		if _, err = stmt.ExecContext(ctx, item.CorrelationID, shortURL, item.OriginalURL, userID); err != nil {
			log.Println("PG exec context error: ", err.Error())
			return nil, err
		}

		batchItemResponse.CorrelationID = item.CorrelationID
		batchItemResponse.ShortURL = fmt.Sprintf("%s/%s", config.Cfg.BaseURL, shortURL)

		batchItemsResponse = append(batchItemsResponse, batchItemResponse)
	}

	if err = tx.Commit(); err != nil {
		log.Println("PG tx commit error: ", err.Error())
		return nil, err
	}

	return batchItemsResponse, nil
}

// UpdateItems batch update items in postgres
func (db *Database) UpdateItems(itemsIDs []string) error {
	formattedItems := make([]string, 0, len(itemsIDs))

	for _, item := range itemsIDs {
		formattedItem := fmt.Sprintf("('%s')", item)
		formattedItems = append(formattedItems, formattedItem)
	}

	stmt := "UPDATE short_url SET deleted = true FROM ( VALUES " + strings.Join(formattedItems, ",") + ") AS update_values (shortURL) WHERE short_url.short_url = update_values.shortURL;"
	_, err := db.PGConn.Exec(stmt)

	if err != nil {
		log.Printf("Items update error: %v\n", err)
		return err
	}

	return nil
}

//InitDB create DB repo and initialize it by config
func InitDB() (*Database, error) {
	storeInPg := len(config.Cfg.DatabaseDSN) > 0
	storeInFile := len(config.Cfg.FileStoragePath) > 0

	DB := Database{
		StoreInFile: storeInFile,
		StoreInPg:   storeInPg,
		Items:       make(map[string]string),
		ArrayItems:  make([]*Item, 0),
		UserItems:   make(map[string][]*UserItem),
		Filename:    config.Cfg.FileStoragePath,
	}

	if storeInFile {
		dataBaseProducer, err := NewProducer(config.Cfg.FileStoragePath)

		if err != nil {
			log.Println("Error producer creation: ", err.Error())
		}

		dataBaseConsumer, err := NewConsumer(config.Cfg.FileStoragePath)

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
		conn, err := sql.Open("pgx", config.Cfg.DatabaseDSN)
		if err != nil {
			log.Printf("Unable to connect to database: %v\n", err)
		}

		DB.PGConn = conn
		_, err = DB.PGConn.Exec(initStmt)

		if err != nil {
			log.Println("short url table creation error: ", err.Error())
			return nil, err
		}

	}

	return &DB, nil
}

// URLCount get saved url in storage
func (db *Database) URLCount() (counter int) {
	db.PGConn.QueryRow("SELECT count(*) FROM short_url").Scan(&counter)
	return
}

// UserCount get uniq url count in storage
func (db *Database) UserCount() (counter int) {
	db.PGConn.QueryRow("SELECT count(DISTINCT user_id) FROM short_url").Scan(&counter)
	return
}
