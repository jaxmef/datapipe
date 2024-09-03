package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jaxmef/datapipe/config"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockHandler struct {
	handle func(ctx context.Context, data map[string]string) ([]HandlerResult, error)
}

func (m *mockHandler) Name() string {
	return "mock"
}

func (m *mockHandler) Handle(ctx context.Context, data map[string]string) ([]HandlerResult, error) {
	return m.handle(ctx, data)
}

func TestNewDataPipe_NoHandlers(t *testing.T) {
	cfg := config.Config{
		Handlers: nil,
	}

	dp, err := NewDataPipe(cfg, zerolog.New(os.Stdout))

	assert.Nil(t, dp)
	assert.EqualError(t, err, "no handlers defined")
}

func TestNewDataPipe_Success(t *testing.T) {
	cfg := config.Config{
		Handlers: &config.HandlerMap{
			{
				Name: "handler1",
				Handler: config.Handler{
					HTTPHandler: config.HTTPHandler{
						URL: "http://example.com",
					},
				},
			},
		},
	}

	dp, err := NewDataPipe(cfg, zerolog.New(os.Stdout))

	assert.NotNil(t, dp)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(dp.(*dataPipe).handlers))
	assert.Equal(t, "handler1", dp.(*dataPipe).handlers[0].Name())
}

func TestDataPipeRun_RunOnStart(t *testing.T) {
	handlerCalls := 0
	mockHandler := &mockHandler{
		handle: func(ctx context.Context, data map[string]string) ([]HandlerResult, error) {
			handlerCalls++
			return []HandlerResult{{"key": json.RawMessage("value")}}, nil
		},
	}

	cfg := config.Config{
		Engine: config.Engine{
			DisableRunOnStart: false,
			Interval:          time.Minute,
		},
		Handlers: &config.HandlerMap{
			{
				Name:    "mock",
				Handler: config.Handler{},
			},
		},
	}

	dp := &dataPipe{
		cfg:      cfg.Engine,
		handlers: []Handler{mockHandler},
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		dp.Run(ctx)
	}()

	// allow some time for the job to start
	time.Sleep(200 * time.Millisecond)
	cancel()

	assert.Equal(t, 1, handlerCalls)
}

func TestDataPipeRun_CancelContext(t *testing.T) {
	mockHandler := &mockHandler{
		handle: func(ctx context.Context, data map[string]string) ([]HandlerResult, error) {
			assert.Fail(t, "handler should not be called")
			return []HandlerResult{{"key": json.RawMessage("value")}}, nil
		},
	}

	cfg := config.Config{
		Engine: config.Engine{
			DisableRunOnStart: true,
			Interval:          time.Minute,
		},
		Handlers: &config.HandlerMap{
			{
				Name:    "mock",
				Handler: config.Handler{},
			},
		},
	}

	dp := &dataPipe{
		cfg:      cfg.Engine,
		handlers: []Handler{mockHandler},
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	runFinished := make(chan struct{})
	go func() {
		dp.Run(ctx)
		close(runFinished)
	}()

	// allow some time for the ticker to start
	time.Sleep(200 * time.Millisecond)
	cancel()

	_, ok := <-runFinished
	assert.False(t, ok)
}

func TestDataPipeRunJob_Success(t *testing.T) {
	handlerCalls := 0
	mockHandler := &mockHandler{
		handle: func(ctx context.Context, data map[string]string) ([]HandlerResult, error) {
			handlerCalls++
			return []HandlerResult{{"key": json.RawMessage("value")}}, nil
		},
	}

	dp := &dataPipe{
		handlers: []Handler{mockHandler},
	}

	err := dp.runJob(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, 1, handlerCalls)
}

func TestDataPipeRunJob_HandlerError(t *testing.T) {
	e := fmt.Errorf("handler error")
	mockHandler := &mockHandler{
		handle: func(ctx context.Context, data map[string]string) ([]HandlerResult, error) {
			return nil, e
		},
	}

	dp := &dataPipe{
		handlers: []Handler{mockHandler},
	}

	err := dp.runJob(context.Background())

	assert.ErrorContains(t, err, e.Error())
}

func TestRunHandlerPipe_Parallel(t *testing.T) {
	handler1Calls := 0
	handler1 := &mockHandler{
		handle: func(ctx context.Context, data map[string]string) ([]HandlerResult, error) {
			handler1Calls++
			return []HandlerResult{{"key": json.RawMessage("value")}, {"key": json.RawMessage("value")}}, nil
		},
	}

	handler2Calls := 0
	runningHandlers2 := atomic.Int32{}
	parallelRunDetected := false
	timer := time.NewTimer(100 * time.Millisecond)
	handler2 := &mockHandler{
		handle: func(ctx context.Context, data map[string]string) ([]HandlerResult, error) {
			runningHandlers2.Add(1)
			defer runningHandlers2.Add(-1)

			handler2Calls++

			for {
				select {
				case <-timer.C:
					if !parallelRunDetected {
						require.Fail(t, "no parallel run detected")
					}
					return nil, nil
				default:
					if runningHandlers2.Load() > 1 {
						parallelRunDetected = true
						return nil, nil
					}
				}
			}
		},
	}

	err := runHandlerPipe(context.Background(), map[string]string{}, []Handler{handler1, handler2})
	assert.NoError(t, err)

	assert.Equal(t, 1, handler1Calls)
	assert.Equal(t, 2, handler2Calls)
	assert.Equal(t, int32(0), runningHandlers2.Load())
	assert.True(t, parallelRunDetected)
}
