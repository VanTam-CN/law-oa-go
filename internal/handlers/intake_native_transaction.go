package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"math/big"
	"reflect"
	"slices"

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

func metadataNumbersEqual(a, b json.Number) bool {
	if a == b {
		return true
	}
	left, ok := new(big.Rat).SetString(a.String())
	if !ok {
		return false
	}
	right, valid := new(big.Rat).SetString(b.String())
	return valid && left.Cmp(right) == 0
}

// Compare exact restore candidates with JSON numeric semantics. Browser
// rounding is only for recognizing echoes; it cannot establish equality.
func equalMetadataValue(left, right any) bool {
	switch a := left.(type) {
	case json.Number:
		b, ok := right.(json.Number)
		return ok && metadataNumbersEqual(a, b)
	case []any:
		b, ok := right.([]any)
		return ok && slices.EqualFunc(a, b, equalMetadataValue)
	case map[string]any:
		b, ok := right.(map[string]any)
		return ok && maps.EqualFunc(a, b, equalMetadataValue)
	default:
		return reflect.DeepEqual(left, right)
	}
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
			result := maps.Clone(patch)
			for key, value := range patch {
				restored, err := preserveBrowserNumbers(old[key], value)
				if err != nil {
					return nil, err
				}
				result[key] = restored
			}
			// Keep a redacted object's hidden fields on an unchanged echo.
			if reflect.DeepEqual(result, redactAggregateSensitiveValue(old)) {
				return old, nil
			}
			return result, nil
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
		if wire := browserNumber(n); !metadataNumbersEqual(n, wire) {
			rounded[wire] = true
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
	// Raw null represents both a literal null and a non-finite browser number.
	echoKey := func(n json.Number, stored bool) json.Number {
		if stored && n != "" {
			return browserNumber(n)
		}
		return n
	}
	containsEcho := func(value any, stored bool) bool {
		found := false
		if len(echoes) != 0 {
			visitMetadataNumbers(value, func(n json.Number) {
				if echoes[echoKey(n, stored)] {
					found = true
				}
			})
		}
		return found
	}
	ambiguous := func() error {
		return services.NewSubjectWorkflowError("INTAKE_METADATA_ARRAY_AMBIGUOUS", "数组顺序或内容已变化，无法核对原数值精度，已停止保存")
	}
	// Compare both interpretations against the original wire document; child
	// restorations return fresh containers and never change that document.
	drop := -1
	if len(echoes) != 0 {
		drop = possibleMetadataDrop(old, values)
	}
	result := slices.Clone(values)
	// If an old slot changes shape, the wire might also be a reorder.
	// Only relax its prefix check for echoes with one exact interpretation.
	var needsConsensus map[json.Number]bool
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
		if drop >= 0 && i >= drop {
			// A shifted survivor retains its exact PG value. An appended item
			// is new, so it retains the value actually supplied on the wire.
			alternative := value
			if i+1 < len(old) {
				alternative = old[i+1]
			}
			if !equalMetadataValue(v, alternative) {
				return nil, ambiguous()
			}
		}
		if (echo || containsEcho(old[i], true)) && !retainsMetadataPrefix(old[i], v) {
			if drop < 0 || i < drop {
				return nil, ambiguous()
			}
			if needsConsensus == nil {
				needsConsensus = make(map[json.Number]bool)
			}
			visitMetadataNumbers(old[i], func(n json.Number) {
				if key := echoKey(n, true); echoes[key] {
					needsConsensus[key] = true
				}
			})
			visitMetadataNumbers(value, func(n json.Number) {
				if echoes[n] {
					needsConsensus[n] = true
				}
			})
		}
		result[i] = v
	}
	if len(needsConsensus) != 0 {
		exact := make(map[json.Number]json.Number)
		agrees := true
		check := func(n json.Number) {
			key := echoKey(n, true)
			if !needsConsensus[key] {
				return
			}
			if prior, found := exact[key]; found && !metadataNumbersEqual(prior, n) {
				agrees = false
			} else {
				exact[key] = n
			}
		}
		visitMetadataNumbers(old, check)
		visitMetadataNumbers(result, check)
		if !agrees {
			return nil, ambiguous()
		}
	}
	return result, nil
}

// Find the earliest deletion compatible with an unchanged prefix, surviving
// neighbours shifted left by one, and a new tail. Every later compatible
// deletion shares a suffix of this shift, so one comparison covers them all.
func possibleMetadataDrop(old, wire []any) int {
	if len(old) < 2 || len(wire) < len(old) {
		return -1
	}
	aligned := 0
	for aligned < len(old) && matchesMetadataEcho(old[aligned], wire[aligned]) {
		aligned++
	}
	if aligned == len(old) && len(old) == len(wire) {
		return -1 // A complete unchanged echo preserves its original order.
	}
	drop := len(old) - 1
	for drop > 0 && matchesMetadataEcho(old[drop], wire[drop-1]) {
		drop--
	}
	if drop == len(old)-1 || drop > aligned {
		return -1
	}
	return drop
}

// A surviving item may return losslessly or with browser-rounded numbers
// and the same field redactions as the legacy GET.
func matchesMetadataEcho(old, wire any) bool {
	switch original := old.(type) {
	case json.Number:
		if wire == original {
			return true
		}
		n := browserNumber(original)
		if n == "" {
			return wire == nil
		}
		return wire == n
	case []any:
		incoming, ok := wire.([]any)
		return ok && slices.EqualFunc(original, incoming, matchesMetadataEcho)
	case map[string]any:
		incoming, ok := wire.(map[string]any)
		if !ok {
			return false
		}
		matched := 0
		for key, value := range original {
			next, present := incoming[key]
			if !present && isAggregateSecretKey(key) {
				continue
			}
			if !present || !matchesMetadataEcho(value, next) {
				return false
			}
			matched++
		}
		return matched == len(incoming)
	default:
		return reflect.DeepEqual(original, wire)
	}
}

const (
	lawyerFactsVersionKey          = "lawyer_facts_version"
	npDescriptionSourceRevisionKey = "np_description_source_revision"
)
