package httpio

import (
	"net/http"

	"github.com/cccteam/ccc/accesstypes"
	"github.com/cccteam/ccc/resource"
	"github.com/go-playground/errors/v5"
)

type nilResource struct{}

func (n nilResource) Resource() accesstypes.Resource {
	return "nil"
}

// StructDecoder is a struct that can be used for decoding http requests and validating those requests
type StructDecoder[Request any] struct {
	validate    ValidatorFunc
	fieldMapper *resource.FieldMapper
	resourceSet *resource.ResourceSet[nilResource, Request]
}

func NewStructDecoder[Request any]() (*StructDecoder[Request], error) {
	target := new(Request)

	m, err := resource.NewFieldMapper(target)
	if err != nil {
		return nil, errors.Wrap(err, "NewFieldMapper()")
	}

	// TODO(jwatson): This will cause a runtime panic right now since nilResource is not convertable to Request
	// BUG(jwatson): This will cause a runtime panic right now since nilResource is not convertable to Request
	rSet, err := resource.NewResourceSet[nilResource, Request]()
	if err != nil {
		return nil, errors.Wrap(err, "NewResourceSet()")
	}

	return &StructDecoder[Request]{
		fieldMapper: m,
		resourceSet: rSet,
	}, nil
}

func (d *StructDecoder[Request]) WithValidator(v ValidatorFunc) *StructDecoder[Request] {
	decoder := *d
	decoder.validate = v

	return &decoder
}

func (d *StructDecoder[Request]) WithPermissionChecker(domainFromReq DomainFromReq, userFromReq UserFromReq, enforcer accesstypes.Enforcer) *StructDecoderWithPermissionChecker[Request] {
	return &StructDecoderWithPermissionChecker[Request]{
		userFromReq:   userFromReq,
		domainFromReq: domainFromReq,
		enforcer:      enforcer,
		resourceSet:   d.resourceSet,
		fieldMapper:   d.fieldMapper,
	}
}

// Decode parses the http request body and validates it against the struct validation rules
// and returns a named patchset
func (d *StructDecoder[Request]) Decode(request *http.Request) (*Request, error) {
	_, target, err := decodeToPatch(d.resourceSet, d.fieldMapper, request, d.validate)
	if err != nil {
		return nil, err
	}

	return target, nil
}

type StructDecoderWithPermissionChecker[Request any] struct {
	userFromReq   UserFromReq
	domainFromReq DomainFromReq
	validate      ValidatorFunc
	enforcer      accesstypes.Enforcer
	resourceSet   *resource.ResourceSet[nilResource, Request]
	fieldMapper   *resource.FieldMapper
}

func (d *StructDecoderWithPermissionChecker[Request]) WithValidator(v ValidatorFunc) *StructDecoderWithPermissionChecker[Request] {
	decoder := *d
	decoder.validate = v

	return &decoder
}

// Decode parses the http request body and validates it against the struct validation rules
func (d *StructDecoderWithPermissionChecker[Request]) Decode(request *http.Request, perm accesstypes.Permission) (*Request, error) {
	p, target, err := decodeToPatch(d.resourceSet, d.fieldMapper, request, d.validate)
	if err != nil {
		return nil, err
	}

	// TODO(jwatson): Verify this works with a nilResource
	if err := checkPermissions(request.Context(), p.Fields(), d.enforcer, d.resourceSet, d.userFromReq(request), d.domainFromReq(request), perm); err != nil {
		return nil, err
	}

	return target, nil
}
