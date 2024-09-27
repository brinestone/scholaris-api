package tenants

import (
	"context"
	"time"

	"encore.dev/beta/errs"
	"github.com/brinestone/scholaris/models"
)

// var secrets struct {
// 	DatabaseUri string
// }

type TenantsResponse struct {
	Tenants []*models.Tenant `json:"tenants"`
}

// Creates a new Tenant
//
//encore:api public method=POST path=/tenants
func NewTenant(ctx context.Context, dto models.NewTenantRequest) error {
	if len(dto.Name) == 0 {
		return &errs.Error{
			Code:    errs.InvalidArgument,
			Message: "invalid value for: \"name\"",
		}
	}

	now := time.Now()
	_, err := tenantDb.Exec(ctx, "INSERT INTO tenants(name,created_at,updated_at) VALUES ($1,$2,$3);", dto.Name, now, now)
	if err != nil {
		return err
	}

	NewTenantsTopic.Publish(ctx, &TenantCreated{
		Date: now,
	})
	return nil
}

// Find all Tenants
//
//encore:api public method=GET path=/tenants
func FindTenants(ctx context.Context) (*TenantsResponse, error) {
	ans := make([]*models.Tenant, 0)
	rows, err := tenantDb.Query(ctx, "SELECT id,name,created_at,updated_at FROM tenants;")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var tenant = new(models.Tenant)
		if err := rows.Scan(&tenant.Id, &tenant.Name, &tenant.CreatedAt, &tenant.UpdatedAt); err != nil {
			return nil, err
		}
		ans = append(ans, tenant)
	}

	return &TenantsResponse{
		Tenants: ans,
	}, nil
}
