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
	"github.com/brinestone/scholaris/helpers"
	"github.com/brinestone/scholaris/models"
	"github.com/brinestone/scholaris/util"
	"github.com/lib/pq"
)

// Invites a new member to a tenant
//
//encore:api auth method=POST path=/tenants/invites/:tenant tag:can_modify_tenant_members
func InviteNewMember(ctx context.Context, tenant uint64, req dto.CreateTenantInviteRequest) (err error) {
	tx, err := tenantDb.Begin(ctx)
	if err != nil {
		rlog.Error(util.MsgDbAccessError, "err", err)
		err = &util.ErrUnknown
		return
	}

	invite, err := createTenantInvite(ctx, tx, dto.PNCanAddMaintainer, req.Email, req.Phone, &req.Names, nil, 7*24*time.Hour, tenant, nil)
	if err != nil {
		tx.Rollback()
		rlog.Error(util.MsgDbAccessError, "err", err)
		err = &util.ErrUnknown
		return
	}

	tx.Commit()

	inviteObj, err := findInviteById(ctx, invite)
	if err != nil {
		rlog.Error(util.MsgDbAccessError, "err", err)
		err = &util.ErrUnknown
		return
	}

	TenantInvites.Publish(ctx, &MemberInvited{
		Id:          invite,
		Email:       req.Email,
		DisplayName: req.Names,
		TenantName:  inviteObj.TenantName,
	})
	return
}

// Gets the members of a tenant
//
//encore:api auth method=GET path=/tenants/members/:id tag:can_view_tenant_members
func FindMembers(ctx context.Context, id uint64) (ans *dto.FindTenantMembersResponse, err error) {
	members, err := findTenantMemberships(ctx, id)
	if errors.Is(err, sqldb.ErrNoRows) {
		err = &util.ErrNotFound
		return
	} else if err != nil {
		rlog.Error(util.MsgDbAccessError, "err", err)
		err = &util.ErrUnknown
		return
	}

	ans = &dto.FindTenantMembersResponse{
		Members: tenantMembershipsToDto(members...),
	}
	return
}

// Checks whether a tenant name exists or not
//
//encore:api public method=GET path=/tenants/name-available
func NameAvailable(ctx context.Context, req dto.TenantNameAvailableRequest) (ans dto.TenantNameAvailableResponse, err error) {
	exists, err := tenantNameExists(ctx, req.Name)
	if err != nil {
		rlog.Error(util.MsgDbAccessError, "err", err)
		err = &util.ErrUnknown
		return
	}

	ans.Available = !exists
	return
}

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
//encore:api auth method=GET path=/tenants/find/:id tag:can_view_tenant
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
func NewTenant(ctx context.Context, req dto.NewTenantRequest) (ans dto.NewTenantResponse, err error) {
	user, _ := auth.UserID()
	userInfo := auth.Data().(*dto.AuthClaims)

	// Check whether a tenant with the same name already exists or not
	nameUnavailable, err := tenantNameExists(ctx, req.Name)
	if err != nil {
		rlog.Error("error while checking tenant name existence", "err", err)
		err = &util.ErrUnknown
		return
	}

	if nameUnavailable {
		err = &errs.Error{
			Code:    errs.AlreadyExists,
			Message: fmt.Sprintf("The name: \"%s\" is not available. Please use another", req.Name),
		}
		return
	}

	// Start DB Transaction
	tx, err := tenantDb.Begin(ctx)
	if err != nil {
		rlog.Error("transaction error", "err", err)
		err = &util.ErrUnknown
		return
	}

	// Create Subscription
	subId, err := createTenantSubscription(ctx, tx, 1) // Use basic plan by default
	if err != nil {
		tx.Rollback()
		rlog.Error("error while creating tenant subscription", "err", err)
		err = &util.ErrUnknown
		return
	}

	// Create Tenant Record
	tenant, err := createTenant(ctx, tx, req, subId)
	if err != nil {
		tx.Rollback()
		rlog.Error("error while creating tenant", "err", err)
		err = &util.ErrUnknown
		return
	}

	// Create invite record for the owner user
	inviteId, err := createTenantInvite(ctx, tx, dto.PNOwner, userInfo.Email, userInfo.Phone, &userInfo.FullName, userInfo.Avatar, time.Hour*24*7, tenant, &userInfo.Sub)
	if err != nil {
		tx.Rollback()
		rlog.Error(util.MsgDbAccessError, "err", err)
		err = &util.ErrUnknown
		return
	}

	// Create the membership record for the just created invite.
	err = createTenantMembership(ctx, tx, inviteId, dto.PNOwner, userInfo.Email, userInfo.FullName, userInfo.Phone, userInfo.Avatar, nil)
	if err != nil {
		tx.Rollback()
		rlog.Error(util.MsgDbAccessError, "err", err)
		err = &util.ErrUnknown
		return
	}

	// Create permission/role tuples
	if err = permissions.SetPermissions(ctx, dto.UpdatePermissionsRequest{
		Updates: []dto.PermissionUpdate{
			{
				Actor:    dto.IdentifierString(dto.PTUser, user),
				Relation: dto.PNOwner,
				Target:   dto.IdentifierString(dto.PTTenant, tenant),
			},
			{
				Actor:    dto.IdentifierString(dto.PTTenant, tenant),
				Relation: dto.PNOwner,
				Target:   dto.IdentifierString(dto.PTSubscription, subId),
			},
		},
	}); err != nil {
		tx.Rollback()
		rlog.Error(util.MsgCallError, "err", err)
		err = &util.ErrUnknown
		return
	}

	// Commit the transaction
	tx.Commit()

	NewTenants.Publish(ctx, &TenantCreated{
		Id:        tenant,
		CreatedBy: &user,
	})

	ans.Id = tenant
	return
}

// Find all Tenants
//
//encore:api auth method=GET path=/tenants
func Lookup(ctx context.Context) (ans *dto.FindTenantResponse, err error) {
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

	found, err := findViewableTenants(ctx, viewable)
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
	response, err := permissions.ListObjectsInternal(ctx, dto.ListObjectsRequest{
		Actor:    dto.IdentifierString(dto.PTUser, uid),
		Relation: dto.PNCanView,
		Type:     string(dto.PTTenant),
	})
	if err != nil {
		return
	}

	ans = response.Relations[dto.PTTenant]
	return
}

func findViewableTenants(ctx context.Context, ids []uint64) (ans []*models.Tenant, err error) {
	query := `
		SELECT 
			id, name, created_at, updated_at, subscription_plan_name
		FROM 
			vw_AllTenants
		WHERE 
			id = ANY(SELECT * FROM UNNEST($1::BIGINT[]))
		ORDER BY 
			created_at DESC
		;
	`

	rows, err := tenantDb.Query(ctx, query, pq.Array(ids))
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		var mod = new(models.Tenant)
		if err = rows.Scan(&mod.Id, &mod.Name, &mod.CreatedAt, &mod.UpdatedAt, &mod.SubscriptionName); err != nil {
			return
		}
		ans = append(ans, mod)
	}

	return
}

const tenantFields = "id,name,created_at,updated_at,subscription"

func tenantNameExists(ctx context.Context, name string) (ans bool, err error) {
	query := "SELECT COUNT(id) FROM tenants WHERE LOWER(name) = LOWER($1);"
	var cnt int
	if err = tenantDb.QueryRow(ctx, query, name).Scan(&cnt); err != nil {
		return
	}
	ans = cnt > 0

	return
}

func createTenant(ctx context.Context, tx *sqldb.Tx, req dto.NewTenantRequest, subId uint64) (id uint64, err error) {
	query := "INSERT INTO tenants(name,subscription) VALUES ($1,$2) RETURNING id;"
	err = tx.QueryRow(ctx, query, req.Name, subId).Scan(&id)
	return
}

func createTenantSubscription(ctx context.Context, tx *sqldb.Tx, planId uint64) (id uint64, err error) {
	// Check whether the subscription plan exists
	row := tx.QueryRow(ctx, `
		SELECT
			SP.BILLING_CYCLE
		FROM
			SUBSCRIPTION_PLANS AS SP
		WHERE
			SP.ID = $1
			AND SP.ENABLED = TRUE;
	`, planId)

	var billingCycle sql.NullInt32

	if err = row.Scan(&billingCycle); err != nil {
		return
	}

	now := time.Now()
	midnight := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	nextBillingCycle := midnight.Add(time.Hour * 24 * time.Duration(billingCycle.Int32)).UTC()

	// Create Subscription record
	row = tx.QueryRow(ctx, "INSERT INTO tenant_subscriptions(subscription_plan,next_billing_cycle) VALUES ($1,$2) RETURNING id;", planId, nextBillingCycle)
	if err = row.Scan(&id); err != nil {
		return
	}

	return
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
			Name:             v.Name,
			Id:               v.Id,
			CreatedAt:        v.CreatedAt,
			UpdatedAt:        v.UpdatedAt,
			SubscriptionPlan: v.SubscriptionName,
		}
		ans[i] = vv
	}

	return
}

func tenantMembershipsToDto(m ...*models.TenantMembership) (ans []dto.TenantMembership) {
	ans = helpers.SliceMap(m, func(m *models.TenantMembership) dto.TenantMembership {
		d := dto.TenantMembership{
			Invite:       m.Invite,
			User:         m.User,
			DisplayName:  m.DisplayName,
			Email:        m.Email,
			InviteStatus: m.InviteStatus,
			Role:         m.Role,
			InvitedAt:    m.InvitedAt,
			Tenant:       m.Tenant,
		}

		if m.UpdatedAt.Valid {
			d.UpdatedAt = &m.UpdatedAt.Time
		}

		if m.CreatedAt.Valid {
			d.JoinedAt = &m.CreatedAt.Time
		}

		if m.InviteExpiresAt != nil && m.InviteExpiresAt.Valid {
			d.InviteExpiresAt = &m.InviteExpiresAt.Time
		}

		if m.Prefs != nil && len(*m.Prefs) > 0 {
			d.Prefs = m.Prefs
		}

		if m.Phone.Valid {
			d.Phone = &m.Phone.String
		}

		if m.Avatar.Valid {
			d.Avatar = &m.Avatar.String
		}

		if m.Id.Valid {
			tmp := uint64(m.Id.Int64)
			d.Id = &tmp
		}

		return d
	})
	return
}

func checkPermissions(ctx context.Context, actor, target string, relation dto.PermissionName) (pass bool, err error) {
	res, err := permissions.CheckPermissionInternal(ctx, dto.InternalRelationCheckRequest{
		Actor:    actor,
		Relation: relation,
		Target:   target,
	})

	if err != nil {
		err = errs.Wrap(err, util.MsgCallError)
		return
	}

	pass = res.Allowed
	return
}

func scanTenantInvitation(s util.RowScanner) (ans *models.TenantMembershipInvitation, err error) {
	ans = new(models.TenantMembershipInvitation)
	err = s.Scan(&ans.Id, &ans.User, &ans.Tenant, &ans.TenantName, &ans.Email, &ans.Phone, &ans.Role, &ans.DisplayName, &ans.RedirectUrl, &ans.ErrorRedirect, &ans.OnboardRedirect, &ans.Url, &ans.Avatar, &ans.CreatedAt, &ans.UpdatedAt, &ans.Status, &ans.ExpiresAt)
	if err != nil {
		err = errs.Wrap(err, "scan error")
		ans = nil
	}
	return
}

func scanTenantMembership(s util.RowScanner) (ans *models.TenantMembership, err error) {
	ans = new(models.TenantMembership)
	var prefsJson string
	err = s.Scan(&ans.Id, &ans.Invite, &ans.User, &ans.DisplayName, &ans.Avatar, &ans.Email, &ans.Phone, &prefsJson, &ans.Tenant, &ans.InvitedAt, &ans.InviteStatus, &ans.InviteExpiresAt, &ans.CreatedAt, &ans.UpdatedAt, &ans.Role)
	if err != nil {
		err = errs.Wrap(err, "scan error")
		ans = nil
	}

	if len(prefsJson) > 2 {
		err = json.Unmarshal([]byte(prefsJson), ans.Prefs)
	}
	return
}

func findTenantMemberships(ctx context.Context, id uint64) (ans []*models.TenantMembership, err error) {
	query := `
		SELECT
			id,
			invite,
			"user",
			display_name,
			avatar,
			email,
			phone,
			prefs,
			tenant,
			invited_at,
			invite_status,
			invite_expires_at,
			created_at,
			updated_at,
			"role"
		FROM
			vw_AllTenantMembers
		WHERE
			tenant=$1;
	`
	rows, err := tenantDb.Query(ctx, query, id)
	if err != nil {
		ans = nil
		return
	}
	defer rows.Close()

	for rows.Next() {
		var m *models.TenantMembership
		m, err = scanTenantMembership(rows)
		if err != nil {
			ans = nil
			return
		}
		ans = append(ans, m)
	}
	return
}

func createTenantInvite(ctx context.Context, tx *sqldb.Tx, role dto.PermissionName, email string, phone, displayName, avatar *string, window time.Duration, tenant uint64, user *uint64) (id uint64, err error) {
	query := `
		INSERT INTO member_invites("user",tenant,email,phone,role,display_name,avatar,"window")
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		RETURNING id;
	`
	err = tx.QueryRow(ctx, query, user, tenant, email, phone, string(role), displayName, avatar, fmt.Sprintf("%f hours", window.Hours())).Scan(&id)
	return
}

func createTenantMembership(ctx context.Context, tx *sqldb.Tx, invite uint64, role dto.PermissionName, email, displayName string, phone, avatar *string, prefs *map[string]string) (err error) {
	query := `
		INSERT INTO
			tenant_memberships(invite,avatar,role,email,phone,display_name,prefs)
		VALUES ($1,$2,$3,$4,$5,$6,$7);
	`

	var prefsJson string = "{}"

	if prefs != nil {
		var rawJson []byte
		rawJson, err = json.Marshal(prefs)
		if err != nil {
			return
		}
		prefsJson = string(rawJson)
	}

	_, err = tx.Exec(ctx, query, invite, avatar, role, email, phone, displayName, prefsJson)
	return
}

func findInviteById(ctx context.Context, id uint64) (ans *models.TenantMembershipInvitation, err error) {
	query := `
		SELECT
			id,
			"user",
			tenant,
			tenant_name,
			email,
			phone,
			"role",
			display_name,
			success_redirect,
			error_redirect,
			onboard_redirect,
			url,
			avatar,
			created_at,
			updated_at,
			invite_status,
			expires_at
		FROM
			vw_AllTenantInvitations
		WHERE
			id=$1;
	`

	ans, err = scanTenantInvitation(tenantDb.QueryRow(ctx, query, id))
	return
}
