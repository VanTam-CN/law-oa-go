package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
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

// Empty denotes the JSON null emitted by JSON.stringify for a non-finite number.
func browserNumber(n json.Number) json.Number {
	f, err := n.Float64()
	if err != nil {
		return ""
	}
	if f == 0 {
		return "0"
	}
	encoded, _ := json.Marshal(f)
	return json.Number(encoded)
}

func visitMetadataNumbers(value any, visit func(json.Number)) {
	switch v := value.(type) {
	case nil:
		visit("")
	case json.Number:
		visit(v)
	case []any:
		for _, item := range v {
			visitMetadataNumbers(item, visit)
		}
	case map[string]any:
		for _, item := range v {
			visitMetadataNumbers(item, visit)
		}
	}
}

func preserveBrowserNumbers(stored, incoming any) (any, error) {
	switch old := stored.(type) {
	case json.Number:
		wire := browserNumber(old)
		if incoming == nil && wire == "" {
			return old, nil
		}
		if n, ok := incoming.(json.Number); ok && n != old && n == wire {
			return old, nil
		}
	case map[string]any:
		if patch, ok := incoming.(map[string]any); ok {
			for key, value := range patch {
				restored, err := preserveBrowserNumbers(old[key], value)
				if err != nil {
					return nil, err
				}
				patch[key] = restored
			}
			// Keep a redacted object's hidden fields on an unchanged echo.
			if reflect.DeepEqual(patch, redactAggregateSensitiveValue(old)) {
				return old, nil
			}
		}
	case []any:
		if values, ok := incoming.([]any); ok {
			return restoreMetadataArray(old, values)
		}
	}
	return incoming, nil
}

func restoreMetadataArray(old, values []any) ([]any, error) {
	// Compare before recursively restoring values in-place. A matching rounded
	// token may have moved to another element during an insert or reorder.
	rounded := make(map[json.Number]bool)
	visitMetadataNumbers(old, func(n json.Number) {
		if n == "" {
			return
		}
		if wire := browserNumber(n); wire != n {
			// A difference in decimal formatting alone does not lose precision.
			exact, ok := new(big.Rat).SetString(n.String())
			parsed, valid := new(big.Rat).SetString(wire.String())
			if !ok || !valid || exact.Cmp(parsed) != 0 {
				rounded[wire] = true
			}
		}
	})
	containsEcho := func(value any) bool {
		found := false
		if len(rounded) > 0 {
			visitMetadataNumbers(value, func(n json.Number) {
				if rounded[n] {
					found = true
				}
			})
		}
		return found
	}
	ambiguous := func() error {
		return services.NewSubjectWorkflowError("INTAKE_METADATA_ARRAY_AMBIGUOUS", "数组顺序或内容已变化，无法核对原数值精度，已停止保存")
	}
	for i, value := range values {
		echo := containsEcho(value)
		// A rounded survivor in a shortened array, or a rounded token in the
		// new tail, could belong to a different occurrence of the old value.
		if echo && (len(values) < len(old) || i >= len(old)) {
			return nil, ambiguous()
		}
		if i >= len(old) {
			continue
		}
		v, err := preserveBrowserNumbers(old[i], value)
		if err != nil {
			return nil, err
		}
		if echo && !reflect.DeepEqual(v, old[i]) {
			return nil, ambiguous()
		}
		values[i] = v
	}
	return values, nil
}

const (
	lawyerFactsVersionKey          = "lawyer_facts_version"
	npDescriptionSourceRevisionKey = "np_description_source_revision"
)
