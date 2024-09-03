package engine

import (
	"encoding/json"
)

type HandlerResult map[string]json.RawMessage

type handlerResponseBody struct {
	Results []HandlerResult `json:"results"`
}
