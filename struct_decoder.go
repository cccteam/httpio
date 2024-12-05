package httpio

import (
	"net/http"

	"github.com/cccteam/ccc/accesstypes"
	"github.com/cccteam/ccc/resource"
	"github.com/go-playground/errors/v5"
)

// StructDecoder is a struct that can be used for decoding http requests and validating those requests
type StructDecoder[T any] struct {
	validate    ValidatorFunc
	fieldMapper *resource.FieldMapper
}

func NewStructDecoder[T any]() (*StructDecoder[T], error) {
	target := new(T)

	m, err := resource.NewFieldMapper(target)
	if err != nil {
		return nil, errors.Wrap(err, "NewFieldMapper()")
	}

	return &StructDecoder[T]{
		fieldMapper: m,
	}, nil
}

func (d *StructDecoder[T]) WithValidator(v ValidatorFunc) *StructDecoder[T] {
	decoder := *d
	decoder.validate = v

	return &decoder
}

func (d *StructDecoder[T]) WithPermissionChecker(
	domainFromReq DomainFromReq, userFromReq UserFromReq, enforcer accesstypes.Enforcer, rSet *resource.ResourceSet,
) *StructDecoderWithPermissionChecker[T] {
	return &StructDecoderWithPermissionChecker[T]{
		userFromReq:   userFromReq,
		domainFromReq: domainFromReq,
		enforcer:      enforcer,
		resourceSet:   rSet,
		fieldMapper:   d.fieldMapper,
	}
}

// Decode parses the http request body and validates it against the struct validation rules
// and returns a named patchset
func (d *StructDecoder[T]) Decode(request *http.Request) (*T, error) {
	target := new(T)
	if _, err := decodeToPatch(d.fieldMapper, request, target, d.validate); err != nil {
		return nil, err
	}

	return target, nil
}

type StructDecoderWithPermissionChecker[T any] struct {
	userFromReq   UserFromReq
	domainFromReq DomainFromReq
	validate      ValidatorFunc
	enforcer      accesstypes.Enforcer
	resourceSet   *resource.ResourceSet
	fieldMapper   *resource.FieldMapper
}

func (d *StructDecoderWithPermissionChecker[T]) WithValidator(v ValidatorFunc) *StructDecoderWithPermissionChecker[T] {
	decoder := *d
	decoder.validate = v

	return &decoder
}

// Decode parses the http request body and validates it against the struct validation rules
func (d *StructDecoderWithPermissionChecker[T]) Decode(request *http.Request, perm accesstypes.Permission) (*T, error) {
	target := new(T)
	p, err := decodeToPatch(d.fieldMapper, request, target, d.validate)
	if err != nil {
		return nil, err
	}

	if err := checkPermissions(request.Context(), p, d.enforcer, d.resourceSet, d.userFromReq(request), d.domainFromReq(request), perm); err != nil {
		return nil, err
	}

	return target, nil
}
