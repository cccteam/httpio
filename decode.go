package httpio

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"reflect"
	"strings"

	"github.com/cccteam/access/accesstypes"
	"github.com/cccteam/access/resourceset"
	"github.com/cccteam/httpio/patching"
	"github.com/go-playground/errors/v5"
)

// ValidatorFunc is a function that validates s
// It returns an error if the validation fails
type ValidatorFunc func(s interface{}) error

type Enforcer interface {
	RequireResource(user accesstypes.User, domain accesstypes.Domain, perms accesstypes.Permission, resources ...resourceset.Name) error
}

// Decoder is a struct that can be used for decoding http requests and validating those requests
type Decoder[T any] struct {
	validate    ValidatorFunc
	fieldMapper *patching.FieldMapper
}

func NewDecoder[T any]() (*Decoder[T], error) {
	target := new(T)

	m, err := patching.NewFieldMapper(target)
	if err != nil {
		return nil, errors.Wrap(err, "NewFieldMapper()")
	}

	return &Decoder[T]{
		fieldMapper: m,
	}, nil
}

func (d *Decoder[T]) WithValidator(v ValidatorFunc) *Decoder[T] {
	decoder := *d
	decoder.validate = v

	return &decoder
}

func (d *Decoder[T]) WithPermissionChecker(userFromCtx func(context.Context) accesstypes.User, enforcer Enforcer, perms *resourceset.Set) *DecoderWithPermissionChecker[T] {
	return &DecoderWithPermissionChecker[T]{
		userFromCtx:       userFromCtx,
		permissionChecker: enforcer,
		permissions:       perms,
		fieldMapper:       d.fieldMapper,
	}
}

// Decode parses the http request body and validates it against the struct validation rules
func (d *Decoder[T]) Decode(request *http.Request) (*T, error) {
	target := new(T)
	if _, err := decodeToMap(d.fieldMapper, request, target, d.validate); err != nil {
		return nil, err
	}

	return target, nil
}

func (d *Decoder[T]) DecodeToPatchSet(request *http.Request) (*patching.PatchSet, error) {
	target := new(T)
	p, err := decodeToMap(d.fieldMapper, request, target, d.validate)
	if err != nil {
		return nil, err
	}

	return p, nil
}

type DecoderWithPermissionChecker[T any] struct {
	userFromCtx       func(context.Context) accesstypes.User
	validate          ValidatorFunc
	permissionChecker Enforcer
	permissions       *resourceset.Set
	fieldMapper       *patching.FieldMapper
}

func (d *DecoderWithPermissionChecker[T]) WithValidator(v ValidatorFunc) *DecoderWithPermissionChecker[T] {
	decoder := *d
	decoder.validate = v

	return &decoder
}

// Decode parses the http request body and validates it against the struct validation rules
func (d *DecoderWithPermissionChecker[T]) Decode(request *http.Request) (*T, error) {
	target := new(T)
	p, err := decodeToMap(d.fieldMapper, request, target, d.validate)
	if err != nil {
		return nil, err
	}

	if err := checkPermissions(p, d.permissionChecker, d.permissions, d.userFromCtx(request.Context())); err != nil {
		return nil, err
	}

	return target, nil
}

func (d *DecoderWithPermissionChecker[T]) DecodeToPatchSet(request *http.Request) (*patching.PatchSet, error) {
	target := new(T)
	p, err := decodeToMap(d.fieldMapper, request, target, d.validate)
	if err != nil {
		return nil, err
	}

	if err := checkPermissions(p, d.permissionChecker, d.permissions, d.userFromCtx(request.Context())); err != nil {
		return nil, err
	}

	return p, nil
}

func decodeToMap[T any](fieldMapper *patching.FieldMapper, request *http.Request, target *T, validate ValidatorFunc) (*patching.PatchSet, error) {
	// This can be optimized with a forkReader
	bodyBuf := &bytes.Buffer{}
	if _, err := io.Copy(bodyBuf, request.Body); err != nil {
		return nil, errors.Wrap(err, "io.Copy()")
	}

	bodyReader := bytes.NewReader(bodyBuf.Bytes())
	if err := json.NewDecoder(bodyReader).Decode(target); err != nil {
		return nil, errors.Wrap(err, "Decoder.Decode()")
	}

	jsonData := make(map[string]any)
	if err := json.Unmarshal(bodyBuf.Bytes(), &jsonData); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal request body")
	}

	if validate != nil {
		if err := validate(target); err != nil {
			return nil, NewBadRequestMessageWithErrorf(err, "failed validating the request")
		}
	}

	vValue := reflect.ValueOf(target)
	if vValue.Kind() == reflect.Ptr {
		vValue = vValue.Elem()
	}
	changes := make(map[string]any)
	for jsonField := range jsonData {
		fieldName, ok := fieldMapper.StructFieldName(jsonField)
		if !ok {
			fieldName, ok = fieldMapper.StructFieldName(strings.ToLower(jsonField))
			if !ok {
				return nil, NewBadRequestMessagef("invalid field in json - %s", jsonField)
			}
		}

		value := vValue.FieldByName(fieldName).Interface()
		if value == nil {
			return nil, NewBadRequestMessagef("invalid field in json - %s", jsonField)
		}

		if _, ok := changes[fieldName]; ok {
			return nil, NewBadRequestMessagef("json field name %s collides with another field name of different case", fieldName)
		}
		changes[fieldName] = value
	}

	patchSet := patching.NewPatchSet(changes)

	return patchSet, nil
}

func checkPermissions(patchSet *patching.PatchSet, permissionChecker Enforcer, resourceSet *resourceset.Set, user accesstypes.User) error {
	for _, fieldName := range patchSet.Fields() {
		if resourceSet.Contains(fieldName) {
			// TODO: Domain must be passed in some how
			if err := permissionChecker.RequireResource(user, accesstypes.GlobalDomain, resourceSet.RequiredPermission(), resourceSet.ResourceName(fieldName)); err != nil {
				return errors.Wrap(err, "permissionChecker.RequireResource()")
			}
		}
	}

	return nil
}
