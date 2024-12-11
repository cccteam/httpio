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
	"github.com/cccteam/ccc/resource"
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
type Decoder[Resource resource.Resourcer, Request any] struct {
	validate    ValidatorFunc
	fieldMapper *resource.FieldMapper
	resourceSet *resource.ResourceSet[Resource, Request]
}

func NewDecoder[Resource resource.Resourcer, Request any](rSet *resource.ResourceSet[Resource, Request]) (*Decoder[Resource, Request], error) {
	target := new(Request)
	m, err := resource.NewFieldMapper(target)
	if err != nil {
		return nil, errors.Wrap(err, "NewFieldMapper()")
	}

	return &Decoder[Resource, Request]{
		fieldMapper: m,
		resourceSet: rSet,
	}, nil
}

func (d *Decoder[Resource, Request]) WithValidator(v ValidatorFunc) *Decoder[Resource, Request] {
	decoder := *d
	decoder.validate = v

	return &decoder
}

func (d *Decoder[Resource, Request]) WithPermissionChecker(domainFromReq DomainFromReq, userFromReq UserFromReq, enforcer accesstypes.Enforcer) *DecoderWithPermissionChecker[Resource, Request] {
	return &DecoderWithPermissionChecker[Resource, Request]{
		userFromReq:   userFromReq,
		domainFromReq: domainFromReq,
		validate:      d.validate,
		enforcer:      enforcer,
		resourceSet:   d.resourceSet,
		fieldMapper:   d.fieldMapper,
	}
}

func (d *Decoder[Resource, Request]) Decode(request *http.Request) (*resource.PatchSet[Resource], error) {
	p, _, err := decodeToPatch(d.resourceSet, d.fieldMapper, request, d.validate)
	if err != nil {
		return nil, err
	}

	return p, nil
}

type DecoderWithPermissionChecker[Resource resource.Resourcer, Request any] struct {
	userFromReq   UserFromReq
	domainFromReq DomainFromReq
	validate      ValidatorFunc
	enforcer      accesstypes.Enforcer
	resourceSet   *resource.ResourceSet[Resource, Request]
	fieldMapper   *resource.FieldMapper
}

func (d *DecoderWithPermissionChecker[Resource, Request]) WithValidator(v ValidatorFunc) *DecoderWithPermissionChecker[Resource, Request] {
	decoder := *d
	decoder.validate = v

	return &decoder
}

func (d *DecoderWithPermissionChecker[Resource, Request]) Decode(request *http.Request, perm accesstypes.Permission) (*resource.PatchSet[Resource], error) {
	p, _, err := decodeToPatch(d.resourceSet, d.fieldMapper, request, d.validate)
	if err != nil {
		return nil, err
	}

	if err := checkPermissions(request.Context(), p.Fields(), d.enforcer, d.resourceSet, d.userFromReq(request), d.domainFromReq(request), perm); err != nil {
		return nil, err
	}

	return p, nil
}

func (d *DecoderWithPermissionChecker[Resource, Request]) DecodeOperation(oper *Operation) (*resource.PatchSet[Resource], error) {
	if oper.Type == OperationDelete {
		ctx, user, domain := oper.Req.Context(), d.userFromReq(oper.Req), d.domainFromReq(oper.Req)
		if ok, missing, err := d.enforcer.RequireResources(ctx, user, domain, accesstypes.Delete, d.resourceSet.BaseResource()); err != nil {
			return nil, errors.Wrap(err, "enforcer.RequireResource()")
		} else if !ok {
			return nil, NewForbiddenMessagef("user %s does not have %s on %s", d.userFromReq(oper.Req), accesstypes.Delete, missing)
		}

		return nil, nil
	}

	patchSet, err := d.Decode(oper.Req, permissionFromType(oper.Type))
	if err != nil {
		return nil, errors.Wrap(err, "httpio.DecoderWithPermissionChecker[Request].Decode()")
	}

	return patchSet, nil
}

func decodeToPatch[Resource resource.Resourcer, Request any](rSet *resource.ResourceSet[Resource, Request], fieldMapper *resource.FieldMapper, req *http.Request, validate ValidatorFunc) (*resource.PatchSet[Resource], *Request, error) {
	request := new(Request)
	pr, pw := io.Pipe()
	tr := io.TeeReader(req.Body, pw)

	var wg sync.WaitGroup
	var err error
	wg.Add(1)
	go func() {
		defer wg.Done()
		err = json.NewDecoder(pr).Decode(request)
	}()

	jsonData := make(map[string]any)
	if err := json.NewDecoder(tr).Decode(&jsonData); err != nil {
		return nil, nil, NewBadRequestMessageWithError(err, "failed to decode request body")
	}

	wg.Wait()
	if err != nil {
		return nil, nil, NewBadRequestMessageWithError(err, "failed to unmarshal request body")
	}

	vValue := reflect.ValueOf(request)
	if vValue.Kind() == reflect.Ptr {
		vValue = vValue.Elem()
	}

	changes := make(map[accesstypes.Field]any)
	for jsonField := range jsonData {
		fieldName, ok := fieldMapper.StructFieldName(jsonField)
		if !ok {
			fieldName, ok = fieldMapper.StructFieldName(strings.ToLower(jsonField))
			if !ok {
				return nil, nil, NewBadRequestMessagef("invalid field in json - %s", jsonField)
			}
		}

		value := vValue.FieldByName(string(fieldName)).Interface()
		if value == nil {
			return nil, nil, NewBadRequestMessagef("invalid field in json - %s", jsonField)
		}

		if _, ok := changes[fieldName]; ok {
			return nil, nil, NewBadRequestMessagef("json field name %s collides with another field name of different case", fieldName)
		}
		changes[fieldName] = value
	}

	patchSet := resource.NewPatchSet(rSet.ResourceMetadata())
	// Add to patchset in order of struct fields
	// Every key in changes is guaranteed to be a field in the struct
	for _, f := range reflect.VisibleFields(vValue.Type()) {
		field := accesstypes.Field(f.Name)
		if value, ok := changes[field]; ok {
			patchSet.Set(field, value)
		}
	}

	if validate != nil {
		switch req.Method {
		case http.MethodPatch:
			fields := make([]string, 0, patchSet.Len())
			for _, field := range patchSet.Fields() {
				fields = append(fields, string(field))
			}
			if err := validate.StructPartial(request, fields...); err != nil {
				return nil, nil, NewBadRequestMessageWithError(err, "failed validating the request")
			}
		default:
			if err := validate.Struct(request); err != nil {
				return nil, nil, NewBadRequestMessageWithError(err, "failed validating the request")
			}
		}
	}

	return patchSet, request, nil
}

func checkPermissions[Resource resource.Resourcer, Request any](
	ctx context.Context, fields []accesstypes.Field, enforcer accesstypes.Enforcer, rSet *resource.ResourceSet[Resource, Request],
	user accesstypes.User, domain accesstypes.Domain, perm accesstypes.Permission,
) error {
	resources := make([]accesstypes.Resource, 0, len(fields)+1)
	resources = append(resources, rSet.BaseResource())
	for _, fieldName := range fields {
		if rSet.PermissionRequired(fieldName, perm) {
			resources = append(resources, rSet.Resource(fieldName))
		}
	}

	if ok, missing, err := enforcer.RequireResources(ctx, user, domain, perm, resources...); err != nil {
		return errors.Wrap(err, "enforcer.RequireResource()")
	} else if !ok {
		return NewForbiddenMessagef("user %s does not have %s on %s", user, perm, missing)
	}

	return nil
}
