// package patchset provides types to store json patch set mapping to struct fields.
package patchset

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

func (p *PatchSet) Set(field string, value any) {
	p.data[field] = value
}

func (p *PatchSet) StructFields() []string {
	return slices.Collect(maps.Keys(p.data))
}

func (p *PatchSet) Len() int {
	return len(p.data)
}

func (p *PatchSet) Data() map[string]any {
	return p.data
}
