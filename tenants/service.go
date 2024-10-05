// CRUD endpoints for tenant objects
package tenants

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"encore.dev/beta/auth"
	"encore.dev/beta/errs"
	"encore.dev/rlog"
	"encore.dev/storage/cache"
	"encore.dev/storage/sqldb"
	"github.com/brinestone/scholaris/core/permissions"
	"github.com/brinestone/scholaris/dto"
	"github.com/brinestone/scholaris/models"
	"github.com/brinestone/scholaris/util"
)

type FindTenantsResponse struct {
	Tenants []*models.Tenant `json:"tenants"`
}

type FindSubscriptionPlansResponse struct {
	SubscriptionPlans []*models.SubscriptionPlan `json:"plans"`
}

// Finds Subscription plans
//
//encore:api public method=GET path=/subscription-plans
func FindSubscriptionPlans(ctx context.Context, params dto.PaginationParams) (*FindSubscriptionPlansResponse, error) {
	plans, err := findSubscriptionPlans(ctx, params)
	if err != nil {
		rlog.Error(err.Error())
		return nil, &util.ErrUnknown
	}

	return &FindSubscriptionPlansResponse{
		SubscriptionPlans: plans,
	}, nil
}

// Finds a tenant using its ID
//
//encore:api auth method=GET path=/tenants/:id
func FindTenant(ctx context.Context, id uint64) (*models.Tenant, error) {
	var t *models.Tenant
	var err error

	t, err = findTenantById(ctx, id)
	if err != nil {
		if errors.Is(err, sqldb.ErrNoRows) {
			return nil, &util.ErrNotFound
		}
		rlog.Error(err.Error())
		return nil, &util.ErrUnknown
	}

	return t, err
}

// Deletes a Tenant
//
//encore:api auth method=DELETE path=/tenants/:id tag:perm_can_delete_tenant
func DeleteTenant(ctx context.Context, id uint64) error {
	tx, err := tenantDb.Begin(ctx)
	if err != nil {
		return err
	}

	err = deleteTenantById(ctx, tx, id)
	if err != nil {
		rlog.Error(err.Error())
		_ = tx.Rollback()
		return &util.ErrUnknown
	}

	userId, _ := auth.UserID()

	timeout, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	if err = permissions.DeletePermissions(timeout, dto.UpdatePermissionsRequest{
		Updates: []dto.PermissionUpdate{
			{
				Subject:  fmt.Sprintf("user:%s", userId),
				Relation: "owner",
				Target:   fmt.Sprintf("tenant:%d", id),
			},
		},
	}); err != nil {
		rlog.Error(err.Error())
		_ = tx.Rollback()
		return &util.ErrUnknown
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

	tenant, err := createTenant(ctx, tx, req)
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
				Subject:  fmt.Sprintf("user:%s", string(*creator)),
				Relation: models.PermOwner,
				Target:   fmt.Sprintf("tenant:%d", tenant.Id),
			},
		},
	}); err != nil {
		rlog.Error(err.Error())
		err = tx.Rollback()
		rlog.Error(err.Error())
		return &util.ErrUnknown
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
func FindTenants(ctx context.Context, req dto.PaginationParams) (*FindTenantsResponse, error) {
	ans, err := findTenants(ctx, req)
	if err != nil {
		return nil, err
	}

	return &FindTenantsResponse{
		Tenants: ans,
	}, nil
}

func findTenants(ctx context.Context, req dto.PaginationParams) ([]*models.Tenant, error) {
	ans := make([]*models.Tenant, 0)
	rows, err := tenantDb.Query(ctx, fmt.Sprintf("SELECT %s FROM tenants WHERE id > $1 ORDER BY updated_at DESC OFFSET 0 LIMIT $2;", tenantFields), req.After, req.Size)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var tenant = new(models.Tenant)
		if err := rows.Scan(&tenant.Id, &tenant.Name, &tenant.CreatedAt, &tenant.UpdatedAt, &tenant.Subscription); err != nil {
			return nil, err
		}
		ans = append(ans, tenant)
	}
	return ans, nil
}

const tenantFields = "id,name,created_at,updated_at,subscription"
const subscriptionPlanFields = "id,name,created_at,updated_at,price,currency,enabled,billing_cycle"

func createTenant(ctx context.Context, tx *sqldb.Tx, req dto.NewTenantRequest) (*models.Tenant, error) {
	now := time.Now()

	// Check whether a tenant with the same name already exists.
	row := tx.QueryRow(ctx, "SELECT COUNT(name) as cnt FROM tenants WHERE name=$1;", req.Name)
	var count int
	_ = row.Scan(&count)
	if count > 0 {
		return nil, &errs.Error{
			Code:    errs.AlreadyExists,
			Message: fmt.Sprintf("An organization with name \"%s\" already exists", req.Name),
		}
	}

	// Create the database record.
	_, err := tx.Exec(ctx, "INSERT INTO tenants(name,created_at,updated_at) VALUES ($1,$2,$3);", req.Name, now, now)
	if err != nil {
		return nil, err
	}

	// Get the tenant from the database
	var tenant = new(models.Tenant)
	row = tx.QueryRow(ctx, fmt.Sprintf("SELECT %s FROM tenants WHERE created_at=$1 AND name=$2;", tenantFields), now, req.Name)
	if err := row.Scan(&tenant.Id, &tenant.Name, &tenant.CreatedAt, &tenant.UpdatedAt, &tenant.Subscription); err != nil {
		return nil, err
	}

	return tenant, nil
}

func deleteTenantById(ctx context.Context, tx *sqldb.Tx, id uint64) error {
	query := "DELETE FROM tenants WHERE id = $1;"

	cnt, err := tx.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	rlog.Info(fmt.Sprintf("deleted %d record(s)", cnt.RowsAffected()))

	return nil
}

func findTenantById(ctx context.Context, id uint64) (*models.Tenant, error) {
	var t *models.Tenant
	var err error

	t, err = findTenantByIdFromCache(ctx, id)
	if err != nil {
		if errors.Is(err, cache.Miss) {
			t, err = findTenantByIdFromDb(ctx, id)
		} else {
			return nil, err
		}
	}

	return t, err
}

func findTenantByIdFromDb(ctx context.Context, id uint64) (*models.Tenant, error) {
	query := fmt.Sprintf("SELECT %s FROM tenants WHERE id = $1;", tenantFields)
	row := tenantDb.QueryRow(ctx, query, id)
	var t = new(models.Tenant)

	if err := row.Scan(&t.Id, &t.Name, &t.CreatedAt, &t.UpdatedAt, &t.Subscription); err != nil {
		return nil, err
	}

	_ = tenantCache.Set(ctx, id, *t)
	return t, nil
}

func findTenantByIdFromCache(ctx context.Context, id uint64) (*models.Tenant, error) {
	t, err := tenantCache.Get(ctx, id)

	if err != nil {
		return nil, err
	}
	return &t, nil
}

func findSubscriptionPlans(ctx context.Context, params dto.PaginationParams) ([]*models.SubscriptionPlan, error) {
	ans := make([]*models.SubscriptionPlan, 0)
	query := `
	SELECT
    	sp.id,
    	sp.name,
    	sp.created_at AS "createdAt",
    	sp.updated_at AS "updatedAt",
    	sp.price,
    	sp.currency,
    	sp.enabled,
    	sp.billing_cycle AS "billingCycle",
    	COALESCE(json_agg(json_build_object(
        	'name', spd.name,
        	'details', spd.details,
			'maxCount', spd.max_count,
			'minCount', spd.min_count,
			'maxCount', spd.max_count
    	)) FILTER (WHERE spd.id IS NOT NULL), '[]') AS "descriptions"
	FROM
    	subscription_plans sp
	LEFT JOIN
    	plan_benefits spd
	ON
    	sp.id = spd.subscription_plan
	WHERE
		sp.id > $1 AND sp.enabled = true
	GROUP BY
    	sp.id
	OFFSET 0
	LIMIT $2;
	`

	rows, err := tenantDb.Query(ctx, query, params.After, params.Size)
	if err != nil {
		return ans, err
	}
	defer rows.Close()

	for rows.Next() {
		plan := new(models.SubscriptionPlan)
		var benefitsJson string
		if err := rows.Scan(&plan.Id, &plan.Name, &plan.CreatedAt, &plan.UpdatedAt, &plan.Price, &plan.Currency, &plan.Enabled, &plan.BillingCycle, &benefitsJson); err != nil {
			if !errors.Is(err, sqldb.ErrNoRows) {
				return ans, err
			} else {
				break
			}
		}

		var benefits []models.SubscriptionPlanBenefit
		if err := json.Unmarshal([]byte(benefitsJson), &benefits); err != nil {
			return ans, err
		}
		plan.Benefits = &benefits
		ans = append(ans, plan)
	}

	return ans, nil
}
