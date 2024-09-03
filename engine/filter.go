package engine

import (
	"context"
	"fmt"

	"github.com/jaxmef/datapipe/config"

	"github.com/expr-lang/expr"
)

type filterHandler struct {
	name string
	cfg  config.FilterHandler
}

func newFilterHandler(name string, cfg config.FilterHandler) Handler {
	return &filterHandler{
		name: name,
		cfg:  cfg,
	}
}

func (f *filterHandler) Name() string {
	return f.name
}

func (f *filterHandler) Handle(ctx context.Context, data map[string]string) ([]HandlerResult, error) {
	expression, ok := replacePlaceholders(f.cfg.Expression, data)
	if !ok {
		return nil, fmt.Errorf("failed to replace placeholders in expression: some data not found")
	}

	expressionResult, err := expr.Eval(expression, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate expression: %s", err)
	}

	expressionBoolResult, ok := expressionResult.(bool)
	if !ok {
		return nil, fmt.Errorf("expression did not return a boolean")
	}

	if expressionBoolResult == f.cfg.ExpectFalse {
		return nil, nil
	}

	return []HandlerResult{{}}, nil
}
