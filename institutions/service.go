// CRUD endpoints for institutions
package institutions

import (
	"context"
	"errors"
	"fmt"

	"encore.dev/beta/errs"
	"encore.dev/rlog"
	"encore.dev/storage/cache"
	"encore.dev/storage/sqldb"
	"github.com/brinestone/scholaris/core/permissions"
	"github.com/brinestone/scholaris/dto"
	"github.com/brinestone/scholaris/models"
	"github.com/brinestone/scholaris/tenants"
	"github.com/brinestone/scholaris/util"
)

// Finds an institution using its provided ID
//
//encore:api auth method=GET path=/institutions/:id tag:perm_can_read_institution
func FindInstitution(ctx context.Context, id uint64) (*models.Institution, error) {
	var i *models.Institution
	var err error

	i, err = findInstitutionByIdFromCache(ctx, id)
	if err != nil {
		if errors.Is(err, cache.Miss) {
			i, err = findInstitutionByIdFromDb(ctx, id)
		} else {
			return nil, err
		}
	}

	if err != nil {
		if errors.Is(err, sqldb.ErrNoRows) {
			return nil, &util.ErrNotFound
		}
		return nil, err
	}

	return i, err
}

// Creates a new institution
//
//encore:api auth method=POST path=/institutions tag:perm_can_create_institution
func New(ctx context.Context, req dto.NewInstitutionRequest) (*models.Institution, error) {

	_, err := tenants.FindTenant(ctx, req.TenantId)
	if err != nil {
		return nil, &util.ErrNotFound
	}

	trx, err := db.Begin(ctx)
	if err != nil {
		rlog.Error(err.Error())
		return nil, &util.ErrUnknown
	}

	if err = createInstitution(ctx, req); err != nil {
		rlog.Error(err.Error())
		_ = trx.Rollback()
		return nil, &util.ErrUnknown
	}

	i, err := findInstitutionByKeyFromDb(ctx, "slug", req.Slug)
	if err != nil {
		rlog.Error(err.Error())
		return nil, nil
	}

	if err := permissions.SetPermissions(ctx, dto.UpdatePermissionsRequest{
		Updates: []dto.PermissionUpdate{
			{
				Subject:  fmt.Sprintf("tenant:%d", req.TenantId),
				Relation: "parent",
				Target:   fmt.Sprintf("institution:%d", i.Id),
			},
		},
	}); err != nil {
		rlog.Error(err.Error())
		_ = trx.Rollback()
		return nil, &util.ErrUnknown
	}

	defer func() {
		_ = trx.Commit()
		NewInstitutions.Publish(ctx, &InstitutionCreated{})
	}()

	return i, nil
}

const institutionFields = "id,name,description,logo,visible,slug,tenant,created_at,updated_at"

func createInstitution(ctx context.Context, req dto.NewInstitutionRequest) error {

	existsQuery := "SELECT COUNT(id) as cnt FROM institutions WHERE slug=$1 AND tenant=$2;"
	cnt := 0
	res := db.QueryRow(ctx, existsQuery, req.Slug, req.TenantId)
	if err := res.Scan(&cnt); err != nil {
		if !errors.Is(err, sqldb.ErrNoRows) {
			return err
		} else if cnt > 0 {
			return &errs.Error{
				Code:    errs.AlreadyExists,
				Message: "There is already an institution with the same slug registered under this tenant",
			}
		}
	}

	insertQuery := "INSERT INTO institutions(name,description,logo,slug,tenant) VALUES ($1,$2,$3,$4,$5);"
	if _, err := db.Exec(ctx, insertQuery, req.Name, req.Description, req.Logo, req.Slug, req.TenantId); err != nil {
		return err
	}

	return nil
}

func findInstitutionByIdFromCache(ctx context.Context, id uint64) (*models.Institution, error) {
	var err error

	i, err := institutionCache.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	return &i, nil
}

func findInstitutionByIdFromDb(ctx context.Context, id uint64) (*models.Institution, error) {
	i, err := findInstitutionByKeyFromDb(ctx, "id", id)
	if err != nil {
		return nil, err
	}

	if err = institutionCache.Set(ctx, i.Id, *i); err != nil {
		return nil, err
	}

	return i, nil
}

func findInstitutionByKeyFromDb(ctx context.Context, key string, value any) (*models.Institution, error) {
	var i = new(models.Institution)
	query := fmt.Sprintf("SELECT %s FROM institutions WHERE %s = $1;", institutionFields, key)
	row := db.QueryRow(ctx, query, value)

	if err := row.Scan(&i.Id, &i.Name, &i.Description, &i.Logo, &i.Visible, &i.Slug, &i.TenantId, &i.CreatedAt, &i.UpdatedAt); err != nil {
		return nil, err
	}

	_ = institutionCache.Set(ctx, i.Id, *i)
	return i, nil
}
