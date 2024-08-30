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

	if oldValue.Kind() == reflect.Pointer {
		oldValue = oldValue.Elem()
	}

	if oldType.Kind() == reflect.Pointer {
		oldType = oldType.Elem()
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

		if match, err := match(oldV, newV); err != nil {
			return nil, err
		} else if !match {
			diff[field] = DiffElem{
				Old: oldV,
				New: newV,
			}
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

func match(v, v2 any) (matched bool, err error) {
	switch t := v.(type) {
	case int:
		return matchPrimitive(t, v2)
	case int8:
		return matchPrimitive(t, v2)
	case int16:
		return matchPrimitive(t, v2)
	case int32:
		return matchPrimitive(t, v2)
	case int64:
		return matchPrimitive(t, v2)
	case uint:
		return matchPrimitive(t, v2)
	case uint8:
		return matchPrimitive(t, v2)
	case uint16:
		return matchPrimitive(t, v2)
	case uint32:
		return matchPrimitive(t, v2)
	case uint64:
		return matchPrimitive(t, v2)
	case float32:
		return matchPrimitive(t, v2)
	case float64:
		return matchPrimitive(t, v2)
	case string:
		return matchPrimitive(t, v2)
	case bool:
		return matchPrimitive(t, v2)
	case *int:
		return matchPrimitivePtr(t, v2)
	case *int8:
		return matchPrimitivePtr(t, v2)
	case *int16:
		return matchPrimitivePtr(t, v2)
	case *int32:
		return matchPrimitivePtr(t, v2)
	case *int64:
		return matchPrimitivePtr(t, v2)
	case *uint:
		return matchPrimitivePtr(t, v2)
	case *uint8:
		return matchPrimitivePtr(t, v2)
	case *uint16:
		return matchPrimitivePtr(t, v2)
	case *uint32:
		return matchPrimitivePtr(t, v2)
	case *uint64:
		return matchPrimitivePtr(t, v2)
	case *float32:
		return matchPrimitivePtr(t, v2)
	case *float64:
		return matchPrimitivePtr(t, v2)
	case *string:
		return matchPrimitivePtr(t, v2)
	case *bool:
		return matchPrimitivePtr(t, v2)
	default:
		strV, ok := marshalText(v)
		if ok {
			strV2, ok := marshalText(v2)
			if ok {
				return strV == strV2, nil
			}
		}

		// FIXME: Add support for slices and Named Types based on primitive types

		panic(errors.Newf("deref(): unsupported type %T", v))
	}
}

func matchPrimitive[T comparable](v T, v2 any) (bool, error) {
	t2, ok := v2.(T)
	if !ok {
		return false, errors.Newf("deref(): attempted to diff incomparable types, old: %T, new: %T", v, v2)
	}
	if v == t2 {
		return true, nil
	}

	return false, nil
}

func matchPrimitivePtr[T comparable](v *T, v2 any) (bool, error) {
	t2, ok := v2.(*T)
	if !ok {
		return false, errors.Newf("deref(): attempted to diff incomparable types, old: %T, new: %T", v, v2)
	}
	if v == nil || t2 == nil {
		if v == nil && t2 == nil {
			return true, nil
		}

		return false, nil
	}
	if *v == *t2 {
		return true, nil
	}

	return false, nil
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
