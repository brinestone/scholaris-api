// CRUD endpoints for tenant objects
package tenants

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
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
//encore:api auth method=POST path=/tenants tag:needs_captcha_ver
func NewTenant(ctx context.Context, req dto.NewTenantRequest) (*models.Tenant, error) {

	tx, err := tenantDb.Begin(ctx)
	if err != nil {
		return nil, err
	}

	tenant, err := createTenant(ctx, tx, req)
	if err != nil {
		rlog.Error(err.Error())
		_ = tx.Rollback()
		return nil, &util.ErrUnknown
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
		return nil, &util.ErrUnknown
	}
	_ = tx.Commit()

	NewTenants.Publish(ctx, &TenantCreated{
		Id:        tenant.Id,
		CreatedBy: creator,
	})
	return tenant, nil
}

// Find all Tenants
//
//encore:api public method=GET path=/tenants
func FindTenants(ctx context.Context, req dto.FindTenantsRequest) (*dto.PaginatedResponse[models.Tenant], error) {
	var ans = make([]*models.Tenant, 0)
	var count uint
	var err error

	if !req.SubscribedOnly {
		ans, count, err = getTenants(ctx, req.After, req.Size)
	} else {
		uid, ok := auth.UserID()
		if !ok {
			return nil, &util.ErrUnauthorized
		}
		ans, count, err = getSubscribedTenants(ctx, uid, req.After, req.Size)
	}
	if err != nil {
		rlog.Error(err.Error())
		return nil, &util.ErrUnknown
	}

	return &dto.PaginatedResponse[models.Tenant]{
		Data: ans,
		Meta: dto.PaginatedResponseMeta{
			Total: count,
		},
	}, nil
}

func getSubscribedTenants(ctx context.Context, uid auth.UID, after uint64, size uint) ([]*models.Tenant, uint, error) {
	ans := make([]*models.Tenant, 0)
	response, err := permissions.ListRelations(ctx, dto.ListRelationsRequest{
		Subject:  fmt.Sprintf("%s:%s", dto.PTUser, uid),
		Relation: "can_view",
		Type:     string(dto.PTTenant),
	})
	if err != nil {
		return ans, 0, err
	}

	if len(response.Relations) == 0 {
		return nil, 0, nil
	}

	query := fmt.Sprintf(`
		SELECT
			%s
		FROM
			tenants
		WHERE
			id IN (%s) AND id > $1
		OFFSET 0
		LIMIT $2;
	`, tenantFields, strings.Join(response.Relations[string(dto.PTTenant)], ","))

	rows, err := tenantDb.Query(ctx, query, after, size)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		var tenant = new(models.Tenant)
		if err := rows.Scan(&tenant.Id, &tenant.Name, &tenant.CreatedAt, &tenant.UpdatedAt, &tenant.Subscription); err != nil {
			return nil, 0, err
		}

		if err := tenantCache.Set(ctx, tenant.Id, *tenant); err != nil {
			return nil, 0, err
		}

		ans = append(ans, tenant)
	}

	q2 := fmt.Sprintf(`
		SELECT COUNT(id) FROM tenants WHERE id IN (%s);
	`, strings.Join(response.Relations[string(dto.PTTenant)], ","))

	var cnt uint = 0
	if err := tenantDb.QueryRow(ctx, q2).Scan(&cnt); err != nil {
		return nil, 0, err
	}

	return ans, cnt, nil
}

func getTenants(ctx context.Context, after uint64, size uint) ([]*models.Tenant, uint, error) {
	ans := make([]*models.Tenant, 0)
	rows, err := tenantDb.Query(ctx, fmt.Sprintf(`
		SELECT 
			%s 
		FROM 
			tenants 
		WHERE 
			id > $1 
		ORDER BY 
			updated_at DESC 
		OFFSET 0 
		LIMIT $2;
	`, tenantFields), after, size)

	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		var tenant = new(models.Tenant)
		if err := rows.Scan(&tenant.Id, &tenant.Name, &tenant.CreatedAt, &tenant.UpdatedAt, &tenant.Subscription); err != nil {
			return nil, 0, err
		}
		tenantCache.Set(ctx, tenant.Id, *tenant)
		ans = append(ans, tenant)
	}

	var count uint
	if err := tenantDb.QueryRow(ctx, `
		SELECT
			COUNT(*)
		FROM
			tenants;
	`).Scan(&count); err != nil {
		return ans, 0, err
	}

	return ans, count, nil
}

const tenantFields = "id,name,created_at,updated_at,subscription"

// const subscriptionPlanFields = "id,name,created_at,updated_at,price,currency,enabled,billing_cycle"

func createTenant(ctx context.Context, tx *sqldb.Tx, req dto.NewTenantRequest) (*models.Tenant, error) {
	// Check whether a tenant with the same name already exists.
	row := tx.QueryRow(ctx, "SELECT COUNT(name) AS cnt FROM tenants WHERE name=$1;", req.Name)
	var count int
	_ = row.Scan(&count)
	if count > 0 {
		return nil, &errs.Error{
			Code:    errs.AlreadyExists,
			Message: fmt.Sprintf("An organization with name \"%s\" already exists", req.Name),
		}
	}

	subId, err := createTenantSubscription(ctx, tx, req.SubscriptionPlanId)
	if err != nil {
		return nil, err
	}

	// Create the database record.
	row = tx.QueryRow(ctx, `
		INSERT INTO
			TENANTS (NAME, SUBSCRIPTION)
		VALUES ($1, $2)
		RETURNING
			ID;
	`, req.Name, *subId)
	var newId uint64

	if err := row.Scan(&newId); err != nil {
		return nil, err
	}

	// Get the tenant from the database
	var tenant = new(models.Tenant)
	row = tx.QueryRow(ctx, fmt.Sprintf("SELECT %s FROM tenants WHERE id = $1;", tenantFields), newId)
	if err := row.Scan(&tenant.Id, &tenant.Name, &tenant.CreatedAt, &tenant.UpdatedAt, &tenant.Subscription); err != nil {
		return nil, err
	}

	return tenant, nil
}

func createTenantSubscription(ctx context.Context, tx *sqldb.Tx, planId uint64) (*uint64, error) {
	var ans = new(uint64)

	// Check whether the subscription plan exists
	row := tx.QueryRow(ctx, `
		SELECT
			COUNT(SP.ID) AS CNT,
			SP.BILLING_CYCLE
		FROM
			SUBSCRIPTION_PLANS AS SP
		WHERE
			SP.ID = $1
			AND SP.ENABLED = TRUE
		GROUP BY
			SP.ID;
	`, planId)

	var count int
	var billingCycle sql.NullInt32

	if err := row.Scan(&count, &billingCycle); err != nil {
		rlog.Error(err.Error())
	}
	if count <= 0 {
		return nil, &errs.Error{
			Code:    errs.NotFound,
			Message: "No such subscription plan exists",
		}
	}

	now := time.Now()
	midnight := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	nextBillingCycle := midnight.Add(time.Hour * 24 * time.Duration(billingCycle.Int32)).UTC()

	// Create Subscription record
	row = tx.QueryRow(ctx, "INSERT INTO tenant_subscriptions(subscription_plan,next_billing_cycle) VALUES ($1,$2) RETURNING id;", planId, nextBillingCycle)
	if err := row.Scan(ans); err != nil {
		return nil, err
	}

	return ans, nil
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
	ORDER BY
		sp.price ASC
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
