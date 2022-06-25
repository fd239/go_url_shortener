package storage

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/fd239/go_url_shortener/internal/app/common"
	"github.com/stretchr/testify/assert"
	"log"
	"regexp"
	"testing"
)

var testUserID = "testUser"
var testItemID = "123"

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

func setupTestDatabase(t *testing.T) (*Database, sqlmock.Sqlmock) {
	conn, mock, err := sqlmock.New()

	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

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

	return db, mock
}

func TestDatabase_SaveShortRoute(t *testing.T) {
	type fields struct {
		Items       map[string]string
		Filename    string
		StoreInFile bool
		StoreInPg   bool
		StoreInArr  bool
		ArrayItems  []*Item
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
		{
			"OK Arr",
			fields{
				ArrayItems:  []*Item{{common.TestShortID, common.TestURL, false, testUserID}},
				Filename:    common.TestDBName,
				StoreInFile: false,
				StoreInArr:  true,
				Producer:    nil,
				Consumer:    nil,
			},
			args{common.TestShortID},
			common.TestURL,
			assert.NoError,
		},
		{
			"Error Arr deleted",
			fields{
				ArrayItems:  []*Item{{common.TestShortID, common.TestURL, true, testUserID}},
				Filename:    common.TestDBName,
				StoreInFile: false,
				StoreInArr:  true,
				Producer:    nil,
				Consumer:    nil,
			},
			args{common.TestShortID},
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
				ArrayItems:  tt.fields.ArrayItems,
				StoreInArr:  tt.fields.StoreInArr,
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
		ArrayItems  []*Item
		PGConn      *sql.DB
		Filename    string
		StoreInFile bool
		StoreInPg   bool
		StoreInArr  bool
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
		{
			name: "OK IN array",
			fields: fields{
				Items:       map[string]string{},
				UserItems:   map[string][]*UserItem{},
				PGConn:      nil,
				Filename:    "",
				StoreInFile: false,
				StoreInPg:   false,
				StoreInArr:  true,
				Producer:    nil,
				Consumer:    nil,
			},
			args: args{
				item:   common.TestURL,
				userID: testUserID,
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
				StoreInArr:  tt.fields.StoreInArr,
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
	defer conn.Close()

	rows := sqlmock.NewRows([]string{"shortURL", "insertResult"}).AddRow(common.TestShortID, PostgresSQLSuccessful)
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
	defer conn.Close()

	rows := sqlmock.NewRows([]string{"shortURL", "insertResult"}).AddRow(common.TestShortID, PostgresSQLDuplicate)
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

func TestGetUrlPostgres(t *testing.T) {
	type args struct {
		id string
	}
	tests := []struct {
		name     string
		args     args
		initMock func(sqlmock.Sqlmock) sqlmock.Sqlmock
		want     string
		wantErr  assert.ErrorAssertionFunc
	}{
		{
			name: "OK",
			args: args{id: common.TestShortID},
			initMock: func(mock sqlmock.Sqlmock) sqlmock.Sqlmock {
				rows := sqlmock.NewRows([]string{"url", "deleted"}).AddRow(common.TestURL, false)
				mock.ExpectQuery(regexp.QuoteMeta(getOriginalURLStmt)).WithArgs(common.TestShortID).WillReturnRows(rows)
				return mock
			},
			want:    common.TestURL,
			wantErr: assert.NoError,
		},
		{
			name: "Error. Deleted",
			args: args{id: common.TestShortID},
			initMock: func(mock sqlmock.Sqlmock) sqlmock.Sqlmock {
				rows := sqlmock.NewRows([]string{"url", "deleted"}).AddRow(common.TestURL, true)
				mock.ExpectQuery(regexp.QuoteMeta(getOriginalURLStmt)).WithArgs(common.TestShortID).WillReturnRows(rows)
				return mock
			},
			want:    "",
			wantErr: assert.Error,
		},
		{
			name: "Query error",
			args: args{id: common.TestShortID},
			initMock: func(mock sqlmock.Sqlmock) sqlmock.Sqlmock {
				mock.ExpectQuery(regexp.QuoteMeta(getOriginalURLStmt)).WithArgs(common.TestShortID).WillReturnError(errors.New("test error"))
				return mock
			},
			want:    "",
			wantErr: assert.Error,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testDB, mock := setupTestDatabase(t)
			defer testDB.PGConn.Close()

			tt.initMock(mock)

			got, err := testDB.Get(tt.args.id)

			if !tt.wantErr(t, err, fmt.Sprintf("Get(%v)", tt.args.id)) {
				return
			}
			assert.Equalf(t, tt.want, got, "Get(%v, %v)", tt.args.id)

			// we make sure that all expectations were met
			if err = mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}

		})
	}

}

func TestGetUserUrlPostgres(t *testing.T) {
	type args struct {
		userID string
	}
	tests := []struct {
		name     string
		args     args
		initMock func(sqlmock.Sqlmock) sqlmock.Sqlmock
		want     []*UserItem
		wantErr  assert.ErrorAssertionFunc
	}{
		{
			name: "OK",
			args: args{userID: testUserID},
			initMock: func(mock sqlmock.Sqlmock) sqlmock.Sqlmock {
				rows := sqlmock.NewRows([]string{"OriginalURL", "ShortURL"}).AddRow(common.TestURL, common.TestShortID)
				mock.ExpectQuery(regexp.QuoteMeta(getUserURL)).WithArgs(testUserID).WillReturnRows(rows)
				return mock
			},
			want: []*UserItem{{
				ShortURL:    common.TestShortID,
				OriginalURL: common.TestURL,
			}},
			wantErr: assert.NoError,
		},
		{
			name: "Error. Rows scan error",
			args: args{userID: testUserID},
			initMock: func(mock sqlmock.Sqlmock) sqlmock.Sqlmock {
				rows := sqlmock.NewRows([]string{"test"}).AddRow("test")
				mock.ExpectQuery(regexp.QuoteMeta(getUserURL)).WithArgs(testUserID).WillReturnRows(rows)
				return mock
			},
			want:    nil,
			wantErr: assert.Error,
		},
		{
			name: "Error. Query error",
			args: args{userID: testUserID},
			initMock: func(mock sqlmock.Sqlmock) sqlmock.Sqlmock {
				mock.ExpectQuery(regexp.QuoteMeta(getUserURL)).WithArgs(testUserID).WillReturnError(fmt.Errorf("some error"))
				return mock
			},
			want:    nil,
			wantErr: assert.Error,
		},
		{
			name: "Row error",
			args: args{userID: testUserID},
			initMock: func(mock sqlmock.Sqlmock) sqlmock.Sqlmock {
				rows := sqlmock.NewRows([]string{"OriginalURL", "ShortURL"}).AddRow(common.TestURL, common.TestShortID)
				rows.RowError(0, errors.New("test error"))
				mock.ExpectQuery(regexp.QuoteMeta(getUserURL)).WithArgs(testUserID).WillReturnRows(rows)
				return mock
			},
			want:    nil,
			wantErr: assert.Error,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testDB, mock := setupTestDatabase(t)
			defer testDB.PGConn.Close()

			tt.initMock(mock)

			got, err := testDB.GetUserURL(tt.args.userID)

			if !tt.wantErr(t, err, "GetUserURL()") {
				return
			}
			assert.Equalf(t, tt.want, got, "GetUserURL()")

			// we make sure that all expectations were met
			if err = mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}

}

func TestCreateItemsPostgres(t *testing.T) {
	type args struct {
		items  []BatchItemRequest
		userID string
	}
	tests := []struct {
		name     string
		args     args
		initMock func(sqlmock.Sqlmock) sqlmock.Sqlmock
		want     []BatchItemResponse
		wantErr  assert.ErrorAssertionFunc
	}{
		{
			name: "OK",
			args: args{userID: testUserID, items: []BatchItemRequest{{
				CorrelationID: testItemID,
				ShortURL:      common.TestShortID,
				OriginalURL:   common.TestURL,
			}}},
			initMock: func(mock sqlmock.Sqlmock) sqlmock.Sqlmock {
				mock.ExpectBegin()
				mock.ExpectPrepare(regexp.QuoteMeta(batchInsert)).ExpectExec().WithArgs(testItemID, common.TestShortID, common.TestURL, testUserID).WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
				return mock
			},
			want: []BatchItemResponse{{
				CorrelationID: testItemID,
				ShortURL:      "/" + common.TestShortID,
			}},
			wantErr: assert.NoError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testDB, mock := setupTestDatabase(t)
			defer testDB.PGConn.Close()

			tt.initMock(mock)

			got, err := testDB.CreateItems(tt.args.items, tt.args.userID)

			if !tt.wantErr(t, err, fmt.Sprintf("Create items(%v, %v)", tt.args.userID, tt.args.items)) {
				return
			}
			assert.Equalf(t, tt.want, got, "Create items(%v, %v)", tt.args.userID, tt.args.items)

			// we make sure that all expectations were met
			if err = mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}

		})
	}
}

func BenchmarkSaveURL_Arr(b *testing.B) {

	var db = &Database{
		Items:       map[string]string{},
		UserItems:   map[string][]*UserItem{},
		ArrayItems:  []*Item{},
		PGConn:      nil,
		Filename:    "",
		StoreInFile: false,
		StoreInArr:  true,
		Producer:    nil,
		Consumer:    nil,
	} // now we execute our method

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		db.Insert(common.TestURL, testUserID)
	}
}

func BenchmarkSaveURL_Map(b *testing.B) {

	var db = &Database{
		Items:       map[string]string{},
		UserItems:   map[string][]*UserItem{},
		ArrayItems:  []*Item{},
		PGConn:      nil,
		Filename:    "",
		StoreInFile: false,
		StoreInArr:  false,
		Producer:    nil,
		Consumer:    nil,
	} // now we execute our method

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		db.Insert(common.TestURL, testUserID)
	}
}

func TestUpdateItemsPostgres(t *testing.T) {
	type args struct {
		itemsIDs []string
	}
	tests := []struct {
		name     string
		args     args
		initMock func(sqlmock.Sqlmock) sqlmock.Sqlmock
		wantErr  assert.ErrorAssertionFunc
	}{
		{
			name: "OK",
			args: args{itemsIDs: []string{testItemID}},
			initMock: func(mock sqlmock.Sqlmock) sqlmock.Sqlmock {
				mock.ExpectExec(
					regexp.QuoteMeta(
						"UPDATE short_url SET deleted = true FROM ( VALUES " + fmt.Sprintf("('%s')", testItemID) + ") AS update_values (shortURL) WHERE short_url.short_url = update_values.shortURL;")).WillReturnResult(sqlmock.NewResult(1, 1))
				return mock
			},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testDB, mock := setupTestDatabase(t)
			defer testDB.PGConn.Close()

			tt.initMock(mock)

			err := testDB.UpdateItems(tt.args.itemsIDs)

			if !tt.wantErr(t, err, "UpdateItems()") {
				return
			}

			// we make sure that all expectations were met
			if err = mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestInitDB(t *testing.T) {
	tests := []struct {
		name    string
		want    *Database
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "OK",
			want: &Database{
				Items:       map[string]string{},
				UserItems:   map[string][]*UserItem{},
				ArrayItems:  make([]*Item, 0),
				PGConn:      nil,
				Filename:    "",
				StoreInFile: false,
				StoreInPg:   false,
				StoreInArr:  false,
				Producer:    nil,
				Consumer:    nil,
			},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := InitDB()
			if !tt.wantErr(t, err, "InitDB()") {
				return
			}
			assert.Equalf(t, tt.want, got, "InitDB()")
		})
	}
}
