package engine

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/jaxmef/datapipe/config"

	"github.com/rs/zerolog"
)

type DataPipe interface {
	Run(ctx context.Context)
}

type dataPipe struct {
	cfg      config.Engine
	handlers []Handler
	logger   zerolog.Logger
}

func NewDataPipe(cfg config.Config, logger zerolog.Logger) (DataPipe, error) {
	dp := &dataPipe{
		cfg:    cfg.Engine,
		logger: logger,
	}

	if cfg.Handlers == nil || len(*cfg.Handlers) == 0 {
		return nil, fmt.Errorf("no handlers defined")
	}
	for _, handlerItem := range *cfg.Handlers {
		h, err := newHandler(handlerItem.Name, handlerItem.Handler)
		if err != nil {
			return nil, fmt.Errorf("failed to create '%s' handler: %s", handlerItem.Name, err)
		}
		dp.handlers = append(dp.handlers, h)
	}

	return dp, nil
}

func (dp *dataPipe) Run(ctx context.Context) {
	if !dp.cfg.DisableRunOnStart {
		err := dp.runJob(ctx)
		if err != nil {
			dp.logger.Error().Msgf("failed to run job: %s", err)
		}
	}

	t, interval, err := createTimer(dp.cfg.Interval, dp.cfg.RunAt)
	if err != nil {
		dp.logger.Error().Msgf("failed to create timer: %s", err)
		return
	}

	for {
		select {
		case <-ctx.Done():
			dp.logger.Info().Msg("data pipe stopped")
			return
		case <-t.C:
			err := dp.runJob(ctx)
			if err != nil {
				dp.logger.Error().Err(err).Msg("failed to run job")
			} else {
				dp.logger.Info().Msg("job completed successfully")
			}
			t.Reset(interval)
		}
	}
}

func (dp *dataPipe) runJob(ctx context.Context) error {
	err := runHandlerPipe(ctx, nil, dp.handlers)
	if err != nil {
		return fmt.Errorf("failed to run handler pipe: %s", err)
	}
	return nil
}

func runHandlerPipe(ctx context.Context, data map[string]string, handlers []Handler) error {
	if len(handlers) == 0 {
		return nil
	}

	results, err := handlers[0].Handle(ctx, data)
	if err != nil {
		return fmt.Errorf("failed to run handler %s: %s", handlers[0].Name(), err)
	}

	wg := sync.WaitGroup{}
	errsChan := make(chan error, len(results))
	for i := 0; i < len(results); i++ {
		newData := copyMap(data)
		for k, v := range results[i] {
			newData[handlers[0].Name()+"."+k] = string(v)
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := runHandlerPipe(ctx, newData, handlers[1:])
			if err != nil {
				errsChan <- err
			}
		}()
	}

	wg.Wait()
	close(errsChan)

	errMsg := ""
	for err := range errsChan {
		if err != nil {
			errMsg += err.Error() + "\n"
		}
	}
	if errMsg != "" {
		return fmt.Errorf("failed to run handler pipe: %s", errMsg)
	}

	return nil
}

func copyMap(originalMap map[string]string) map[string]string {
	newMap := make(map[string]string, len(originalMap))

	for key, value := range originalMap {
		newMap[key] = value
	}

	return newMap
}

func createTimer(interval time.Duration, runAt string) (*time.Timer, time.Duration, error) {
	t := time.NewTimer(interval)
	if runAt != "" {
		runAtTime, err := time.Parse("15:04", runAt)
		if err != nil {
			return nil, interval, fmt.Errorf("failed to parse run_at time: %s", err)
		}
		now := time.Now()
		nextRun := time.Date(now.Year(), now.Month(), now.Day(), runAtTime.Hour(), runAtTime.Minute(), 0, 0, now.Location())
		if nextRun.Before(now) {
			nextRun = nextRun.Add(24 * time.Hour)
		}
		t.Reset(nextRun.Sub(now))
		interval = 24 * time.Hour
	}
	return t, interval, nil
}
