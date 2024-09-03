package main

import (
	"context"
	"github.com/rs/zerolog"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/jaxmef/datapipe/config"
	"github.com/jaxmef/datapipe/engine"
)

const (
	ConfigFilePathEnvVar  = "CONFIG_FILE_PATH"
	DefaultConfigFilePath = "./config.yaml"
)

func main() {
	configFilePath := os.Getenv(ConfigFilePathEnvVar)
	if configFilePath == "" {
		configFilePath = DefaultConfigFilePath
	}

	cfg := config.NewConfig()
	err := cfg.ParseFromYamlFile(configFilePath)
	if err != nil {
		log.Fatalf("failed to parse config: %s", err)
	}

	err = cfg.Validate()
	if err != nil {
		log.Fatalf("config validation failed: %s", err)
	}

	logger := zerolog.New(os.Stderr).Level(cfg.Engine.Log.Level.ToZerolog())
	for k, v := range cfg.Engine.Log.StaticFields {
		logger = logger.With().Str(k, v).Logger()
	}

	dp, err := engine.NewDataPipe(*cfg, logger)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to create data pipe")
	}

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
		<-stop
		logger.Info().Msg("shutting down")
		cancel()
	}()

	dp.Run(ctx)
}
