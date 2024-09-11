package patching

import (
	"maps"
	"slices"
)

type PatchSet struct {
	data map[string]any
}

func NewPatchSet(data map[string]any) *PatchSet {
	return &PatchSet{
		data: data,
	}
}

func (s *PatchSet) Set(field string, value any) {
	s.data[field] = value
}

func (s *PatchSet) Fields() []string {
	return slices.Collect(maps.Keys(s.data))
}

func (s *PatchSet) Len() int {
	return len(s.data)
}
