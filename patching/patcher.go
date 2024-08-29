// patching provides functionality to patch resources
package patching

import (
	"encoding"
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/go-playground/errors/v5"
)

type Patcher struct {
	tagName string

	mu    sync.RWMutex
	cache map[reflect.Type]map[string]string
}

func NewSpannerPatcher() *Patcher {
	return &Patcher{
		cache:   make(map[reflect.Type]map[string]string),
		tagName: "spanner",
	}
}

func (tm *Patcher) get(t reflect.Type) (map[string]string, error) {
	tm.mu.RLock()
	if tagMap, ok := tm.cache[t]; ok {
		defer tm.mu.RUnlock()

		return tagMap, nil
	}
	tm.mu.RUnlock()

	tm.mu.Lock()
	defer tm.mu.Unlock()

	if tagMap, ok := tm.cache[t]; ok {
		return tagMap, nil
	}

	tm.cache[t] = structTags(t, tm.tagName)

	return tm.cache[t], nil
}

// Resolve returns a map with the keys set to the database struct tags found on databaseType, and the values set to the values in patchSet.
func (tm *Patcher) Resolve(patchSet *PatchSet, databaseType any) (map[string]any, error) {
	t := reflect.TypeOf(databaseType)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return nil, errors.Newf("expected struct, got %s", t.Kind())
	}

	fieldTagMapping, err := tm.get(t)
	if err != nil {
		return nil, err
	}

	newMap := make(map[string]any, len(patchSet.data))
	for structField, value := range patchSet.data {
		tag, ok := fieldTagMapping[structField]
		if !ok {
			return nil, errors.Newf("field %s not found in struct", structField)
		}
		newMap[tag] = value
	}

	return newMap, nil
}

func (tm *Patcher) Diff(old any, patchSet *PatchSet) (map[string]DiffElem, error) {
	oldValue := reflect.ValueOf(old)
	oldType := reflect.TypeOf(old)

	if oldType.Kind() == reflect.Pointer {
		oldType.Elem()
	}

	if kind := oldType.Kind(); kind != reflect.Struct {
		return nil, errors.Newf("Patcher.Diff(): old must be of kind struct, found kind %s", kind.String())
	}

	newMap := patchSet.data

	oldMap := map[string]any{}
	for _, field := range reflect.VisibleFields(oldType) {
		oldMap[field.Name] = oldValue.FieldByName(field.Name).Interface()
	}

	diff := map[string]DiffElem{}
	for field, newV := range newMap {
		oldV, foundInOld := oldMap[field]
		if !foundInOld {
			return nil, errors.Newf("Patcher.Diff(): field %s in patchSet does not exist in old", field)
		}

		switch ot := oldV.(type) {
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64, string, bool:
			switch nt := newV.(type) {
			case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64, string, bool:
				if ot != nt {
					return nil, errors.Newf("Patcher.Diff(): attempted to diff incomparable types, old: %T, new: %T", ot, nt)
				}
			default:
				return nil, errors.Newf("Patcher.Diff(): attempted to diff incomparable types, old: %T, new: %T", ot, nt)
			}

			if oldV != newV {
				diff[field] = DiffElem{
					Old: oldV,
					New: newV,
				}
			}
		case *int, *int8, *int16, *int32, *int64, *uint, *uint8, *uint16, *uint32, *uint64, *float32, *float64, *string, *bool:
			derefOldV, ot := derefPrimitive(oldV)
			switch nt := newV.(type) {
			case *int, *int8, *int16, *int32, *int64, *uint, *uint8, *uint16, *uint32, *uint64, *float32, *float64, *string, *bool:
				derefNewV, nt := derefPrimitive(newV)
				if ot != nt {
					return nil, errors.Newf("Patcher.Diff(): attempted to diff incomparable types, old: %T, new: %T", ot, nt)
				}
				if derefOldV != derefNewV {
					diff[field] = DiffElem{
						Old: oldV,
						New: newV,
					}
				}
			default:
				return nil, errors.Newf("Patcher.Diff(): attempted to diff incomparable types, old: %T, new: %T", ot, nt)
			}
		case []int, []int8, []int16, []int32, []int64, []uint, []uint8, []uint16, []uint32, []uint64, []float32, []float64, []string, []bool:
			return nil, errors.Newf("Patcher.Diff(): Not implemented for types old: %T, new: %T", oldV, newV)
		case *[]int, *[]int8, *[]int16, *[]int32, *[]int64, *[]uint, *[]uint8, *[]uint16, *[]uint32, *[]uint64, *[]float32, *[]float64, *[]string, *[]bool:
			return nil, errors.Newf("Patcher.Diff(): Not implemented for types old: %T, new: %T", oldV, newV)
		default:
			oldStringV, ok := marshalText(oldV)
			if ok {
				newStringV, ok := marshalText(newV)
				if ok {
					if oldStringV != newStringV {
						diff[field] = DiffElem{
							Old: oldV,
							New: newV,
						}
					}
				}
			}

			return nil, errors.Newf("Patcher.Diff(): Not implemented for types old: %T, new: %T", oldV, newV)
		}
	}

	return diff, nil
}

func structTags(t reflect.Type, key string) map[string]string {
	tagMap := make(map[string]string)
	for i := range t.NumField() {
		field := t.Field(i)
		tag := field.Tag.Get(key)

		list := strings.Split(tag, ",")
		if len(list) == 0 || list[0] == "" || list[0] == "-" {
			continue
		}

		tagMap[field.Name] = list[0]
	}

	return tagMap
}

func derefPrimitive(v any) (derefv, vType any) {
	switch t := v.(type) {
	case *int:
		return any((*t)), t
	case *int8:
		return any((*t)), t
	case *int16:
		return any((*t)), t
	case *int32:
		return any((*t)), t
	case *int64:
		return any((*t)), t
	case *uint:
		return any((*t)), t
	case *uint8:
		return any((*t)), t
	case *uint16:
		return any((*t)), t
	case *uint32:
		return any((*t)), t
	case *uint64:
		return any((*t)), t
	case *float32:
		return any((*t)), t
	case *float64:
		return any((*t)), t
	case *string:
		return any((*t)), t
	case *bool:
		return any((*t)), t
	}

	panic(errors.Newf("deref(): unsupported type %T", v))
}

func marshalText(v any) (val string, ok bool) {
	t := reflect.TypeOf(v)
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	interfaceType := reflect.TypeOf((*encoding.TextMarshaler)(nil)).Elem()
	if reflect.PointerTo(t).Implements(interfaceType) {
		enc, ok := v.(encoding.TextMarshaler)
		if !ok {
			panic(fmt.Sprintf("type assertion failed for %T", v))
		}
		text, err := enc.MarshalText()
		if err != nil {
			panic(errors.Wrap(err, "MarshalText()"))
		}

		return string(text), true
	}

	return "", false
}

type DiffElem struct {
	Old any
	New any
}
