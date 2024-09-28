package tenants

import (
	"context"
	"fmt"
	"time"

	"encore.dev/beta/auth"
	"encore.dev/beta/errs"
	"encore.dev/rlog"
	"github.com/brinestone/scholaris/core/permissions"
	"github.com/brinestone/scholaris/dto"
	"github.com/brinestone/scholaris/models"
)

// var secrets struct {
// 	DatabaseUri string
// }

type TenantsResponse struct {
	Tenants []*models.Tenant `json:"tenants"`
}

// Deletes a Tenant
//
//encore:api auth method=DELETE path=/tenants/:id
func DeleteTenant(ctx context.Context, id uint64) error {
	err := deleteTenantById(ctx, id)
	if err != nil {
		rlog.Error(err.Error())
		return &errs.Error{
			Message: "An Unknown error occured",
			Code:    errs.Internal,
		}
	}

	DeletedTenants.Publish(ctx, &TenantDeleted{
		Id:        id,
		DeletedAt: time.Now(),
	})
	return nil
}

// Creates a new Tenant
//
//encore:api auth method=POST path=/tenants
func NewTenant(ctx context.Context, req dto.NewTenantRequest) error {

	tx, err := tenantDb.Begin(ctx)
	if err != nil {
		return err
	}

	tenant, err := createTenant(ctx, req)
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	var creator *auth.UID

	if i, ok := auth.UserID(); ok {
		creator = &i
	}

	if err = permissions.SetPermissions(ctx, dto.UpdatePermissionsRequest{
		Updates: []dto.PermissionUpdate{
			{
				User:     fmt.Sprintf("user:%s", string(*creator)),
				Relation: models.PermOwner,
				Object:   fmt.Sprintf("tenant:%d", tenant.Id),
			},
		},
	}); err != nil {
		_ = tx.Rollback()
		rlog.Error(err.Error())
		return err
	}
	defer tx.Commit()

	NewTenants.Publish(ctx, &TenantCreated{
		Id:        tenant.Id,
		CreatedBy: creator,
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

const tenantFields = "id,name,created_at,updated_at"

func createTenant(ctx context.Context, req dto.NewTenantRequest) (*models.Tenant, error) {
	now := time.Now()

	// Check whether a tenant with the same name already exists.
	row := tenantDb.QueryRow(ctx, "SELECT COUNT(*) as cnt FROM tenants WHERE name=$1;", req.Name)
	var count int
	_ = row.Scan(&count)
	if count > 0 {
		return nil, &errs.Error{
			Code:    errs.AlreadyExists,
			Message: fmt.Sprintf("An organization with name \"%s\" already exists", req.Name),
		}
	}

	// Create the database record.
	_, err := tenantDb.Exec(ctx, "INSERT INTO tenants(name,created_at,updated_at) VALUES ($1,$2,$3);", req.Name, now, now)
	if err != nil {
		return nil, err
	}

	// Get the tenant from the database
	var tenant = new(models.Tenant)
	row = tenantDb.QueryRow(ctx, fmt.Sprintf("SELECT %s FROM tenants WHERE created_at=$1 AND name=$2;", tenantFields), now, req.Name)
	if err := row.Scan(&tenant.Id, &tenant.Name, &tenant.CreatedAt, &tenant.UpdatedAt); err != nil {
		return nil, err
	}

	return tenant, nil
}

func deleteTenantById(ctx context.Context, id uint64) error {
	query := "DELETE FROM tenants WHERE id = $1;"
	tx, err := tenantDb.Begin(ctx)
	if err != nil {
		return err
	}

	cnt, err := tenantDb.Exec(ctx, query, id)
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	_ = tx.Commit()
	rlog.Info(fmt.Sprintf("deleted %d record(s)", cnt.RowsAffected()))

	return nil
}
