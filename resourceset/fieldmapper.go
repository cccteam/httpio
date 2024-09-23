package resourceset

import (
	"maps"
	"reflect"
	"slices"
	"strings"

	"github.com/go-playground/errors/v5"
)

type FieldMapper struct {
	jsonTagToFields map[string]string
}

func NewFieldMapper(v any) (*FieldMapper, error) {
	jsonTagToFields, err := tagToFieldMap(v)
	if err != nil {
		return nil, err
	}

	return &FieldMapper{
		jsonTagToFields: jsonTagToFields,
	}, nil
}

func (f *FieldMapper) StructFieldName(tag string) (string, bool) {
	fieldName, ok := f.jsonTagToFields[tag]

	return fieldName, ok
}

func (f *FieldMapper) Len() int {
	return len(f.jsonTagToFields)
}

func (f *FieldMapper) Fields() []string {
	return slices.Collect(maps.Values(f.jsonTagToFields))
}

func tagToFieldMap(v any) (map[string]string, error) {
	vType := reflect.TypeOf(v)

	if vType.Kind() == reflect.Ptr {
		vType = vType.Elem()
	}
	if vType.Kind() != reflect.Struct {
		return nil, errors.Newf("argument v must be a struct, received %v", vType.Kind())
	}

	tfMap := make(map[string]string)
	for _, field := range reflect.VisibleFields(vType) {
		tag := field.Tag.Get("json")
		if tag == "" {
			if _, ok := tfMap[field.Name]; ok {
				return nil, errors.Newf("field name %s collides with another field tag", field.Name)
			}
			tfMap[field.Name] = field.Name
			if lowerFieldName := strings.ToLower(field.Name); lowerFieldName != field.Name {
				if _, ok := tfMap[lowerFieldName]; ok {
					return nil, errors.Newf("field name %s has multiple matches", field.Name)
				}
				tfMap[lowerFieldName] = field.Name
			}

			continue
		}

		if before, _, found := strings.Cut(tag, ","); found {
			tag = before
		}

		if tag == "-" {
			continue
		}

		if _, ok := tfMap[tag]; ok {
			return nil, errors.Newf("tag %s has multiple matches", tag)
		}
		tfMap[tag] = field.Name
	}

	return tfMap, nil
}
