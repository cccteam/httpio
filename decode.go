package httpio

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"reflect"
	"strings"
	"sync"

	"github.com/cccteam/ccc/accesstypes"
	"github.com/cccteam/httpio/patchset"
	"github.com/cccteam/httpio/resourceset"
	"github.com/go-playground/errors/v5"
)

// ValidatorFunc is a function that validates s
// It returns an error if the validation fails
type ValidatorFunc func(s interface{}) error

type Enforcer interface {
	RequireResources(ctx context.Context, user accesstypes.User, domain accesstypes.Domain, perms accesstypes.Permission, resources ...accesstypes.Resource) error
}

type (
	DomainFromReq func(*http.Request) accesstypes.Domain
	UserFromReq   func(*http.Request) accesstypes.User
)

// Decoder is a struct that can be used for decoding http requests and validating those requests
type Decoder[T any] struct {
	validate    ValidatorFunc
	fieldMapper *patchset.FieldMapper
}

func NewDecoder[T any]() (*Decoder[T], error) {
	target := new(T)

	m, err := patchset.NewFieldMapper(target)
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

func (d *Decoder[T]) WithPermissionChecker(domainFromReq DomainFromReq, userFromReq UserFromReq, permissionChecker Enforcer, perms *resourceset.ResourceSet) *DecoderWithPermissionChecker[T] {
	return &DecoderWithPermissionChecker[T]{
		userFromReq:       userFromReq,
		domainFromReq:     domainFromReq,
		permissionChecker: permissionChecker,
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

func (d *Decoder[T]) DecodeToPatchSet(request *http.Request) (*patchset.PatchSet, error) {
	target := new(T)
	p, err := decodeToMap(d.fieldMapper, request, target, d.validate)
	if err != nil {
		return nil, err
	}

	return p, nil
}

type DecoderWithPermissionChecker[T any] struct {
	userFromReq       UserFromReq
	domainFromReq     DomainFromReq
	validate          ValidatorFunc
	permissionChecker Enforcer
	permissions       *resourceset.ResourceSet
	fieldMapper       *patchset.FieldMapper
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

	if err := checkPermissions(request.Context(), p, d.permissionChecker, d.permissions, d.domainFromReq(request), d.userFromReq(request)); err != nil {
		return nil, err
	}

	return target, nil
}

func (d *DecoderWithPermissionChecker[T]) DecodeToPatchSet(request *http.Request) (*patchset.PatchSet, error) {
	target := new(T)
	p, err := decodeToMap(d.fieldMapper, request, target, d.validate)
	if err != nil {
		return nil, err
	}

	if err := checkPermissions(request.Context(), p, d.permissionChecker, d.permissions, d.domainFromReq(request), d.userFromReq(request)); err != nil {
		return nil, err
	}

	return p, nil
}

func decodeToMap[T any](fieldMapper *patchset.FieldMapper, request *http.Request, target *T, validate ValidatorFunc) (*patchset.PatchSet, error) {
	pr, pw := io.Pipe()
	tr := io.TeeReader(request.Body, pw)

	var wg sync.WaitGroup
	var err error
	wg.Add(1)
	go func() {
		defer wg.Done()
		err = json.NewDecoder(pr).Decode(target)
	}()

	jsonData := make(map[string]any)
	if err := json.NewDecoder(tr).Decode(&jsonData); err != nil {
		return nil, errors.Wrap(err, "failed to decode request body into map")
	}

	wg.Wait()
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal request body into struct")
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

	patchSet := patchset.NewPatchSet(changes)

	return patchSet, nil
}

func checkPermissions(ctx context.Context, patchSet *patchset.PatchSet, permissionChecker Enforcer, resourceSet *resourceset.ResourceSet, domain accesstypes.Domain, user accesstypes.User) error {
	resources := make([]accesstypes.Resource, 0, len(patchSet.Fields()))
	for _, fieldName := range patchSet.Fields() {
		if resourceSet.Contains(fieldName) {
			resources = append(resources, accesstypes.Resource(fieldName))
		}
	}

	if err := permissionChecker.RequireResources(ctx, user, domain, resourceSet.RequiredPermission(), resources...); err != nil {
		return errors.Wrap(err, "permissionChecker.RequireResource()")
	}

	return nil
}
