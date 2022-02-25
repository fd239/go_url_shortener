package app

import (
	"fmt"
	"github.com/fd239/go_url_shortener/internal/app/common"
	"github.com/stretchr/testify/assert"
	"testing"
)

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
				Items:       map[string]string{},
				Filename:    common.TestDBName,
				StoreInFile: true,
				Producer:    NewProducer(common.TestDBName),
				Consumer:    NewConsumer(common.TestDBName),
			},
			args{common.TestUrl},
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

			got, err := db.SaveShortRoute(tt.args.url)
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
		routeId string
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
				Items:       map[string]string{common.TestShortId: common.TestUrl},
				Filename:    common.TestDBName,
				StoreInFile: true,
				Producer:    NewProducer(common.TestDBName),
				Consumer:    NewConsumer(common.TestDBName),
			},
			args{common.TestShortId},
			common.TestUrl,
			assert.NoError,
		},
		{
			"Error",
			fields{
				Items:       map[string]string{},
				Filename:    common.TestDBName,
				StoreInFile: true,
				Producer:    NewProducer(common.TestDBName),
				Consumer:    NewConsumer(common.TestDBName),
			},
			args{common.TestUrl},
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
			got, err := db.GetShortRoute(tt.args.routeId)
			if !tt.wantErr(t, err, fmt.Sprintf("GetShortRoute(%v)", tt.args.routeId)) {
				return
			}
			assert.Equalf(t, tt.want, got, "GetShortRoute(%v)", tt.args.routeId)
		})
	}
}
