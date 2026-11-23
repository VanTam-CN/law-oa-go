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

// An echo-bearing slot must retain its previous contents. Recursion permits
// an append inside a child array without mistaking it for a moved number.
func retainsMetadataPrefix(old, restored any) bool {
	switch before := old.(type) {
	case []any:
		after, ok := restored.([]any)
		if !ok || len(after) < len(before) {
			return false
		}
		for i := range before {
			if !retainsMetadataPrefix(before[i], after[i]) {
				return false
			}
		}
		return true
	case map[string]any:
		after, ok := restored.(map[string]any)
		if !ok || len(after) != len(before) {
			return false
		}
		for k, v := range before {
			next, exists := after[k]
			if !exists || !retainsMetadataPrefix(v, next) {
				return false
			}
		}
		return true
	default:
		return reflect.DeepEqual(old, restored)
	}
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
	// A raw wire number can denote either a rounded old value or its exact
	// neighbour. Guard every possible source slot as well as its destination.
	echoes := make(map[json.Number]bool)
	if len(rounded) != 0 {
		visitMetadataNumbers(values, func(n json.Number) {
			if rounded[n] {
				echoes[n] = true
			}
		})
	}
	containsEcho := func(value any, stored bool) bool {
		found := false
		if len(echoes) != 0 {
			visitMetadataNumbers(value, func(n json.Number) {
				if stored && n != "" {
					n = browserNumber(n)
				}
				if echoes[n] {
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
		echo := containsEcho(value, false)
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
		if (echo || containsEcho(old[i], true)) && !retainsMetadataPrefix(old[i], v) {
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
