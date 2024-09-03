package engine

import (
	"context"
	"testing"

	"github.com/jaxmef/datapipe/config"

	"github.com/stretchr/testify/assert"
)

func TestFilter_Name(t *testing.T) {
	name := "test-filter"
	cfg := config.FilterHandler{
		Expression:  "",
		ExpectFalse: false,
	}
	f := newFilterHandler(name, cfg)

	assert.Equal(t, name, f.Name())
}

func TestFilter_Handle(t *testing.T) {
	tests := []struct {
		name      string
		cfg       config.FilterHandler
		data      map[string]string
		expectErr bool
		expectRes []HandlerResult
	}{
		{
			name: "successful boolean evaluation - true",
			cfg: config.FilterHandler{
				Expression:  "2 > 1",
				ExpectFalse: false,
			},
			data:      map[string]string{},
			expectErr: false,
			expectRes: []HandlerResult{{}},
		},
		{
			name: "successful boolean evaluation - false",
			cfg: config.FilterHandler{
				Expression:  "2 < 1",
				ExpectFalse: true,
			},
			data:      map[string]string{},
			expectErr: false,
			expectRes: []HandlerResult{{}},
		},
		{
			name: "expression evaluates to false without ExpectFalse",
			cfg: config.FilterHandler{
				Expression:  "2 < 1",
				ExpectFalse: false,
			},
			data:      map[string]string{},
			expectErr: false,
			expectRes: nil,
		},
		{
			name: "expression evaluates to true with ExpectFalse",
			cfg: config.FilterHandler{
				Expression:  "2 > 1",
				ExpectFalse: true,
			},
			data:      map[string]string{},
			expectErr: false,
			expectRes: nil,
		},
		{
			name: "failed placeholder replacement",
			cfg: config.FilterHandler{
				Expression:  "{{missing_key}}",
				ExpectFalse: false,
			},
			data:      map[string]string{},
			expectErr: true,
			expectRes: nil,
		},
		{
			name: "non-boolean expression result",
			cfg: config.FilterHandler{
				Expression:  `"string_result"`,
				ExpectFalse: false,
			},
			data:      map[string]string{},
			expectErr: true,
			expectRes: nil,
		},
		{
			name: "success with placeholder replacement",
			cfg: config.FilterHandler{
				Expression:  `{{key}} == 12`,
				ExpectFalse: false,
			},
			data: map[string]string{
				"key": "12",
			},
			expectErr: false,
			expectRes: []HandlerResult{{}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newFilterHandler("test-filter", tt.cfg)
			ctx := context.Background()

			res, err := f.Handle(ctx, tt.data)

			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.expectRes, res)
		})
	}
}
