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
	"github.com/cccteam/ccc/patchset"
	"github.com/cccteam/ccc/resourceset"
	"github.com/go-playground/errors/v5"
)

// ValidatorFunc is a function that validates s
// It returns an error if the validation fails
type ValidatorFunc interface {
	Struct(s interface{}) error
	StructPartial(s interface{}, fields ...string) error
}

type (
	DomainFromReq func(*http.Request) accesstypes.Domain
	UserFromReq   func(*http.Request) accesstypes.User
)

// Decoder is a struct that can be used for decoding http requests and validating those requests
type Decoder[T any] struct {
	validate    ValidatorFunc
	fieldMapper *resourceset.FieldMapper
}

func NewDecoder[T any]() (*Decoder[T], error) {
	target := new(T)
	m, err := resourceset.NewFieldMapper(target)
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

func (d *Decoder[T]) WithPermissionChecker(domainFromReq DomainFromReq, userFromReq UserFromReq, enforcer accesstypes.Enforcer, rSet *resourceset.ResourceSet) *DecoderWithPermissionChecker[T] {
	return &DecoderWithPermissionChecker[T]{
		userFromReq:   userFromReq,
		domainFromReq: domainFromReq,
		enforcer:      enforcer,
		resourceSet:   rSet,
		fieldMapper:   d.fieldMapper,
	}
}

func (d *Decoder[T]) Decode(request *http.Request) (*patchset.PatchSet, error) {
	target := new(T)
	p, err := decodeToMap(d.fieldMapper, request, target, d.validate)
	if err != nil {
		return nil, err
	}

	return p, nil
}

type DecoderWithPermissionChecker[T any] struct {
	userFromReq   UserFromReq
	domainFromReq DomainFromReq
	validate      ValidatorFunc
	enforcer      accesstypes.Enforcer
	resourceSet   *resourceset.ResourceSet
	fieldMapper   *resourceset.FieldMapper
}

func (d *DecoderWithPermissionChecker[T]) WithValidator(v ValidatorFunc) *DecoderWithPermissionChecker[T] {
	decoder := *d
	decoder.validate = v

	return &decoder
}

func (d *DecoderWithPermissionChecker[T]) Decode(request *http.Request, perm accesstypes.Permission) (*patchset.PatchSet, error) {
	target := new(T)
	p, err := decodeToMap(d.fieldMapper, request, target, d.validate)
	if err != nil {
		return nil, err
	}

	if err := checkPermissions(request.Context(), p, d.enforcer, d.resourceSet, d.domainFromReq(request), d.userFromReq(request), perm); err != nil {
		return nil, err
	}

	return p, nil
}

func (d *DecoderWithPermissionChecker[T]) DecodeOperation(oper *Operation) (*patchset.PatchSet, error) {
	if oper.Type == OperationDelete {
		return nil, nil
	}

	patchSet, err := d.Decode(oper.Req, permissionFromType(oper.Type))
	if err != nil {
		return nil, errors.Wrap(err, "httpio.DecoderWithPermissionChecker[T].Decode()")
	}

	return patchSet, nil
}

func decodeToMap[T any](fieldMapper *resourceset.FieldMapper, request *http.Request, target *T, validate ValidatorFunc) (*patchset.PatchSet, error) {
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
		return nil, NewBadRequestMessageWithError(err, "failed to decode request body")
	}

	wg.Wait()
	if err != nil {
		return nil, NewBadRequestMessageWithError(err, "failed to unmarshal request body")
	}

	vValue := reflect.ValueOf(target)
	if vValue.Kind() == reflect.Ptr {
		vValue = vValue.Elem()
	}
	changes := make(map[accesstypes.Field]any)
	for jsonField := range jsonData {
		fieldName, ok := fieldMapper.StructFieldName(jsonField)
		if !ok {
			fieldName, ok = fieldMapper.StructFieldName(strings.ToLower(jsonField))
			if !ok {
				return nil, NewBadRequestMessagef("invalid field in json - %s", jsonField)
			}
		}

		value := vValue.FieldByName(string(fieldName)).Interface()
		if value == nil {
			return nil, NewBadRequestMessagef("invalid field in json - %s", jsonField)
		}

		if _, ok := changes[fieldName]; ok {
			return nil, NewBadRequestMessagef("json field name %s collides with another field name of different case", fieldName)
		}
		changes[fieldName] = value
	}

	patchSet := patchset.NewPatchSet(changes)

	if validate != nil {
		switch request.Method {
		case http.MethodPatch:
			fields := make([]string, 0, patchSet.Len())
			for _, field := range patchSet.StructFields() {
				fields = append(fields, string(field))
			}
			if err := validate.StructPartial(target, fields...); err != nil {
				return nil, NewBadRequestMessageWithError(err, "failed validating the request")
			}
		default:
			if err := validate.Struct(target); err != nil {
				return nil, NewBadRequestMessageWithError(err, "failed validating the request")
			}
		}
	}

	return patchSet, nil
}

func checkPermissions(
	ctx context.Context, patchSet *patchset.PatchSet, enforcer accesstypes.Enforcer, resourceSet *resourceset.ResourceSet,
	domain accesstypes.Domain, user accesstypes.User, perm accesstypes.Permission,
) error {
	resources := make([]accesstypes.Resource, 0, patchSet.Len()+1)
	resources = append(resources, resourceSet.BaseResource())
	for _, fieldName := range patchSet.StructFields() {
		if resourceSet.PermissionRequired(fieldName, perm) {
			resources = append(resources, resourceSet.Resource(fieldName))
		}
	}

	if ok, missing, err := enforcer.RequireResources(ctx, user, domain, perm, resources...); err != nil {
		return errors.Wrap(err, "enforcer.RequireResource()")
	} else if !ok {
		return NewForbiddenMessagef("user %s does not have %s on %s", user, perm, missing)
	}

	return nil
}
