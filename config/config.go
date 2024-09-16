package config

import (
	"fmt"
	"io"
	"os"

	yaml "gopkg.in/yaml.v3"
)

type Config struct {
	Engine   Engine      `yaml:"engine"`
	Handlers *HandlerMap `yaml:"handlers"`
}

func (c *Config) Validate() error {
	if c.Handlers == nil || len(*c.Handlers) == 0 {
		return fmt.Errorf("no handlers defined")
	}
	if err := c.Engine.Validate(); err != nil {
		return fmt.Errorf("invalid engine config: %s", err)
	}
	handlerNames := make(map[string]struct{})
	for _, handlerItem := range *c.Handlers {
		if _, ok := handlerNames[handlerItem.Name]; ok {
			return fmt.Errorf("duplicate handler name: '%s'", handlerItem.Name)
		}
		handlerNames[handlerItem.Name] = struct{}{}

		if err := handlerItem.Handler.Validate(); err != nil {
			return fmt.Errorf("config for '%s' handler is invalid: %s", handlerItem.Name, err)
		}
	}
	return nil
}

func NewConfig() *Config {
	return &Config{}
}

func (c *Config) ParseFromYamlFile(filePath string) error {
	yamlData, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open yaml config file: %s", err)
	}
	defer yamlData.Close()

	rawYamlData, err := io.ReadAll(yamlData)
	if err != nil {
		return fmt.Errorf("failed to read yaml config file: %s", err)
	}

	err = c.ParseFromYaml(rawYamlData)
	if err != nil {
		return fmt.Errorf("failed to parse yaml: %s", err)
	}

	return nil
}

func (c *Config) ParseFromYaml(yamlData []byte) error {
	err := yaml.Unmarshal(yamlData, c)
	if err != nil {
		return err
	}
	return nil
}
