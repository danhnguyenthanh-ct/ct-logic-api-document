package gcsutils

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestGcsUtils_GetFolderPath(t *testing.T) {
	t.Parallel()
	type args struct {
		ctx       context.Context
		startTime time.Time
		endTime   time.Time
	}
	tests := []struct {
		name    string
		args    args
		want    []string
		wantErr bool
	}{
		{
			name: "Test GetFolderPath",
			args: args{
				ctx:       context.Background(),
				startTime: time.Unix(1733108400, 0).UTC(), // 2024-12-02 03:00:00
				endTime:   time.Unix(1733112000, 0).UTC(), // 2024-12-02 04:00:00
			},
			want: []string{
				"logs/proxy/2024-12-02/03/",
				"logs/proxy/2024-12-02/04/",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			folderPaths, err := GetFolderPath(tt.args.ctx, tt.args.startTime, tt.args.endTime)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.Equal(t, len(tt.want), len(folderPaths))
				require.Equal(t, tt.want, folderPaths)
			}
		})
	}
}
