package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"

	"gorm.io/gorm"
	"law-oa-go/internal/repositories"
	"law-oa-go/internal/services"
)

func (h *DemoAggregateHandler) intakeWriteTransaction(tx *gorm.DB) *DemoAggregateHandler {
	bound := *h
	bound.db = tx
	if h.authz != nil {
		bound.authz = services.NewAuthorizationService(repositories.NewCaseRepository(tx), repositories.NewClientRepository(tx), repositories.NewUserRepository(tx), repositories.NewDocumentRepository(tx))
	}
	return &bound
}

func decodeIntakeMetadata(raw interface{}) (map[string]interface{}, error) {
	if raw == nil {
		return map[string]interface{}{}, nil
	}
	var data []byte
	switch v := raw.(type) {
	case map[string]interface{}:
		return v, nil
	case []byte:
		data = v
	case string:
		data = []byte(v)
	default:
		return nil, fmt.Errorf("无效的接案 metadata 类型 %T", raw)
	}
	var result map[string]interface{}
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()
	if !json.Valid(data) || dec.Decode(&result) != nil || result == nil {
		return nil, errors.New("接案 metadata 必须为 JSON 对象")
	}
	return result, nil
}

// A full form may echo a stored JSON number after JSON.parse rounded it.
// Only the exact wire representation of that rounded number is ambiguous.
func preserveBrowserNumbers(stored, incoming any) any {
	switch old := stored.(type) {
	case json.Number:
		if n, ok := incoming.(json.Number); ok && n != old {
			if f, err := old.Float64(); err == nil {
				rounded, _ := json.Marshal(f)
				if string(rounded) == n.String() {
					return old
				}
			}
		}
	case map[string]any:
		if patch, ok := incoming.(map[string]any); ok {
			for key, value := range patch {
				patch[key] = preserveBrowserNumbers(old[key], value)
			}
			// Keep a redacted object's hidden fields on an unchanged echo.
			if reflect.DeepEqual(patch, redactAggregateSensitiveValue(old)) {
				return old
			}
		}
	case []any:
		if values, ok := incoming.([]any); ok && len(values) == len(old) {
			for i, value := range values {
				values[i] = preserveBrowserNumbers(old[i], value)
			}
		}
	}
	return incoming
}

const (
	lawyerFactsVersionKey          = "lawyer_facts_version"
	npDescriptionSourceRevisionKey = "np_description_source_revision"
)
