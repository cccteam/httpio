// package resourceset is a set of resources that provides a way to map permissions to fields in a struct.
package resourceset

import (
	"fmt"
	reflect "reflect"

	"github.com/cccteam/ccc/accesstypes"
	"github.com/go-playground/errors/v5"
)

type ResourceSet struct {
	requiredPermission accesstypes.Permission
	requiredPermFields []string
	resource           accesstypes.Resource
}

func New(v any, resource accesstypes.Resource, requiredPermission accesstypes.Permission) (*ResourceSet, error) {
	requiredPermFields, err := permissionsFromTags(v)
	if err != nil {
		panic(err)
	}

	return &ResourceSet{
		requiredPermission: requiredPermission,
		requiredPermFields: requiredPermFields,
		resource:           resource,
	}, nil
}

func (r *ResourceSet) Fields() []string {
	return r.requiredPermFields
}

func (r *ResourceSet) RequiredPermission() accesstypes.Permission {
	return r.requiredPermission
}

func (r *ResourceSet) Contains(fieldName string) bool {
	for _, required := range r.requiredPermFields {
		if required == fieldName {
			return true
		}
	}

	return false
}

func (r *ResourceSet) Resource(fieldName string) accesstypes.Resource {
	return accesstypes.Resource(fmt.Sprintf("%s.%s", r.resource, fieldName))
}

func permissionsFromTags(v any) (fields []string, err error) {
	vType := reflect.TypeOf(v)
	if vType.Kind() == reflect.Ptr {
		vType = vType.Elem()
	}
	if vType.Kind() != reflect.Struct {
		return nil, errors.Newf("expected a struct, got %s", vType.Kind())
	}

	for i := range vType.NumField() {
		field := vType.Field(i)
		tagList := field.Tag.Get("perm") // `perm:"required"`
		if tagList == "required" {
			fields = append(fields, field.Name)
		}
	}

	return fields, nil
}
