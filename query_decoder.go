package httpio

import (
	"context"
	"net/http"

	"github.com/cccteam/ccc/accesstypes"
	"github.com/cccteam/ccc/resource"
	"github.com/go-playground/errors/v5"
)

type (
	DomainFromCtx func(context.Context) accesstypes.Domain
	UserFromCtx   func(context.Context) accesstypes.User
)

// QueryDecoder is a struct that returns columns that a given user has access to view
type QueryDecoder[Resource resource.Resourcer, Request any] struct {
	fieldMapper       *resource.FieldMapper
	resourceSet       *resource.ResourceSet
	permissionChecker accesstypes.Enforcer
	domainFromCtx     DomainFromCtx
	userFromCtx       UserFromCtx
}

func NewQueryDecoder[Resource resource.Resourcer, Request any](rSet *resource.ResourceSet, permissionChecker accesstypes.Enforcer, domainFromCtx DomainFromCtx, userFromCtx UserFromCtx) (*QueryDecoder[Resource, Request], error) {
	target := new(Request)

	m, err := resource.NewFieldMapper(target)
	if err != nil {
		return nil, errors.Wrap(err, "NewFieldMapper()")
	}

	return &QueryDecoder[Resource, Request]{
		fieldMapper:       m,
		resourceSet:       rSet,
		permissionChecker: permissionChecker,
		domainFromCtx:     domainFromCtx,
		userFromCtx:       userFromCtx,
	}, nil
}

func (d *QueryDecoder[Resource, Request]) Decode(request *http.Request) (*resource.Query[Resource], error) {
	fields, err := d.fields(request.Context())
	if err != nil {
		return nil, err
	}

	row := resource.NewRow[Resource]()
	qSet := resource.NewQuerySet(row)
	for _, field := range fields {
		qSet.AddField(field)
	}

	return &resource.Query[Resource]{Set: qSet}, nil
}

func (d *QueryDecoder[Resource, Request]) fields(ctx context.Context) ([]accesstypes.Field, error) {
	domain, user := d.domainFromCtx(ctx), d.userFromCtx(ctx)

	if ok, _, err := d.permissionChecker.RequireResources(ctx, user, domain, d.resourceSet.Permission(), d.resourceSet.BaseResource()); err != nil {
		return nil, errors.Wrap(err, "accesstypes.Enforcer.RequireResources()")
	} else if !ok {
		return nil, NewForbiddenMessagef("user %s does not have %s permission on %s", user, d.resourceSet.Permission(), d.resourceSet.BaseResource())
	}

	fields := make([]accesstypes.Field, 0, d.fieldMapper.Len())
	for _, field := range d.fieldMapper.Fields() {
		if !d.resourceSet.PermissionRequired(field, d.resourceSet.Permission()) {
			fields = append(fields, field)
		} else {
			if hasPerm, _, err := d.permissionChecker.RequireResources(ctx, user, domain, d.resourceSet.Permission(), d.resourceSet.Resource(field)); err != nil {
				return nil, errors.Wrap(err, "hasPermission()")
			} else if hasPerm {
				fields = append(fields, field)
			}
		}
	}

	if len(fields) == 0 {
		return nil, NewForbiddenMessagef("user %s does not have %s permission on any fields in %s", user, d.resourceSet.Permission(), d.resourceSet.BaseResource())
	}

	return fields, nil
}
