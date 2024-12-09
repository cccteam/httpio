package httpio

import (
	"net/http"

	"github.com/cccteam/ccc/accesstypes"
	"github.com/cccteam/ccc/resource"
	"github.com/go-playground/errors/v5"
)

// StructDecoder is a struct that can be used for decoding http requests and validating those requests
type StructDecoder[Request any] struct {
	validate    ValidatorFunc
	fieldMapper *resource.FieldMapper
}

func NewStructDecoder[Request any]() (*StructDecoder[Request], error) {
	target := new(Request)

	m, err := resource.NewFieldMapper(target)
	if err != nil {
		return nil, errors.Wrap(err, "NewFieldMapper()")
	}

	return &StructDecoder[Request]{
		fieldMapper: m,
	}, nil
}

func (d *StructDecoder[Request]) WithValidator(v ValidatorFunc) *StructDecoder[Request] {
	decoder := *d
	decoder.validate = v

	return &decoder
}

func (d *StructDecoder[Request]) WithPermissionChecker(
	domainFromReq DomainFromReq, userFromReq UserFromReq, enforcer accesstypes.Enforcer, rSet *resource.ResourceSet,
) *StructDecoderWithPermissionChecker[Request] {
	return &StructDecoderWithPermissionChecker[Request]{
		userFromReq:   userFromReq,
		domainFromReq: domainFromReq,
		enforcer:      enforcer,
		resourceSet:   rSet,
		fieldMapper:   d.fieldMapper,
	}
}

// Decode parses the http request body and validates it against the struct validation rules
// and returns a named patchset
func (d *StructDecoder[Request]) Decode(request *http.Request) (*Request, error) {
	_, target, err := decodeToPatch[nilResouce, Request](d.fieldMapper, request, d.validate)
	if err != nil {
		return nil, err
	}

	return target, nil
}

type nilResouce struct{}

func (n nilResouce) Resource() accesstypes.Resource {
	return "nil"
}

type StructDecoderWithPermissionChecker[Request any] struct {
	userFromReq   UserFromReq
	domainFromReq DomainFromReq
	validate      ValidatorFunc
	enforcer      accesstypes.Enforcer
	resourceSet   *resource.ResourceSet
	fieldMapper   *resource.FieldMapper
}

func (d *StructDecoderWithPermissionChecker[Request]) WithValidator(v ValidatorFunc) *StructDecoderWithPermissionChecker[Request] {
	decoder := *d
	decoder.validate = v

	return &decoder
}

// Decode parses the http request body and validates it against the struct validation rules
func (d *StructDecoderWithPermissionChecker[Resouce]) Decode(request *http.Request, perm accesstypes.Permission) (*Resouce, error) {
	p, target, err := decodeToPatch[nilResouce, Resouce](d.fieldMapper, request, d.validate)
	if err != nil {
		return nil, err
	}

	if err := checkPermissions(request.Context(), p.Fields(), d.enforcer, d.resourceSet, d.userFromReq(request), d.domainFromReq(request), perm); err != nil {
		return nil, err
	}

	return target, nil
}
