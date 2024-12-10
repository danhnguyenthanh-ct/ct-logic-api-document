package buildstructure

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGenerateSchema(t *testing.T) {
	t.Parallel()
	type args struct {
		ctx  context.Context
		body string
	}
	tests := []struct {
		name string
		args args
		want map[string]any
	}{
		{
			name: "Test GenerateSchema - 01",
			args: args{
				ctx:  context.Background(),
				body: `{"name": "John", "age": 30, "city": "New York"}`,
			},
			want: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"name": map[string]any{"type": "string"},
					"age":  map[string]any{"type": "number"},
					"city": map[string]any{"type": "string"},
				},
			},
		},
		{
			name: "Test GenerateSchema - 02",
			args: args{
				ctx:  context.Background(),
				body: `{"name": "John", "age": 30, "city": "New York", "jobs": [{"title": "software engineer", "company": "Chotot"},{"title": "software engineer", "company": "Carousell", "start_at": 2024}]  ,"address": {"street": "Main St", "zip": [12345, 67890]}}`,
			},
			want: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"name": map[string]any{"type": "string"},
					"age":  map[string]any{"type": "number"},
					"city": map[string]any{"type": "string"},
					"jobs": map[string]any{
						"type": "array",
						"items": map[string]any{
							"type": "object",
							"properties": map[string]any{
								"title":    map[string]any{"type": "string"},
								"company":  map[string]any{"type": "string"},
								"start_at": map[string]any{"type": "number"},
							},
						},
					},
					"address": map[string]any{
						"type": "object",
						"properties": map[string]any{
							"street": map[string]any{"type": "string"},
							"zip": map[string]any{
								"type":  "array",
								"items": map[string]any{"type": "number"},
							},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			data := map[string]any{}
			err := json.Unmarshal([]byte(tt.args.body), &data)
			require.NoError(t, err)
			schema := generateSchema(data)
			// schemaJSON, _ := json.MarshalIndent(schema, "", "  ")
			// fmt.Printf("debug - build_structure_helper_test.go line 85 - schemaJSON: %+v\n", string(schemaJSON))
			// wantJSON, _ := json.MarshalIndent(tt.want, "", "  ")
			// fmt.Printf("debug - build_structure_helper_test.go line 87 - wantJSON: %+v\n", string(wantJSON))
			require.Equal(t, tt.want, schema)
		})
	}
}

func TestGenerateSchemaForArray(t *testing.T) {
	t.Parallel()
	type args struct {
		ctx  context.Context
		body string
	}
	tests := []struct {
		name string
		args args
		want map[string]any
	}{
		{
			name: "Test GenerateSchemaForArray - 01",
			args: args{
				ctx:  context.Background(),
				body: `[{ "name": "John", "age": 30, "city": "New York" }, { "name": "Jane", "age": 25, "city": "Los Angeles" }]`,
			},
			want: map[string]any{
				"type": "array",
				"items": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"name": map[string]any{"type": "string"},
						"age":  map[string]any{"type": "number"},
						"city": map[string]any{"type": "string"},
					},
				},
			},
		},
		{
			name: "Test GenerateSchemaForArray - 02",
			args: args{
				ctx:  context.Background(),
				body: `[1,2,3,4,5]`,
			},
			want: map[string]any{
				"type": "array",
				"items": map[string]any{
					"type": "number",
				},
			},
		},
		{
			name: "Test GenerateSchemaForArray - 02",
			args: args{
				ctx:  context.Background(),
				body: `[]`,
			},
			want: map[string]any{
				"type": "array",
				"items": map[string]any{
					"type": "null",
				},
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			data := []any{}
			err := json.Unmarshal([]byte(tt.args.body), &data)
			require.NoError(t, err)
			schema := generateSchemaForArray(data)
			require.Equal(t, tt.want, schema)
		})
	}
}

func TestMergeMaps(t *testing.T) {
	t.Parallel()
}
