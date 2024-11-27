// CRUD endpoints for tenant objects
package tenants

import (
	"context"
	"database/sql"
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
	"github.com/lib/pq"
)

// Finds Subscription plans
//
//encore:api public method=GET path=/subscription-plans
func FindSubscriptionPlans(ctx context.Context) (*dto.FindSubscriptionPlansResponse, error) {
	plans, err := findSubscriptionPlans(ctx, 0, 1000000, 0)
	if err != nil {
		rlog.Error(err.Error())
		return nil, &util.ErrUnknown
	}

	return &dto.FindSubscriptionPlansResponse{
		Plans: subscriptionPlansToDto(plans...),
	}, nil
}

// Finds a tenant using its ID
//
//encore:api auth method=GET path=/tenants/:id tag:can_view_tenant
func FindTenant(ctx context.Context, id uint64) (*dto.TenantLookup, error) {
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

	return &tenantsToDto(t)[0], err
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
				Actor:    fmt.Sprintf("user:%s", userId),
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
func NewTenant(ctx context.Context, req dto.NewTenantRequest) (err error) {

	tx, err := tenantDb.Begin(ctx)
	if err != nil {
		rlog.Error(util.MsgDbAccessError, "err", err)
		err = &util.ErrUnknown
		return
	}

	user, _ := auth.UserID()

	tenant, err := createTenant(ctx, tx, req, &user)
	if err != nil {
		tx.Rollback()
		rlog.Error(util.MsgDbAccessError, "err", err)
		err = &util.ErrUnknown
		return
	}

	tx.Commit()

	NewTenants.Publish(ctx, &TenantCreated{
		Id:        tenant.Id,
		CreatedBy: &user,
	})
	return
}

// Find all Tenants
//
//encore:api auth method=GET path=/tenants
func Lookup(ctx context.Context, req dto.PageBasedPaginationParams) (ans *dto.FindTenantResponse, err error) {
	uid, _ := auth.UserID()

	viewable, err := lookupViewableTenantIds(ctx, uid)
	if err != nil {
		rlog.Error(util.MsgCallError, "msg", err.Error())
		err = &util.ErrUnknown
		return
	}

	if len(viewable) == 0 {
		err = &util.ErrNotFound
		return
	}

	found, err := findViewableTenants(ctx, req.Page, req.Size, viewable)
	if errors.Is(err, sql.ErrNoRows) {
		err = &util.ErrNotFound
		return
	} else if err != nil {
		rlog.Error(util.MsgDbAccessError, "msg", err.Error())
		err = &util.ErrUnknown
		return
	}

	ans = &dto.FindTenantResponse{
		Tenants: tenantsToDto(found...),
	}

	return
}

func lookupViewableTenantIds(ctx context.Context, uid auth.UID) (ans []uint64, err error) {
	response, err := permissions.ListRelations(ctx, dto.ListRelationsRequest{
		Actor:    dto.IdentifierString(dto.PTUser, uid),
		Relation: models.PermCanView,
		Type:     string(dto.PTTenant),
	})
	if err != nil {
		return
	}

	ans = response.Relations[dto.PTTenant]
	return
}

func findViewableTenants(ctx context.Context, page, size uint, ids []uint64) (ans []*models.Tenant, err error) {
	query := "SELECT id,name,created_at,updated_at FROM vw_AllTenants WHERE id=ANY($1) OFFSET=$2 LIMIT $3;"

	rows, err := tenantDb.Query(ctx, query, pq.Array(ids), page*size, size)
	if err != nil {
		rlog.Debug("here")
		return
	}
	defer rows.Close()

	for rows.Next() {
		var mod = new(models.Tenant)
		if err = rows.Scan(&mod.Id, &mod.Name, &mod.CreatedAt, &mod.UpdatedAt); err != nil {
			return
		}
		ans = append(ans, mod)
	}

	return
}

const tenantFields = "id,name,created_at,updated_at,subscription"

func createTenant(ctx context.Context, tx *sqldb.Tx, req dto.NewTenantRequest, owner *auth.UID) (*models.Tenant, error) {
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

	if err = permissions.SetPermissions(ctx, dto.UpdatePermissionsRequest{
		Updates: []dto.PermissionUpdate{
			{
				Actor:    fmt.Sprintf("%s:%s", dto.PTUser, string(*owner)),
				Relation: models.PermOwner,
				Target:   fmt.Sprintf("%s:%d", dto.PTTenant, tenant.Id),
			},
			{
				Actor:    fmt.Sprintf("%s:%d", dto.PTTenant, tenant.Id),
				Relation: models.PermOwner,
				Target:   fmt.Sprintf("%s:%d", dto.PTSubscription, *subId),
			},
		},
	}); err != nil {
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

func findSubscriptionPlans(ctx context.Context, page, size uint, cursor uint64) (ans []*models.SubscriptionPlan, err error) {
	query := `
		SELECT 
			* 
		FROM
			vw_AllSubscriptionPlans
		WHERE
			enabled=true AND id > $1 OFFSET $2 LIMIT $3
		;
	`

	rows, err := tenantDb.Query(ctx, query, cursor, page*size, size)
	if err != nil {
		return
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

		plan.Benefits = benefits
		ans = append(ans, plan)
	}

	return ans, nil
}

func subscriptionPlansToDto(plans ...*models.SubscriptionPlan) (ans []dto.SubscriptionPlan) {
	for _, plan := range plans {
		var t = dto.SubscriptionPlan{
			Id:           plan.Id,
			Name:         plan.Name,
			CreatedAt:    plan.CreatedAt,
			UpdatedAt:    plan.UpdatedAt,
			Enabled:      plan.Enabled,
			Benefits:     make([]dto.SubscriptionPlanBenefit, len(plan.Benefits)),
			BillingCycle: plan.BillingCycle,
		}

		if plan.Currency.Valid {
			t.Currency = &plan.Currency.String
		}

		if plan.Price.Valid {
			t.Price = &plan.Price.Float64
		}

		for i, v := range plan.Benefits {
			t.Benefits[i] = dto.SubscriptionPlanBenefit{
				Name:     v.Name,
				Details:  v.Details,
				MinCount: v.MinCount,
				MaxCount: v.MaxCount,
			}
		}

		ans = append(ans, t)
	}
	return
}

func tenantsToDto(t ...*models.Tenant) (ans []dto.TenantLookup) {
	ans = make([]dto.TenantLookup, len(t))

	for i, v := range t {
		vv := dto.TenantLookup{
			Name:      v.Name,
			Id:        v.Id,
			CreatedAt: v.CreatedAt,
			UpdatedAt: v.UpdatedAt,
		}
		ans[i] = vv
	}

	return
}
