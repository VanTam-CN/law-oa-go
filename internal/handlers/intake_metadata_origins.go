package handlers

import (
	"encoding/json"
	"fmt"
)

// A browser projection can stand for several different exact stored values.
type metadataOrigins map[string][]any

func (origins metadataOrigins) add(original any) error {
	key, err := json.Marshal(metadataArrayProjection(original))
	if err != nil {
		return fmt.Errorf("序列化 JSON: %w", err)
	}
	group := origins[string(key)]
	if len(group) == 0 || !equalMetadataValue(group[0], original) {
		origins[string(key)] = append(group, original)
	}
	return nil
}

func (origins metadataOrigins) check(wire, restored any) error {
	projection := metadataArrayProjection(wire)
	key, err := json.Marshal(projection)
	if err != nil {
		return fmt.Errorf("序列化 JSON: %w", err)
	}
	group := origins[string(key)]
	if len(group) > 1 && equalMetadataValue(wire, projection) {
		return ambiguousMetadataArray()
	}
	// A lossless client can discriminate numbers sharing a browser key.
	for _, original := range group {
		if matchesMetadataEcho(original, wire) && !equalMetadataValue(original, restored) {
			return ambiguousMetadataArray()
		}
	}
	return nil
}

// A cut and paste between two nested arrays has no common local source slot.
// Index displaced protected objects across the field, including removed slots.
func checkMovedHiddenObjects(stored, wire, restored any) error {
	moved := make(metadataOrigins)
	err := visitMetadataObjects(stored, wire, nil, func(object map[string]any, atPath, _ any) error {
		if matchesMetadataEcho(object, atPath) || equalMetadataValue(object, redactAggregateSensitiveValue(object)) {
			return nil
		}
		return moved.add(object)
	})
	if err != nil || len(moved) == 0 {
		return err
	}
	return visitMetadataObjects(wire, stored, restored, func(object map[string]any, prior, saved any) error {
		// An unchanged peer at its original path is not a paste destination.
		if matchesMetadataEcho(prior, object) && equalMetadataValue(prior, saved) {
			return nil
		}
		return moved.check(object, saved)
	})
}

// Walk objects with the values at the same path in two other JSON trees.
func visitMetadataObjects(value, peer, saved any, visit func(map[string]any, any, any) error) error {
	switch v := value.(type) {
	case map[string]any:
		if err := visit(v, peer, saved); err != nil {
			return err
		}
		other, _ := peer.(map[string]any)
		result, _ := saved.(map[string]any)
		for k, item := range v {
			if err := visitMetadataObjects(item, other[k], result[k], visit); err != nil {
				return err
			}
		}
	case []any:
		other, _ := peer.([]any)
		result, _ := saved.([]any)
		for i, item := range v {
			var previous, current any
			if i < len(other) {
				previous = other[i]
			}
			if i < len(result) {
				current = result[i]
			}
			if err := visitMetadataObjects(item, previous, current, visit); err != nil {
				return err
			}
		}
	}
	return nil
}
