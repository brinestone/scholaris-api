// CRUD endpoints for institutions
package institutions

import (
	"context"
	"errors"
	"fmt"
	"time"

	"encore.dev/beta/auth"
	"encore.dev/beta/errs"
	"encore.dev/rlog"
	"encore.dev/storage/sqldb"
	"github.com/brinestone/scholaris/core/permissions"
	"github.com/brinestone/scholaris/dto"
	"github.com/brinestone/scholaris/models"
	"github.com/brinestone/scholaris/tenants"
	"github.com/brinestone/scholaris/util"
)

// API Functions

// Looks up institutions
//
//encore:api public method=GET path=/institutions
func Lookup(ctx context.Context, req *dto.PageBasedPaginationParams) (*dto.PaginatedResponse[dto.InstitutionLookup], error) {
	var uid *auth.UID
	if temp, ok := auth.UserID(); ok {
		uid = &temp
	}

	ans, cnt, err := lookupInstitutions(ctx, req.Page, req.Size, uid)
	if err != nil {
		rlog.Error(err.Error())
		return nil, &util.ErrUnknown
	}

	return &dto.PaginatedResponse[dto.InstitutionLookup]{
			Data: ans,
			Meta: dto.PaginatedResponseMeta{
				Total: cnt,
			},
		},
		nil
}

// Creates a new institution
//
//encore:api auth method=POST path=/institutions tag:perm_can_create tag:needs_captcha_ver
func NewInstitution(ctx context.Context, req dto.NewInstitutionRequest) (*dto.InstitutionLookup, error) {

	_, err := tenants.FindTenant(ctx, req.TenantId)
	if err != nil {
		rlog.Error(err.Error())
		return nil, &util.ErrUnknown
	}

	trx, err := db.Begin(ctx)
	if err != nil {
		rlog.Error(err.Error())
		return nil, &util.ErrUnknown
	}

	id, err := createInstitution(ctx, trx, req)
	if err != nil {
		rlog.Error(err.Error())
		_ = trx.Rollback()
		return nil, &util.ErrUnknown
	}

	i, err := findInstitutionByIdFromDbTrx(ctx, trx, id)
	if err != nil {
		rlog.Error(err.Error())
		_ = trx.Rollback()
		return nil, &util.ErrUnknown
	}

	if err := permissions.SetPermissions(ctx, dto.UpdatePermissionsRequest{
		Updates: []dto.PermissionUpdate{
			{
				Subject:  fmt.Sprintf("%s:%d", dto.PTTenant, req.TenantId),
				Relation: models.PermParent,
				Target:   fmt.Sprintf("%s:%d", dto.PTInstitution, i.Id),
			},
		},
	}); err != nil {
		rlog.Error(err.Error())
		_ = trx.Rollback()
		return nil, &util.ErrUnknown
	}

	userId, _ := auth.UserID()

	if err = trx.Commit(); err != nil {
		rlog.Error(err.Error())
		return nil, &util.ErrUnknown
	}

	defer func() {
		NewInstitutions.Publish(ctx, &InstitutionCreated{
			Id:        i.Id,
			CreatedBy: userId,
			Timestamp: time.Now(),
		})
	}()

	return &dto.InstitutionLookup{
		Name:        i.Name,
		Description: i.Description,
		Logo:        i.Logo,
		Visible:     i.Visible,
		Slug:        i.Slug,
		Id:          i.Id,
		TenantId:    i.TenantId,
		CreatedAt:   i.CreatedAt,
		UpdatedAt:   i.UpdatedAt,
		IsMember:    true,
	}, nil
}

// Private section

const institutionFields = "id,name,description,logo,visible,slug,tenant,created_at,updated_at"

func createInstitution(ctx context.Context, tx *sqldb.Tx, req dto.NewInstitutionRequest) (uint64, error) {
	// Check whether the institution already exists under the same tenant.
	existsQuery := `
		SELECT 
			COUNT(id) as cnt 
		FROM 
			institutions 
		WHERE 
			slug=$1 AND tenant=$2;
	`

	var cnt uint
	if err := tx.QueryRow(ctx, existsQuery, req.Slug, req.TenantId).Scan(&cnt); err != nil {
		return 0, err
	}

	if cnt > 0 {
		return 0, &errs.Error{
			Code:    errs.AlreadyExists,
			Message: "An institution already exists with the same identifier",
		}
	}

	insertQuery := `
		INSERT INTO 
			institutions(name,description,logo,slug,tenant) 
		VALUES ($1,$2,$3,$4,$5) RETURNING id;
	`

	var newId uint64
	var description *string = &req.Description
	var logo *string = &req.Logo
	if len(*description) == 0 {
		description = nil
	}

	if len(*logo) == 0 {
		logo = nil
	}

	if err := tx.QueryRow(ctx, insertQuery, req.Name, description, logo, req.Slug, req.TenantId).Scan(&newId); err != nil {
		return 0, err
	}

	return newId, nil
}

// func findInstitutionByIdFromCache(ctx context.Context, id uint64) (*models.Institution, error) {
// 	return findInstitutionByKeyFromCache(ctx, "id", id)
// }

// func findInstitutionByKeyFromCache(ctx context.Context, key string, value any) (*models.Institution, error) {
// 	ans, err := institutionCache.Get(ctx, fmt.Sprintf("%s=%v", key, value))
// 	if err != nil {
// 		return nil, err
// 	}
// 	return &ans, nil
// }

func findInstitutionByKeyFromDbTrx(ctx context.Context, trx *sqldb.Tx, key string, value any) (*models.Institution, error) {
	var i = new(models.Institution)
	query := fmt.Sprintf(`
		SELECT
			%s
		FROM
			institutions
		WHERE
			%s = $1;
	`, institutionFields, key)
	if err := trx.QueryRow(ctx, query, value).Scan(&i.Id, &i.Name, &i.Description, &i.Logo, &i.Visible, &i.Slug, &i.TenantId, &i.CreatedAt, &i.UpdatedAt); err != nil {
		return nil, err
	}

	return i, nil
}

func findInstitutionByIdFromDbTrx(ctx context.Context, tx *sqldb.Tx, id uint64) (*models.Institution, error) {
	return findInstitutionByKeyFromDbTrx(ctx, tx, "id", id)
}

// func findInstitutionByIdFromDb(ctx context.Context, id uint64) (*models.Institution, error) {
// 	i, err := findInstitutionByKeyFromDb(ctx, "id", id)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return i, nil
// }

// Finds an item from the cache using a key and the value of that key as the cache key
// func findInstitutionByKeyFromDb(ctx context.Context, key string, value any) (*models.Institution, error) {
// 	var i = new(models.Institution)
// 	query := fmt.Sprintf("SELECT %s FROM institutions WHERE %s = $1;", institutionFields, key)
// 	row := db.QueryRow(ctx, query, value)

// 	if err := row.Scan(&i.Id, &i.Name, &i.Description, &i.Logo, &i.Visible, &i.Slug, &i.TenantId, &i.CreatedAt, &i.UpdatedAt); err != nil {
// 		return nil, err
// 	}

// 	_ = institutionCache.Set(ctx, fmt.Sprintf("%s=%v", key, value), *i)
// 	return i, nil
// }

func lookupInstitutions(ctx context.Context, page uint, size uint, uid *auth.UID) ([]*dto.InstitutionLookup, uint, error) {
	var cnt uint
	ans := make([]*dto.InstitutionLookup, 0)

	if err := db.QueryRow(ctx, `
		SELECT
			COUNT(id)
		FROM
			institutions;
	`).Scan(&cnt); err != nil {
		return ans, 0, err
	}

	var query = fmt.Sprintf(
		`
		SELECT
			%s
		FROM
			institutions
		ORDER BY
			id ASC
		OFFSET $1
		LIMIT $2;
	`, institutionFields)

	rows, err := db.Query(ctx, query, page*size, size)
	if err != nil {
		if !errors.Is(err, sqldb.ErrNoRows) {
			return ans, 0, err
		}
	}
	defer rows.Close()

	for rows.Next() {
		var i = new(dto.InstitutionLookup)
		if err := rows.Scan(&i.Id, &i.Name, &i.Description, &i.Logo, &i.Visible, &i.Slug, &i.TenantId, &i.CreatedAt, &i.UpdatedAt); err != nil {
			return ans, 0, err
		}
		ans = append(ans, i)
	}

	if uid != nil {
		memberedInstitutions, err := permissions.ListRelations(ctx, dto.ListRelationsRequest{
			Subject:  fmt.Sprintf("%s:%v", dto.PTUser, *uid),
			Relation: models.PermMember,
			Type:     string(dto.PTInstitution),
		})
		if err != nil {
			return ans, cnt, err
		}

		for _, i := range ans {
			for _, j := range memberedInstitutions.Relations[dto.PTInstitution] {
				i.IsMember = fmt.Sprintf("%d", i.Id) == j
			}
		}
	}

	return ans, cnt, nil
}
