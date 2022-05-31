package storage

import (
	"database/sql"
	"fmt"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/fd239/go_url_shortener/internal/app/common"
	"github.com/stretchr/testify/assert"
	"log"
	"regexp"
	"testing"
)

var testUserID = "testUser"

func getProducer() *producer {
	prod, err := NewProducer(common.TestDBName)
	if err != nil {
		log.Println("Consumer creation error: ")
	}
	return prod
}

func getConsumer() *consumer {
	cons, err := NewConsumer(common.TestDBName)
	if err != nil {
		log.Println("Consumer creation error: ", err.Error())
	}
	return cons
}

func TestDatabase_SaveShortRoute(t *testing.T) {
	type fields struct {
		Items       map[string]string
		Filename    string
		StoreInFile bool
		StoreInPg   bool
		Producer    *producer
		Consumer    *consumer
	}
	type args struct {
		url string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr assert.ErrorAssertionFunc
	}{
		{
			"OK",
			fields{
				Items:       map[string]string{common.TestURL: common.TestShortID},
				Filename:    common.TestDBName,
				StoreInFile: true,
				StoreInPg:   false,
				Producer:    getProducer(),
				Consumer:    getConsumer(),
			},
			args{common.TestURL},
			common.TestShortID,
			assert.NoError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := &Database{
				Items:       tt.fields.Items,
				Filename:    tt.fields.Filename,
				StoreInFile: tt.fields.StoreInFile,
				Producer:    tt.fields.Producer,
				Consumer:    tt.fields.Consumer,
			}

			got, err := db.Get(tt.args.url)
			if !tt.wantErr(t, err, fmt.Sprintf("SaveShortRoute(%v)", tt.args.url)) {
				return
			}
			assert.Equalf(t, tt.want, got, "SaveShortRoute(%v)", tt.args.url)
		})
	}
}

func TestDatabase_GetShortRoute(t *testing.T) {
	type fields struct {
		Items       map[string]string
		Filename    string
		StoreInFile bool
		StoreInPg   bool
		Producer    *producer
		Consumer    *consumer
	}
	type args struct {
		routeID string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr assert.ErrorAssertionFunc
	}{
		{
			"OK",
			fields{
				Items:       map[string]string{common.TestShortID: common.TestURL},
				Filename:    common.TestDBName,
				StoreInFile: true,
				StoreInPg:   false,
				Producer:    getProducer(),
				Consumer:    getConsumer(),
			},
			args{common.TestShortID},
			common.TestURL,
			assert.NoError,
		},
		{
			"Error",
			fields{
				Items:       map[string]string{},
				Filename:    common.TestDBName,
				StoreInFile: true,
				StoreInPg:   false,
				Producer:    getProducer(),
				Consumer:    getConsumer(),
			},
			args{common.TestURL},
			"",
			assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := &Database{
				Items:       tt.fields.Items,
				Filename:    tt.fields.Filename,
				StoreInFile: tt.fields.StoreInFile,
				Producer:    tt.fields.Producer,
				Consumer:    tt.fields.Consumer,
			}
			got, err := db.Get(tt.args.routeID)
			if !tt.wantErr(t, err, fmt.Sprintf("GetShortRoute(%v)", tt.args.routeID)) {
				return
			}
			assert.Equalf(t, tt.want, got, "GetShortRoute(%v)", tt.args.routeID)
		})
	}
}

func TestDatabase_Insert(t *testing.T) {
	type fields struct {
		Items       map[string]string
		UserItems   map[string][]*UserItem
		PGConn      *sql.DB
		Filename    string
		StoreInFile bool
		StoreInPg   bool
		Producer    *producer
		Consumer    *consumer
	}
	type args struct {
		item   string
		userID string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "OK IN memory",
			fields: fields{
				Items:       map[string]string{},
				UserItems:   map[string][]*UserItem{},
				PGConn:      nil,
				Filename:    "",
				StoreInFile: false,
				StoreInPg:   false,
				Producer:    nil,
				Consumer:    nil,
			},
			args: args{
				item:   common.TestURL,
				userID: "test",
			},
			want:    common.TestShortID,
			wantErr: assert.NoError,
		},
		{
			name: "OK IN file",
			fields: fields{
				Items:       map[string]string{},
				UserItems:   map[string][]*UserItem{},
				PGConn:      nil,
				Filename:    "",
				StoreInFile: true,
				StoreInPg:   false,
				Producer:    getProducer(),
				Consumer:    nil,
			},
			args: args{
				item:   common.TestURL,
				userID: "test",
			},
			want:    common.TestShortID,
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := &Database{
				Items:       tt.fields.Items,
				UserItems:   tt.fields.UserItems,
				PGConn:      tt.fields.PGConn,
				Filename:    tt.fields.Filename,
				StoreInFile: tt.fields.StoreInFile,
				StoreInPg:   tt.fields.StoreInPg,
				Producer:    tt.fields.Producer,
				Consumer:    tt.fields.Consumer,
			}
			got, err := db.Insert(tt.args.item, tt.args.userID)
			if !tt.wantErr(t, err, fmt.Sprintf("Insert(%v, %v)", tt.args.item, tt.args.userID)) {
				return
			}
			assert.Equalf(t, tt.want, got, "Insert(%v, %v)", tt.args.item, tt.args.userID)
		})
	}
}

func TestInsertURLOK(t *testing.T) {
	var res string

	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer func(conn *sql.DB) {
		err = conn.Close()
		if err != nil {
			t.Errorf("conn close error: %v", err)
		}
	}(conn)

	rows := sqlmock.NewRows([]string{"shortURL", "insertResult"}).AddRow(common.TestShortID, PostgreSQLSuccessful)
	mock.ExpectQuery(regexp.QuoteMeta(insertStmt)).WithArgs(common.TestURL, common.TestShortID, testUserID).WillReturnRows(rows)

	var db = &Database{
		Items:       map[string]string{},
		UserItems:   map[string][]*UserItem{},
		PGConn:      conn,
		Filename:    "",
		StoreInFile: false,
		StoreInPg:   true,
		Producer:    nil,
		Consumer:    nil,
	} // now we execute our method

	if res, err = db.Insert(common.TestURL, testUserID); err != nil {
		t.Errorf("error was not expected while inserting: %s", err)
	}

	assert.Equal(t, res, common.TestShortID)

	// we make sure that all expectations were met
	if err = mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestInsertDuplicateErr(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer func(conn *sql.DB) {
		err = conn.Close()
		if err != nil {
			t.Errorf("conn close error: %v", err)
		}
	}(conn)

	rows := sqlmock.NewRows([]string{"shortURL", "insertResult"}).AddRow(common.TestShortID, PostgreSQLDuplicate)
	mock.ExpectQuery(regexp.QuoteMeta(insertStmt)).WithArgs(common.TestURL, common.TestShortID, testUserID).WillReturnRows(rows)

	var db = &Database{
		Items:       map[string]string{},
		UserItems:   map[string][]*UserItem{},
		PGConn:      conn,
		Filename:    "",
		StoreInFile: false,
		StoreInPg:   true,
		Producer:    nil,
		Consumer:    nil,
	} // now we execute our method

	if _, err = db.Insert(common.TestURL, testUserID); err == nil {
		t.Error("error was expected while inserting")
	}

	assert.ErrorIs(t, err, common.ErrOriginalURLConflict)

	// we make sure that all expectations were met
	if err = mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestGetURLOK(t *testing.T) {
	var res string

	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer func(conn *sql.DB) {
		err = conn.Close()
		if err != nil {
			t.Errorf("conn close error: %v", err)
		}
	}(conn)

	rows := sqlmock.NewRows([]string{"url", "deleted"}).AddRow(common.TestURL, false)
	mock.ExpectQuery(regexp.QuoteMeta(getOriginalURLStmt)).WithArgs(common.TestShortID).WillReturnRows(rows)

	var db = &Database{
		Items:       map[string]string{},
		UserItems:   map[string][]*UserItem{},
		PGConn:      conn,
		Filename:    "",
		StoreInFile: false,
		StoreInPg:   true,
		Producer:    nil,
		Consumer:    nil,
	} // now we execute our method

	if res, err = db.Get(common.TestShortID); err != nil {
		t.Errorf("error was not expected while get: %s", err)
	}

	assert.Equal(t, res, common.TestURL)

	// we make sure that all expectations were met
	if err = mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestGetURLDeletedError(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer func(conn *sql.DB) {
		err := conn.Close()
		if err != nil {

		}
	}(conn)

	rows := sqlmock.NewRows([]string{"url", "deleted"}).AddRow(common.TestURL, true)
	mock.ExpectQuery(regexp.QuoteMeta(getOriginalURLStmt)).WithArgs(common.TestShortID).WillReturnRows(rows)

	var db = &Database{
		Items:       map[string]string{},
		UserItems:   map[string][]*UserItem{},
		PGConn:      conn,
		Filename:    "",
		StoreInFile: false,
		StoreInPg:   true,
		Producer:    nil,
		Consumer:    nil,
	} // now we execute our method

	if _, err = db.Get(common.TestShortID); err == nil {
		t.Errorf("error was expected while inserting: %s", err)
	}

	assert.ErrorIs(t, err, common.ErrURLDeleted)

	// we make sure that all expectations were met
	if err = mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestGetUserURLOK(t *testing.T) {
	var res []*UserItem

	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer func(conn *sql.DB) {
		err := conn.Close()
		if err != nil {

		}
	}(conn)

	rows := sqlmock.NewRows([]string{"OriginalURL", "ShortURL"}).AddRow(common.TestURL, common.TestShortID)
	mock.ExpectQuery(regexp.QuoteMeta(getUserURL)).WithArgs(testUserID).WillReturnRows(rows)

	var db = &Database{
		Items:       map[string]string{},
		UserItems:   map[string][]*UserItem{},
		PGConn:      conn,
		Filename:    "",
		StoreInFile: false,
		StoreInPg:   true,
		Producer:    nil,
		Consumer:    nil,
	} // now we execute our method

	if res, err = db.GetUserURL(testUserID); err != nil {
		t.Errorf("error was not expected while get user url: %s", err)
	}

	assert.Equal(t, res, []*UserItem{{
		ShortURL:    common.TestShortID,
		OriginalURL: common.TestURL,
	}})

	// we make sure that all expectations were met
	if err = mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestGetUserURLPgErr(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer func(conn *sql.DB) {
		err := conn.Close()
		if err != nil {

		}
	}(conn)

	rows := sqlmock.NewRows([]string{"test"}).AddRow("test")
	mock.ExpectQuery(regexp.QuoteMeta(getUserURL)).WithArgs(testUserID).WillReturnRows(rows)

	var db = &Database{
		Items:       map[string]string{},
		UserItems:   map[string][]*UserItem{},
		PGConn:      conn,
		Filename:    "",
		StoreInFile: false,
		StoreInPg:   true,
		Producer:    nil,
		Consumer:    nil,
	} // now we execute our method

	if _, err = db.GetUserURL(testUserID); err == nil {
		t.Error("error was expected while get user url")
	}

	assert.Error(t, err)
}

func TestGetUserURLPgRowsErr(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer func(conn *sql.DB) {
		err = conn.Close()
		if err != nil {
			t.Errorf("conn close error: %v", err)
		}
	}(conn)

	mock.ExpectQuery(regexp.QuoteMeta(getUserURL)).WithArgs(testUserID).WillReturnError(fmt.Errorf("some error"))

	var db = &Database{
		Items:       map[string]string{},
		UserItems:   map[string][]*UserItem{},
		PGConn:      conn,
		Filename:    "",
		StoreInFile: false,
		StoreInPg:   true,
		Producer:    nil,
		Consumer:    nil,
	} // now we execute our method

	if _, err = db.GetUserURL(testUserID); err == nil {
		t.Error("error was expected while get user url")
	}

	assert.Error(t, err)
}
