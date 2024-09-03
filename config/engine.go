package config

import (
	"fmt"
	"time"

	"github.com/rs/zerolog"
)

type Engine struct {
	DisableRunOnStart bool          `yaml:"disable_run_on_start"`
	Interval          time.Duration `yaml:"interval"`
	RunAt             string        `yaml:"run_at"`

	Log Log `yaml:"log"`
}

func (e Engine) Validate() error {
	if e.Interval <= 0 {
		return fmt.Errorf("'interval' must be greater than 0")
	}
	return nil
}

type Log struct {
	Level        LogLevel          `yaml:"level"`
	StaticFields map[string]string `yaml:"static_fields"`
}

type LogLevel string

const (
	LogLevelDebug    LogLevel = "debug"
	LogLevelInfo     LogLevel = "info"
	LogLevelWarn     LogLevel = "warn"
	LogLevelError    LogLevel = "error"
	LogLevelDisabled LogLevel = "disabled"
)

func (l LogLevel) ToZerolog() zerolog.Level {
	switch l {
	case LogLevelDebug:
		return zerolog.DebugLevel
	case LogLevelInfo:
		return zerolog.InfoLevel
	case LogLevelWarn:
		return zerolog.WarnLevel
	case LogLevelError:
		return zerolog.ErrorLevel
	case LogLevelDisabled:
		return zerolog.Disabled
	default:
		return zerolog.InfoLevel
	}
}
