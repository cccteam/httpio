// package columnset provides types to store columns that a given user has access to view
package columnset

import (
	"context"

	"github.com/cccteam/ccc/accesstypes"
	"github.com/cccteam/httpio"
	"github.com/cccteam/httpio/resourceset"
	"github.com/go-playground/errors/v5"
)

type (
	DomainFromCtx func(context.Context) accesstypes.Domain
	UserFromCtx   func(context.Context) accesstypes.User
)

type ColumnSet interface {
	StructFields(ctx context.Context) ([]string, error)
}

type Enforcer interface {
	RequireResources(ctx context.Context, user accesstypes.User, domain accesstypes.Domain, perms accesstypes.Permission, resources ...accesstypes.Resource) error
}

// columnSet is a struct that returns columns that a given user has access to view
type columnSet[T any] struct {
	fieldMapper       *resourceset.FieldMapper
	resourceSet       *resourceset.ResourceSet
	permissionChecker Enforcer
	domainFromCtx     DomainFromCtx
	userFromCtx       UserFromCtx
}

func NewColumnSet[T any](rSet *resourceset.ResourceSet, permissionChecker Enforcer, domainFromCtx DomainFromCtx, userFromCtx UserFromCtx) (ColumnSet, error) {
	target := new(T)

	m, err := resourceset.NewFieldMapper(target)
	if err != nil {
		return nil, errors.Wrap(err, "NewFieldMapper()")
	}

	return &columnSet[T]{
		fieldMapper:       m,
		resourceSet:       rSet,
		permissionChecker: permissionChecker,
		domainFromCtx:     domainFromCtx,
		userFromCtx:       userFromCtx,
	}, nil
}

func (p *columnSet[T]) StructFields(ctx context.Context) ([]string, error) {
	fields := make([]string, 0, p.fieldMapper.Len())
	domain, user := p.domainFromCtx(ctx), p.userFromCtx(ctx)
	for _, field := range p.fieldMapper.Fields() {
		if !p.resourceSet.Contains(field) {
			fields = append(fields, field)
		} else {
			if hasPerm, err := hasPermission(ctx, p.permissionChecker, p.resourceSet, domain, user, p.resourceSet.Resource(field)); err != nil {
				return nil, errors.Wrap(err, "hasPermission()")
			} else if hasPerm {
				fields = append(fields, field)
			}
		}
	}

	return fields, nil
}

func hasPermission(ctx context.Context, enforcer Enforcer, resourceSet *resourceset.ResourceSet, domain accesstypes.Domain, user accesstypes.User, resource accesstypes.Resource) (bool, error) {
	if err := enforcer.RequireResources(ctx, user, domain, resourceSet.RequiredPermission(), resource); err != nil {
		if httpio.CauseIsError(err) {
			return false, errors.Wrap(err, "Enforcer.RequireResources()")
		}

		return false, nil
	}

	return true, nil
}
