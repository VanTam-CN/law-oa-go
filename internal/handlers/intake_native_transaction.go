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
	// 超长数字串直接拒绝：防 big.Rat 解析指数放大内存（gosec G113 加固）
	if len(a.String()) > 64 || len(b.String()) > 64 {
		return false
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
	restored, err := restoreMetadataValue(stored, incoming)
	if err != nil {
		return nil, err
	}
	if err := checkMovedHiddenObjects(stored, incoming, restored); err != nil {
		return nil, err
	}
	return restored, nil
}

func restoreMetadataValue(stored, incoming any) (any, error) {
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
			// Test the untouched GET projection before restoring any nested keys.
			if matchesMetadataEcho(old, patch) {
				return old, nil
			}
			for key := range old {
				if _, sent := patch[key]; !sent && isAggregateSecretKey(key) {
					// The full-save protocol cannot identify a replacement object.
					return nil, ambiguousMetadataArray()
				}
			}
			result := maps.Clone(patch)
			for key, value := range patch {
				restored, err := restoreMetadataValue(old[key], value)
				if err != nil {
					return nil, err
				}
				result[key] = restored
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
		return equalMetadataValue(old, restored)
	}
}

// Use the legacy GET representation to locate whole array elements that could
// have moved. Only complete stored elements carry hidden fields and exact numbers.
func metadataArrayProjection(value any) any {
	switch v := value.(type) {
	case json.Number:
		if wire := browserNumber(v); wire != "" {
			return wire
		}
		return nil
	case map[string]any:
		out := make(map[string]any, len(v))
		for k, x := range v {
			if !isAggregateSecretKey(k) {
				out[k] = metadataArrayProjection(x)
			}
		}
		return out
	case []any:
		out := make([]any, len(v))
		for i, x := range v {
			out[i] = metadataArrayProjection(x)
		}
		return out
	default:
		return v
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
	// neighbor. Guard every possible source slot as well as its destination.
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

	// Compare both interpretations against the original wire document; child
	// restorations return fresh containers and never change that document.
	drop := -1
	if len(echoes) != 0 {
		drop = possibleMetadataDrop(old, values)
	}
	result := slices.Clone(values)

	for i, value := range values {
		echo := containsEcho(value, false)
		// A rounded survivor in a shortened array, or a rounded token in the
		// new tail, could belong to a different occurrence of the old value.
		if echo && (len(values) < len(old) || i >= len(old)) {
			return nil, ambiguousMetadataArray()
		}
		if i >= len(old) {
			continue
		}
		v, err := restoreMetadataValue(old[i], value)
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
				return nil, ambiguousMetadataArray()
			}
		}
		if (drop < 0 || i < drop) && (echo || containsEcho(old[i], true)) && !retainsMetadataPrefix(old[i], v) {
			return nil, ambiguousMetadataArray()
		}
		result[i] = v
	}
	if drop >= 0 {
		// The deleted slot could instead have moved to the tail and then grown.
		for i := len(old) - 1; i < len(values); i++ {
			if matchesMetadataProjection(old[drop], values[i], true) && !retainsMetadataPrefix(old[drop], result[i]) {
				return nil, ambiguousMetadataArray()
			}
		}
	}
	if err := checkDisplacedMetadataSlots(old, values, result); err != nil {
		return nil, err
	}
	return result, nil
}

func ambiguousMetadataArray() error {
	return services.NewSubjectWorkflowError("INTAKE_METADATA_ARRAY_AMBIGUOUS", "数组顺序或内容已变化，无法确定原数据的对应关系，已停止保存")
}

// Compare only whole elements whose original position no longer matches the GET
// projection. Their exact values include keys that the client could not see.
func checkDisplacedMetadataSlots(old, values, result []any) error {
	displaced := make(metadataOrigins)
	for i, original := range old {
		if i < len(values) && matchesMetadataEcho(original, values[i]) {
			continue
		}
		if err := displaced.add(original); err != nil {
			return err
		}
	}
	if len(displaced) == 0 {
		return nil
	}
	for i, value := range values {
		if err := displaced.check(value, result[i]); err != nil {
			return err
		}
	}
	return nil
}

// Find the earliest deletion compatible with an unchanged prefix, surviving
// neighbors shifted left by one, and a new tail. Every later compatible
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
	return matchesMetadataProjection(old, wire, false)
}

func matchesMetadataProjection(old, wire any, allowAppend bool) bool {
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
		if !ok || len(incoming) < len(original) || (!allowAppend && len(incoming) != len(original)) {
			return false
		}
		for i, item := range original {
			if !matchesMetadataProjection(item, incoming[i], allowAppend) {
				return false
			}
		}
		return true
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
			if !present || !matchesMetadataProjection(value, next, allowAppend) {
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
