package config

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name        string
		cfg         *Config
		errContains string
	}{
		{
			name: "Valid",
			cfg: &Config{
				Engine: Engine{
					Interval: time.Minute,
				},
				Handlers: &HandlerMap{
					{
						Name: "handler1",
						Handler: Handler{
							HTTPHandler: HTTPHandler{
								Method: "POST",
								URL:    "http://example.com",
							},
						},
					},
				},
			},
		},
		{
			name: "NoHandlers",
			cfg: &Config{
				Engine: Engine{
					Interval: time.Minute,
				},
			},
			errContains: "no handlers defined",
		},
		{
			name: "InvalidEngine",
			cfg: &Config{
				Handlers: &HandlerMap{
					{
						Name: "handler1",
						Handler: Handler{
							HTTPHandler: HTTPHandler{
								Method: "POST",
								URL:    "http://example.com",
							},
						},
					},
				},
			},
			errContains: "invalid engine config",
		},
		{
			name: "InvalidHandler",
			cfg: &Config{
				Engine: Engine{
					Interval: time.Minute,
				},
				Handlers: &HandlerMap{
					{
						Name: "handler1",
						Handler: Handler{
							HTTPHandler: HTTPHandler{
								Method: "POST",
							},
						},
					},
				},
			},
			errContains: "config for 'handler1' handler is invalid",
		},
		{
			name: "DuplicateHandlerName",
			cfg: &Config{
				Engine: Engine{
					Interval: time.Minute,
				},
				Handlers: &HandlerMap{
					{
						Name: "handler1",
						Handler: Handler{
							HTTPHandler: HTTPHandler{
								Method: "POST",
								URL:    "http://example.com",
							},
						},
					},
					{
						Name: "handler1",
						Handler: Handler{
							HTTPHandler: HTTPHandler{
								Method: "POST",
								URL:    "http://example.com",
							},
						},
					},
				},
			},
			errContains: "duplicate handler name: 'handler1'",
		},
		{
			name: "EmptyFilterExpression",
			cfg: &Config{
				Engine: Engine{
					Interval: time.Minute,
				},
				Handlers: &HandlerMap{
					{
						Name: "filter1",
						Handler: Handler{
							Type: HandlerTypeFilter,
							FilterHandler: FilterHandler{
								Expression: "",
							},
						},
					},
				},
			},
			errContains: "'expression' is required",
		},
		{
			name: "InvalidFilterExpression",
			cfg: &Config{
				Engine: Engine{
					Interval: time.Minute,
				},
				Handlers: &HandlerMap{
					{
						Name: "filter1",
						Handler: Handler{
							Type: HandlerTypeFilter,
							FilterHandler: FilterHandler{
								Expression: "invalid-expression",
							},
						},
					},
				},
			},
			errContains: "failed to parse expression",
		},
		{
			name: "ValidFilter",
			cfg: &Config{
				Engine: Engine{
					Interval: time.Minute,
				},
				Handlers: &HandlerMap{
					{
						Name: "filter1",
						Handler: Handler{
							Type: HandlerTypeFilter,
							FilterHandler: FilterHandler{
								Expression: `{{ key }} == "value"`,
							},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if tt.errContains != "" {
				assert.Error(t, err)
				assert.True(t, strings.Contains(err.Error(), tt.errContains))
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestHandler_Validate(t *testing.T) {
	tests := []struct {
		name        string
		handler     Handler
		errContains string
	}{
		{
			name: "Valid",
			handler: Handler{
				HTTPHandler: HTTPHandler{
					Method: "POST",
					URL:    "http://example.com",
				},
			},
		},
		{
			name: "NoURL",
			handler: Handler{
				HTTPHandler: HTTPHandler{Method: "POST"},
			},
			errContains: "'url' is required",
		},
		{
			name: "NoMethod",
			handler: Handler{
				HTTPHandler: HTTPHandler{URL: "http://example.com"},
			},
			errContains: "'method' is required",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.handler.Validate()
			if tt.errContains != "" {
				assert.Error(t, err)
				assert.True(t, strings.Contains(err.Error(), tt.errContains))
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestConfig_ParseFromYamlFile(t *testing.T) {
	cfg := NewConfig()
	err := cfg.ParseFromYamlFile("../config.example.yaml")
	assert.NoError(t, err)
}
