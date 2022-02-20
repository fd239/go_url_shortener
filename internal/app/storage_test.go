package app

import (
	"fmt"
	"github.com/fd239/go_url_shortener/internal/app/common"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSaveShortRoute(t *testing.T) {
	tests := []struct {
		name string
		args string
		want string
	}{
		{"OK", common.TestUrl, common.TestShortId},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, SaveShortRoute(tt.args), "SaveShortRoute(%s)", tt.args)
		})
	}
}

func TestGetShortRoute(t *testing.T) {
	tests := []struct {
		name    string
		routeId string
		want    string
		wantErr assert.ErrorAssertionFunc
	}{
		{"OK", common.TestShortId, common.TestUrl, assert.NoError},
		{"With error", "123", "", assert.Error},
	}
	urlMap[common.TestShortId] = common.TestUrl
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetShortRoute(tt.routeId)
			if !tt.wantErr(t, err, fmt.Sprintf("GetShortRoute(%v)", tt.routeId)) {
				return
			}
			assert.Equalf(t, tt.want, got, "GetShortRoute(%v)", tt.routeId)
		})
	}
}
