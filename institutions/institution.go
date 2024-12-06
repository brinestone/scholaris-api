// CRUD endpoints for institutions
package institutions

import (
	"context"
	"crypto/md5"
	_ "embed"
	"encoding/hex"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"time"

	"encore.dev/beta/auth"
	"encore.dev/beta/errs"
	"encore.dev/rlog"
	"encore.dev/storage/cache"
	"encore.dev/storage/sqldb"
	"github.com/brinestone/scholaris/core/permissions"
	"github.com/brinestone/scholaris/dto"
	"github.com/brinestone/scholaris/models"
	"github.com/brinestone/scholaris/settings"
	"github.com/brinestone/scholaris/tenants"
	"github.com/brinestone/scholaris/util"
	"github.com/lib/pq"
	"gopkg.in/yaml.v3"
)

// API Functions

// Gets more information for an institution
//
//encore:api public method=GET path=/institutions/:identifier
func GetInstitution(ctx context.Context, identifier string) (*dto.Institution, error) {
	return findInstitutionByGenericIdentifier(ctx, identifier)
}

// Looks up institutions
//
//encore:api public method=GET path=/institutions
func Lookup(ctx context.Context, req dto.LookupInstitutionsRequest) (*dto.LookupInstitutionsResponse, error) {
	uid, authed := auth.UserID()
	var memberedRefs []uint64

	if authed {
		memberedInstitutions, err := permissions.ListRelationsInternal(ctx, dto.ListObjectsRequest{
			Actor:    dto.IdentifierString(dto.PTUser, uid),
			Relation: models.PermMember,
			Type:     string(dto.PTInstitution),
		})
		if err != nil {
			return nil, err
		}
		memberedRefs = memberedInstitutions.Relations[dto.PTInstitution]
	}
	var ans []*models.Institution
	var err error
	if req.SubscribedOnly {
		ans, err = lookupSubscribedInstitutions(ctx, req.Page, req.Size, memberedRefs)
	} else {
		ans, err = lookupInstitutions(ctx, req.Page, req.Size, memberedRefs)
	}
	if err != nil {
		rlog.Error(err.Error())
		return nil, &util.ErrUnknown
	}

	return &dto.LookupInstitutionsResponse{
			Institutions: toInstitutionDtos(ans...),
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

	exists, err := institutionExists(ctx, req.Slug, req.TenantId)
	if err != nil {
		rlog.Error(util.MsgDbAccessError, "msg", err.Error())
		return nil, &util.ErrUnknown
	}

	if exists {
		return nil, &errs.Error{
			Code:    errs.AlreadyExists,
			Message: "An institution already exists with this name",
		}
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
				Actor:    dto.IdentifierString(dto.PTTenant, req.TenantId),
				Relation: models.PermParent,
				Target:   dto.IdentifierString(dto.PTInstitution, i.Id),
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

	ans := &dto.InstitutionLookup{
		Name:      i.Name,
		Visible:   i.Visible,
		Slug:      i.Slug,
		Id:        i.Id,
		TenantId:  i.TenantId,
		CreatedAt: i.CreatedAt,
		UpdatedAt: i.UpdatedAt,
		IsMember:  true,
	}

	if i.Description.Valid {
		ans.Description = &i.Description.String
	}
	if i.Logo.Valid {
		ans.Logo = &i.Logo.String
	}

	return ans, nil
}

// Private section

const institutionFields = "id,name,description,logo,visible,slug,tenant,created_at,updated_at,verified"

func institutionExists(ctx context.Context, slug string, tenant uint64) (ans bool, err error) {
	// Check whether the institution already exists under the same tenant.
	existsQuery := `
		SELECT 
			COUNT(id) as cnt 
		FROM 
			institutions 
		WHERE 
			slug=$1 AND tenant=$2
		;
	`

	var cnt uint
	if err = db.QueryRow(ctx, existsQuery, slug, tenant).Scan(&cnt); err != nil {
		return
	}

	ans = cnt > 0
	return
}

func createInstitution(ctx context.Context, tx *sqldb.Tx, req dto.NewInstitutionRequest) (uint64, error) {
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

func findInstitutionBySlugFromCache(ctx context.Context, slug string) (*dto.Institution, error) {
	return findInstitutionByKeyFromCache(ctx, "slug", slug)
}

func findInstitutionByIdFromCache(ctx context.Context, id uint64) (*dto.Institution, error) {
	return findInstitutionByKeyFromCache(ctx, "id", id)
}

func findInstitutionByKeyFromCache(ctx context.Context, key string, value any) (*dto.Institution, error) {
	sum := md5.Sum([]byte(fmt.Sprintf("%s=%v", key, value)))
	ans, err := institutionCache.Get(ctx, hex.EncodeToString(sum[:]))
	if err != nil {
		return nil, err
	}
	return &ans, nil
}

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
	if err := trx.QueryRow(ctx, query, value).Scan(&i.Id, &i.Name, &i.Description, &i.Logo, &i.Visible, &i.Slug, &i.TenantId, &i.CreatedAt, &i.UpdatedAt, &i.Verified); err != nil {
		return nil, err
	}

	return i, nil
}

func findInstitutionbySlugFromDb(ctx context.Context, slug string) (*models.Institution, error) {
	return findInstitutionByKeyFromDb(ctx, "slug", slug)
}

func findInstitutionByIdFromDbTrx(ctx context.Context, tx *sqldb.Tx, id uint64) (*models.Institution, error) {
	return findInstitutionByKeyFromDbTrx(ctx, tx, "id", id)
}

func findInstitutionByIdFromDb(ctx context.Context, id uint64) (*models.Institution, error) {
	i, err := findInstitutionByKeyFromDb(ctx, "id", id)
	if err != nil {
		return nil, err
	}

	return i, nil
}

// Finds an item from the cache using a key and the value of that key as the cache key
func findInstitutionByKeyFromDb(ctx context.Context, key string, value any) (*models.Institution, error) {
	var i = new(models.Institution)
	query := fmt.Sprintf("SELECT %s FROM institutions WHERE %s = $1;", institutionFields, key)
	row := db.QueryRow(ctx, query, value)

	if err := row.Scan(&i.Id, &i.Name, &i.Description, &i.Logo, &i.Visible, &i.Slug, &i.TenantId, &i.CreatedAt, &i.UpdatedAt, &i.Verified); err != nil {
		return nil, err
	}

	sum := md5.Sum([]byte(fmt.Sprintf("%s=%v", key, value)))
	_ = institutionCache.Set(ctx, hex.EncodeToString(sum[:]), toInstitutionDto(i)[0])
	return i, nil
}

func toInstitutionDtos(in ...*models.Institution) (res []dto.InstitutionLookup) {
	res = make([]dto.InstitutionLookup, len(in))
	for i, v := range in {
		u := dto.InstitutionLookup{
			Name:      v.Name,
			Visible:   v.Visible,
			Slug:      v.Slug,
			Id:        v.Id,
			CreatedAt: v.CreatedAt,
			UpdatedAt: v.UpdatedAt,
			Verified:  v.Verified,
			IsMember:  false,
		}

		if v.Logo.Valid {
			u.Logo = &v.Logo.String
		}

		if v.Description.Valid {
			u.Description = &v.Description.String
		}

		res[i] = u
	}
	return
}

func toInstitutionDto(in ...*models.Institution) (res []dto.Institution) {
	res = make([]dto.Institution, len(in))
	for i, v := range in {
		u := dto.Institution{
			Name:      v.Name,
			Visible:   v.Visible,
			Slug:      v.Slug,
			Id:        v.Id,
			CreatedAt: v.CreatedAt,
			UpdatedAt: v.UpdatedAt,
			Verified:  v.Verified,
			IsMember:  false,
		}

		if v.CurrentYear.Valid {
			tmp := uint64(v.CurrentYear.Int64)
			u.CurrentYear = &tmp
		}

		if v.CurrentTerm.Valid {
			tmp := uint64(v.CurrentTerm.Int64)
			u.CurrentTerm = &tmp
		}

		if v.Logo.Valid {
			u.Logo = &v.Logo.String
		}

		if v.Description.Valid {
			u.Description = &v.Description.String
		}

		res[i] = u
	}

	return
}

func lookupSubscribedInstitutions(ctx context.Context, page, size uint, ids []uint64) (ans []*models.Institution, err error) {
	query := "SELECT id,name,description,logo,visible,slug,tenant,verified,created_at,updated_at,current_year,current_term FROM vw_AllInstitutions WHERE id=ANY($1) OFFSET $2 LIMIT $3;"

	rows, err := db.Query(ctx, query, pq.Array(ids), page*size, size)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		var i = new(models.Institution)
		if err = rows.Scan(&i.Id, &i.Name, &i.Description, &i.Logo, &i.Visible, &i.Slug, &i.TenantId, &i.Verified, &i.CreatedAt, &i.UpdatedAt, &i.CurrentYear, &i.CurrentTerm); err != nil {
			return
		}
		ans = append(ans, i)
	}
	return
}

func lookupInstitutions(ctx context.Context, page, size uint, ids []uint64) ([]*models.Institution, error) {
	ans := make([]*models.Institution, 0)
	query := "SELECT id,name,description,logo,visible,slug,tenant,verified,created_at,updated_at,current_year,current_term FROM vw_AllInstitutions WHERE id=ANY($1) OR verified=true OFFSET $2 LIMIT $3;"

	rows, err := db.Query(ctx, query, pq.Array(ids), page*size, size)
	if err != nil {
		if !errors.Is(err, sqldb.ErrNoRows) {
			return ans, err
		}
	}
	defer rows.Close()

	for rows.Next() {
		var i = new(models.Institution)
		if err := rows.Scan(&i.Id, &i.Name, &i.Description, &i.Logo, &i.Visible, &i.Slug, &i.TenantId, &i.Verified, &i.CreatedAt, &i.UpdatedAt, &i.CurrentYear, &i.CurrentTerm); err != nil {
			return ans, err
		}
		ans = append(ans, i)
	}

	return ans, nil
}

func findInstitutionByGenericIdentifier(ctx context.Context, identifier string) (*dto.Institution, error) {
	if match, _ := regexp.MatchString(`^\d+$`, identifier); match {
		id, _ := strconv.ParseUint(identifier, 10, 64)
		ans, err := findInstitutionByIdFromCache(ctx, id)
		if errors.Is(err, cache.Miss) {
			institution, err := findInstitutionByIdFromDb(ctx, id)
			if errors.Is(err, sqldb.ErrNoRows) {
				return nil, &util.ErrNotFound
			} else if err != nil {
				rlog.Error(util.MsgDbAccessError, "msg", err.Error())
				return nil, &util.ErrUnknown
			}

			return &toInstitutionDto(institution)[0], nil
		} else if err != nil {
			rlog.Error(util.MsgCacheAccessError, "msg", err.Error())
			return nil, &util.ErrUnknown
		}
		return ans, nil
	} else {
		ans, err := findInstitutionBySlugFromCache(ctx, identifier)
		if errors.Is(err, cache.Miss) {
			institution, err := findInstitutionbySlugFromDb(ctx, identifier)
			if errors.Is(err, sqldb.ErrNoRows) {
				return nil, &util.ErrNotFound
			} else if err != nil {
				rlog.Error(util.MsgDbAccessError, "msg", err.Error())
				return nil, &util.ErrUnknown
			}

			return &toInstitutionDto(institution)[0], nil
		} else if err != nil {
			rlog.Error(util.MsgCacheAccessError, "msg", err.Error())
			return nil, &util.ErrUnknown
		}
		return ans, nil
	}
}

//go:embed default-settings.yml
var defSettings []byte

type DefaultSetting struct {
	Label       string   `yaml:"label"`
	Value       []string `yaml:"value"`
	Description string   `yaml:"description"`
	MultiValues bool     `yaml:"multiValues"`
	Parent      string   `yaml:"parent"`
}

type DefaultSettings struct {
	Settings map[string]DefaultSetting `yaml:"settings"`
}

func defineInstitutionDefaultSettings(ctx context.Context, id uint64) error {
	var sMap = DefaultSettings{
		Settings: make(map[string]DefaultSetting),
	}

	if err := yaml.Unmarshal(defSettings, sMap); err != nil {
		return err
	}

	req := dto.UpdateSettingsInternalRequest{
		OwnerType: string(dto.PTInstitution),
		Owner:     id,
	}

	for k, v := range sMap.Settings {
		req.Updates = append(req.Updates, dto.SettingUpdate{
			Key:         k,
			Label:       v.Label,
			Description: &v.Description,
			MultiValues: v.MultiValues,
		})
	}
	if err := settings.UpdateSettingsInternal(ctx, req); err != nil {
		return err
	}

	req2 := dto.SetSettingValueRequest{
		Owner:     id,
		OwnerType: string(dto.PTInstitution),
	}

	for k, v := range sMap.Settings {
		s := dto.SettingValueUpdate{
			Key: k,
		}
		for _, v := range v.Value {
			s.Value = append(s.Value, dto.SetValue{
				Value: &v,
			})
		}
		req2.Updates = append(req2.Updates, s)
	}

	return settings.SetSettingValuesInternal(ctx, req2)
}

func getInstitutionStats(ctx context.Context) (*models.InstitutionStatistics, error) {
	var ans models.InstitutionStatistics
	query := `
		SELECT
			total,verified,unverified
		FROM
			vw_InstitutionStatistics
		;
	`
	if err := db.QueryRow(ctx, query).Scan(&ans.TotalInstitutions, &ans.TotalVerified, &ans.TotalUnverified); err != nil {
		return nil, err
	}
	return &ans, nil
}
