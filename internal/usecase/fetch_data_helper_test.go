package usecase

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseUrl(t *testing.T) {
	t.Parallel()
	type args struct {
		ctx    context.Context
		rawUrl string
	}
	tests := []struct {
		name     string
		args     args
		wantHost string
		wantPath string
	}{
		{
			name: "Test ParseUrl",
			args: args{
				ctx:    context.Background(),
				rawUrl: "https://gateway.chotot.org:443/v1/private/bank_transfer/contract-history/1282?limit=20&page=0",
			},
			wantHost: "gateway.chotot.org",
			wantPath: "/v1/private/bank_transfer/contract-history/1282",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			host, path := parseRawUrl(tt.args.ctx, tt.args.rawUrl)
			require.Equal(t, tt.wantHost, host)
			require.Equal(t, tt.wantPath, path)
		})
	}
}

func TestFindParameterInPath(t *testing.T) {
	t.Parallel()
	type args struct {
		ctx  context.Context
		path string
	}
	tests := []struct {
		name        string
		args        args
		wantUpdated string
	}{
		{
			name: "Test FindParameterInPath - id",
			args: args{
				ctx:  context.Background(),
				path: "/v1/private/bank_transfer/contract-history/1282",
			},
			wantUpdated: "/v1/private/bank_transfer/contract-history/{id}",
		},
		{
			name: "Test FindParameterInPath - uuid",
			args: args{
				ctx:  context.Background(),
				path: "/v1/private/bank_transfer/contract-history/123e4567-e89b-12d3-a456-426614174000",
			},
			wantUpdated: "/v1/private/bank_transfer/contract-history/{uuid}",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			updated := findParameterInPath(tt.args.ctx, tt.args.path)
			require.Equal(t, tt.wantUpdated, updated)
		})
	}
}
