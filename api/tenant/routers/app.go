package routers

import (
	"context"

	"github.com/superwhys/rent-manager-server/model"
)

type App interface {
	CreateTenant(ctx context.Context, req *model.CreateTenantRequest) (*model.Tenant, error)
	UpdateTenant(ctx context.Context, req *model.UpdateTenantRequest) (*model.Tenant, error)
	DeleteTenant(ctx context.Context, req *model.DeleteTenantRequest) error
}
