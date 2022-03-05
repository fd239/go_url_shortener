package storage

import (
	"fmt"
	"github.com/fd239/go_url_shortener/internal/app/common"
	"github.com/stretchr/testify/assert"
	"log"
	"testing"
)

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
				Items:       map[string]string{common.TestURL: common.TestShortId},
				Filename:    common.TestDBName,
				StoreInFile: true,
				Producer:    getProducer(),
				Consumer:    getConsumer(),
			},
			args{common.TestURL},
			common.TestShortId,
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
				Items:       map[string]string{common.TestShortId: common.TestURL},
				Filename:    common.TestDBName,
				StoreInFile: true,
				Producer:    getProducer(),
				Consumer:    getConsumer(),
			},
			args{common.TestShortId},
			common.TestURL,
			assert.NoError,
		},
		{
			"Error",
			fields{
				Items:       map[string]string{},
				Filename:    common.TestDBName,
				StoreInFile: true,
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
