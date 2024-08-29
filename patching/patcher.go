// patching provides functionality to patch resources
package patching

import (
	"iter"
	"maps"
	"reflect"
	"slices"
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
		return nil, errors.Newf("old must be of kind struct, found kind %s", kind.String())
	}

	newMap := patchSet.data
	oldMap := map[string]any{}
	for _, field := range reflect.VisibleFields(oldType) {
		oldMap[field.Name] = oldValue.FieldByName(field.Name).Interface()
	}

	diff := map[string]DiffElem{}
	for _, key := range unionKeys(maps.Keys(oldMap), maps.Keys(newMap)) {
		oldV, foundInOld := oldMap[key]
		newV, foundInNew := newMap[key]

		if !foundInOld || !foundInNew {
			diff[key] = DiffElem{
				Old: oldV,
				New: newV,
			}

			continue
		}

		if o, n := reflect.TypeOf(oldV), reflect.TypeOf(newV); !o.Comparable() || !n.Comparable() {
			return nil, errors.Newf("attempted to diff incomparable types, old: %s, new: %s", o.Name(), n.Name())
		}

		if oldV != newV {
			diff[key] = DiffElem{
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

func unionKeys(seqs ...iter.Seq[string]) []string {
	union := []string{}

	for _, seq := range seqs {
		union = slices.AppendSeq(union, seq)
	}

	slices.Sort(union)
	union = slices.Compact(union)

	return union
}

type DiffElem struct {
	Old any
	New any
}
