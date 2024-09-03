package config

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/expr-lang/expr"
	"gopkg.in/yaml.v3"
)

type HandlerType string

const (
	HandlerTypeHTTP   HandlerType = "http"
	HandlerTypeFilter HandlerType = "filter"
)

type Handler struct {
	Type HandlerType `yaml:"type"`

	HTTPHandler   HTTPHandler   `yaml:"http"`
	FilterHandler FilterHandler `yaml:"filter"`
}

type HTTPHandler struct {
	URL                  string            `yaml:"url"`
	Method               string            `yaml:"method"`
	Body                 string            `yaml:"body"`
	Headers              map[string]string `yaml:"headers"`
	QueryParams          map[string]string `yaml:"query_params"`
	Timeout              time.Duration     `yaml:"timeout"`
	ExpectedResponseCode int               `yaml:"expected_response_code"`
	Retries              int               `yaml:"retries"`
	RetryInterval        time.Duration     `yaml:"retry_interval"`
	ParallelRun          bool              `yaml:"parallel_run"`
}

func (h HTTPHandler) Validate() error {
	if h.Method == "" {
		return fmt.Errorf("'method' is required")
	}
	if h.URL == "" {
		return fmt.Errorf("'url' is required")
	}
	return nil
}

type FilterHandler struct {
	ExpectFalse bool   `yaml:"expect_false"`
	Expression  string `yaml:"expression"`
}

func (h FilterHandler) Validate() error {
	if h.Expression == "" {
		return fmt.Errorf("'expression' is required")
	}
	_, err := expr.Eval(
		replacePlaceholdersForValidation(h.Expression),
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to parse expression: %s", err)
	}
	return nil
}

func replacePlaceholdersForValidation(e string) string {
	re := regexp.MustCompile(`\{\{\s*([^\s}]+)\s*\}\}`)

	return re.ReplaceAllStringFunc(e, func(m string) string {
		key := strings.TrimSpace(m[2 : len(m)-2])
		return fmt.Sprintf(`"%s"`, key)
	})
}

func (h Handler) Validate() error {
	switch h.Type {
	case HandlerTypeHTTP, "":
		return h.HTTPHandler.Validate()
	case HandlerTypeFilter:
		return h.FilterHandler.Validate()
	default:
		return fmt.Errorf("invalid 'type' value: %s", h.Type)
	}
}

// HandlerMap is a list of HandlerMapItem. It is used to guarantee the order of the handlers.
type HandlerMap []HandlerMapItem

func (hm *HandlerMap) UnmarshalYAML(node *yaml.Node) error {
	*hm = HandlerMap{}
	if node.Kind != yaml.MappingNode {
		return fmt.Errorf("HandlerMap must be a mapping node")
	}

	for i := 0; i < len(node.Content); i += 2 {
		keyNode := node.Content[i]
		valueNode := node.Content[i+1]

		handler := Handler{}
		err := valueNode.Decode(&handler)
		if err != nil {
			return fmt.Errorf("failed to decode handler: %s", err)
		}

		*hm = append(*hm, HandlerMapItem{
			Name:    keyNode.Value,
			Handler: handler,
		})
	}
	return nil
}

type HandlerMapItem struct {
	Name    string
	Handler Handler
}
