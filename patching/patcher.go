// patching provides functionality to patch resources
package patching

import (
	"bytes"
	"encoding"
	"fmt"
	"iter"
	"reflect"
	"slices"
	"strings"
	"sync"

	"github.com/go-playground/errors/v5"
)

type dbType string

const (
	spanner  dbType = "spanner"
	postgres dbType = "postgres"
)

type Patcher struct {
	tagName string
	dbType  dbType

	mu    sync.RWMutex
	cache map[reflect.Type]map[string]string
}

func NewSpannerPatcher() *Patcher {
	return &Patcher{
		cache:   make(map[reflect.Type]map[string]string),
		tagName: "spanner",
		dbType:  spanner,
	}
}

func NewPostgresPatcher() *Patcher {
	return &Patcher{
		cache:   make(map[reflect.Type]map[string]string),
		tagName: "db",
		dbType:  postgres,
	}
}

func (tm *Patcher) Columns(patchSet *PatchSet, databaseType any) string {
	fieldTagMapping, err := tm.get(databaseType)
	if err != nil {
		panic(err)
	}

	columns := []string{}
	for _, field := range patchSet.Fields() {
		tag, ok := fieldTagMapping[field]
		if !ok {
			panic(errors.Newf("field %s not found in struct", field))
		}

		columns = append(columns, tag)
	}
	slices.Sort(columns)

	switch tm.dbType {
	case spanner:
		return strings.Join(columns, ", ")
	case postgres:
		return fmt.Sprintf(`"%s"`, strings.Join(columns, `", "`))
	default:
		panic(errors.Newf("unsupported tag name: %s", tm.tagName))
	}
}

func (tm *Patcher) get(v any) (map[string]string, error) {
	tm.mu.RLock()

	t := reflect.TypeOf(v)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

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

	if t.Kind() != reflect.Struct {
		return nil, errors.Newf("expected struct, got %s", t.Kind())
	}

	tm.cache[t] = structTags(t, tm.tagName)

	return tm.cache[t], nil
}

// Resolve returns a map with the keys set to the database struct tags found on databaseType, and the values set to the values in patchSet.
func (tm *Patcher) Resolve(pkeys PrimaryKeys, patchSet *PatchSet, databaseType any) (map[string]any, error) {
	if len(pkeys) == 0 {
		return nil, errors.New("must include at least one primary key in call to Resolve")
	}

	fieldTagMapping, err := tm.get(databaseType)
	if err != nil {
		return nil, err
	}

	newMap := make(map[string]any, len(pkeys)+len(patchSet.data))
	for structField, value := range all(patchSet.data, pkeys) {
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

// all returns an iterator over key-value pairs from m.
//   - all is a similar to maps.All but it takes a variadic
//   - duplicate keys will not be deduped and will be yielded once for each duplication
func all[Map ~map[K]V, K comparable, V any](maps ...Map) iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		for _, m := range maps {
			for k, v := range m {
				if !yield(k, v) {
					return
				}
			}
		}
	}
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
	case *int:
		return matchPrimitivePtr(t, v2)
	case []int:
		return matchSlice(t, v2)
	case []*int:
		return matchSlice(t, v2)
	case int8:
		return matchPrimitive(t, v2)
	case *int8:
		return matchPrimitivePtr(t, v2)
	case []int8:
		return matchSlice(t, v2)
	case []*int8:
		return matchSlice(t, v2)
	case int16:
		return matchPrimitive(t, v2)
	case *int16:
		return matchPrimitivePtr(t, v2)
	case []int16:
		return matchSlice(t, v2)
	case []*int16:
		return matchSlice(t, v2)
	case int32:
		return matchPrimitive(t, v2)
	case *int32:
		return matchPrimitivePtr(t, v2)
	case []int32:
		return matchSlice(t, v2)
	case []*int32:
		return matchSlice(t, v2)
	case int64:
		return matchPrimitive(t, v2)
	case *int64:
		return matchPrimitivePtr(t, v2)
	case []int64:
		return matchSlice(t, v2)
	case []*int64:
		return matchSlice(t, v2)
	case uint:
		return matchPrimitive(t, v2)
	case *uint:
		return matchPrimitivePtr(t, v2)
	case []uint:
		return matchSlice(t, v2)
	case []*uint:
		return matchSlice(t, v2)
	case uint8:
		return matchPrimitive(t, v2)
	case *uint8:
		return matchPrimitivePtr(t, v2)
	case []uint8:
		return matchSlice(t, v2)
	case []*uint8:
		return matchSlice(t, v2)
	case uint16:
		return matchPrimitive(t, v2)
	case *uint16:
		return matchPrimitivePtr(t, v2)
	case []uint16:
		return matchSlice(t, v2)
	case []*uint16:
		return matchSlice(t, v2)
	case uint32:
		return matchPrimitive(t, v2)
	case *uint32:
		return matchPrimitivePtr(t, v2)
	case []uint32:
		return matchSlice(t, v2)
	case []*uint32:
		return matchSlice(t, v2)
	case uint64:
		return matchPrimitive(t, v2)
	case *uint64:
		return matchPrimitivePtr(t, v2)
	case []uint64:
		return matchSlice(t, v2)
	case []*uint64:
		return matchSlice(t, v2)
	case float32:
		return matchPrimitive(t, v2)
	case *float32:
		return matchPrimitivePtr(t, v2)
	case []float32:
		return matchSlice(t, v2)
	case []*float32:
		return matchSlice(t, v2)
	case float64:
		return matchPrimitive(t, v2)
	case *float64:
		return matchPrimitivePtr(t, v2)
	case []float64:
		return matchSlice(t, v2)
	case []*float64:
		return matchSlice(t, v2)
	case string:
		return matchPrimitive(t, v2)
	case *string:
		return matchPrimitivePtr(t, v2)
	case []string:
		return matchSlice(t, v2)
	case []*string:
		return matchSlice(t, v2)
	case bool:
		return matchPrimitive(t, v2)
	case *bool:
		return matchPrimitivePtr(t, v2)
	case []bool:
		return matchSlice(t, v2)
	case []*bool:
		return matchSlice(t, v2)
	case encoding.TextMarshaler:
		return matchTextMarshaler(t, v2)
	case fmt.Stringer:
		return matchStringer(t, v2)
	}

	if reflect.TypeOf(v) != reflect.TypeOf(v2) {
		return false, errors.Newf("attempted to compare values having a different type, v.(type) = %T, v2.(type) = %T", v, v2)
	}

	return reflect.DeepEqual(v, v2), nil
}

func matchSlice[T comparable](v []T, v2 any) (bool, error) {
	t2, ok := v2.([]T)
	if !ok {
		return false, errors.Newf("matchSlice(): attempted to diff incomparable types, old: %T, new: %T", v, v2)
	}
	if len(v) != len(t2) {
		return false, nil
	}

	for i := range v {
		if match, err := match(v[i], t2[i]); err != nil {
			return false, err
		} else if !match {
			return false, nil
		}
	}

	return true, nil
}

func matchPrimitive[T comparable](v T, v2 any) (bool, error) {
	t2, ok := v2.(T)
	if !ok {
		return false, errors.Newf("matchPrimitive(): attempted to diff incomparable types, old: %T, new: %T", v, v2)
	}
	if v == t2 {
		return true, nil
	}

	return false, nil
}

func matchPrimitivePtr[T comparable](v *T, v2 any) (bool, error) {
	t2, ok := v2.(*T)
	if !ok {
		return false, errors.Newf("matchPrimitivePtr(): attempted to diff incomparable types, old: %T, new: %T", v, v2)
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

func matchTextMarshaler(v encoding.TextMarshaler, v2 any) (bool, error) {
	vText, err := v.MarshalText()
	if err != nil {
		return false, errors.Wrap(err, "encoding.TextMarshaler.MarshalText()")
	}

	t2, ok := v2.(encoding.TextMarshaler)
	if !ok {
		return false, errors.Newf("matchTextMarshaler(): v2 does not implement encoding.TextMarshaler: %T", v2)
	}

	v2Text, err := t2.MarshalText()
	if err != nil {
		return false, errors.Wrap(err, "encoding.TextMarshaler.MarshalText()")
	}

	if bytes.Equal(vText, v2Text) {
		return true, nil
	}

	return false, nil
}

func matchStringer(v fmt.Stringer, v2 any) (bool, error) {
	t2, ok := v2.(fmt.Stringer)
	if !ok {
		return false, errors.Newf("matchStringer(): v2 does not implement fmt.Stringer: %T", v2)
	}

	if v.String() == t2.String() {
		return true, nil
	}

	return false, nil
}

type PrimaryKeys map[string]any

func PKey(key string, value any) PrimaryKeys {
	return PrimaryKeys{key: value}
}

func (p PrimaryKeys) Add(key string, value any) PrimaryKeys {
	p[key] = value

	return p
}

type DiffElem struct {
	Old any
	New any
}
