package permissions

import (
	"context"
	"strconv"
	"strings"
	"time"

	"encore.dev"
	"encore.dev/beta/auth"
	"encore.dev/rlog"
	"github.com/brinestone/scholaris/dto"
	"github.com/brinestone/scholaris/helpers"
	"github.com/brinestone/scholaris/util"
	openfga "github.com/openfga/go-sdk"
	"github.com/openfga/go-sdk/client"
	"github.com/openfga/go-sdk/credentials"
)

//encore:service
type Service struct {
	fgaClient *client.OpenFgaClient
}

var secrets struct {
	FgaUrl          string `encore:"sensitive"`
	FgaStoreId      string `encore:"sensitive"`
	FgaClientSecret string `encore:"sensitive"`
	FgaClientId     string `encore:"sensitive"`
	FgaAudience     string `encore:"sensitive"`
	FgaIssuer       string `encore:"sensitive"`
}

func initService() (*Service, error) {
	var err error
	config := &client.ClientConfiguration{
		ApiUrl:  secrets.FgaUrl,
		StoreId: secrets.FgaStoreId,
	}

	switch encore.Meta().Environment.Cloud {
	case encore.CloudAWS, encore.EncoreCloud, encore.CloudAzure, encore.CloudGCP:
		config.Credentials = &credentials.Credentials{
			Method: credentials.CredentialsMethodClientCredentials,
			Config: &credentials.Config{
				ClientCredentialsClientId:       secrets.FgaClientId,
				ClientCredentialsClientSecret:   secrets.FgaClientSecret,
				ClientCredentialsApiAudience:    secrets.FgaAudience,
				ClientCredentialsApiTokenIssuer: secrets.FgaIssuer,
			},
		}
	}
	fgaClient, err := client.NewSdkClient(config)
	if err != nil {
		return nil, err
	}

	return &Service{
		fgaClient: fgaClient,
	}, nil
}

// List Objects with valid relations
//
// encore:api auth method=POST path=/permissions/related
func (s *Service) ListRelations(ctx context.Context, req dto.ListRelationsRequest) (ans *dto.ListRelationsResponse, err error) {
	ans = new(dto.ListRelationsResponse)
	uid, _ := auth.UserID()
	cacheKey := relationsCacheKey(uid, req.Target, req.Permissions...)
	ans.Relations, err = relationsCache.Items(ctx, cacheKey)
	if len(ans.Relations) == 0 {
		body := client.ClientListRelationsRequest{
			User:      dto.IdentifierString(dto.PTUser, uid),
			Relations: req.Permissions,
			Object:    req.Target,
		}

		var res *client.ClientListRelationsResponse
		res, err = s.fgaClient.ListRelations(ctx).
			Body(body).
			Execute()
		if err != nil {
			rlog.Error(util.MsgCallError, "err", err)
			err = &util.ErrUnknown
			return
		}

		ans.Relations = res.Relations
		defer func() {
			if len(res.Relations) == 0 {
				return
			}
			relationsCache.RemoveAll(ctx, cacheKey, "")
			for i, v := range res.Relations {
				if err := relationsCache.Set(ctx, cacheKey, int64(i), v); err != nil {
					rlog.Error(util.MsgCacheAccessError, "err", err)
					break
				}
			}
		}()
	} else if err != nil {
		rlog.Error(util.MsgCacheAccessError, "err", err)
		err = &util.ErrUnknown
		ans = nil
		return
	}
	return
}

// List Objects with valid relations (Internal API)
//
//encore:api private method=POST path=/permissions/related/internal
func (s *Service) ListObjectsInternal(ctx context.Context, req dto.ListObjectsRequest) (*dto.ListObjectsResponse, error) {
	reqBody := client.ClientListObjectsRequest{
		User:     req.Actor,
		Relation: string(req.Relation),
		Type:     string(req.Type),
	}

	if len(req.Context) > 0 {
		reqBody.Context = contextEntriesToMap(req.Context...)
	}

	data, err := s.fgaClient.ListObjects(ctx).
		Body(reqBody).
		Execute()
	if err != nil {
		return nil, err
	}

	resultMap := make(map[dto.PermissionType][]uint64)

	for _, rel := range data.GetObjects() {
		arr := strings.Split(rel, ":")
		p, _ := dto.ParsePermissionType(arr[0])
		id, _ := strconv.ParseUint(arr[1], 10, 64)
		resultMap[p] = append(resultMap[p], id)
	}

	return &dto.ListObjectsResponse{
		Relations: resultMap,
	}, nil
}

// Checks whether a permission is valid or not
//
//encore:api auth method=POST path=/permissions/check
func (s *Service) CheckPermission(ctx context.Context, req dto.BatchRelationCheckRequest) (ans *dto.BatchRelationCheckResponse, err error) {
	ans = new(dto.BatchRelationCheckResponse)
	ans.Results = make(map[string]bool)

	uid, _ := auth.UserID()
	actor := dto.IdentifierString(dto.PTUser, uid)
	res, err := s.fgaClient.BatchCheck(ctx).
		Body(helpers.SliceMap(req.Checks, func(c dto.RelationCheck) client.ClientCheckRequest {
			return client.ClientCheckRequest{
				User:     actor,
				Relation: c.Relation,
				Object:   c.Target,
			}
		})).
		Execute()
	if err != nil {
		rlog.Error(util.MsgCallError, "err", err)
		err = &util.ErrUnknown
	}

	if res == nil {
		ans = nil
		return
	}

	ans.Results = helpers.SliceReduce(*res, func(c client.ClientBatchCheckSingleResponse, r map[string]bool) map[string]bool {
		r[c.Request.Relation] = *c.Allowed
		return r
	}, helpers.WithSeed(ans.Results))
	return
}

// Checks whether a permission is valid or not (Internal API)
//
//encore:api private method=POST path=/permissions/check/internal
func (s *Service) CheckPermissionInternal(ctx context.Context, req dto.InternalRelationCheckRequest) (ans *dto.RelationCheckResponse, err error) {
	ans, err = s.doPermissionCheck(ctx, req.Actor, req.Relation, req.Target, req.Condition)
	return
}

// Deletes permission Tuples
//
//encore:api private method=POST path=/permissions/down
func (s *Service) DeletePermissions(ctx context.Context, req dto.UpdatePermissionsRequest) error {
	if _, err := s.fgaClient.Write(ctx).Body(client.ClientWriteRequest{
		Deletes: toOpenFgaDeletes(req.Updates),
	}).Execute(); err != nil {
		return err
	}
	return nil
}

// Updates permission Tuples
//
//encore:api private method=POST path=/permissions/up
func (s *Service) SetPermissions(ctx context.Context, req dto.UpdatePermissionsRequest) error {
	if _, err := s.fgaClient.Write(ctx).Body(client.ClientWriteRequest{
		Writes: toOpenFgaWrites(req.Updates),
	}).Execute(); err != nil {
		return err
	}
	return nil
}

func toOpenFgaDeletes(updates []dto.PermissionUpdate) []openfga.TupleKeyWithoutCondition {
	ans := make([]client.ClientTupleKeyWithoutCondition, 0)

	for _, u := range updates {
		ans = append(ans, client.ClientTupleKeyWithoutCondition{
			User:     u.Actor,
			Relation: string(u.Relation),
			Object:   u.Target,
		})
	}
	return ans
}

func toOpenFgaWrites(updates []dto.PermissionUpdate) []openfga.TupleKey {
	ans := make([]client.ClientTupleKey, 0)

	for _, u := range updates {
		var condition *openfga.RelationshipCondition
		if u.Condition != nil {
			c := make(map[string]any)

			for _, v := range u.Condition.Context {
				c[v.Name] = v.Value
			}

			condition = &openfga.RelationshipCondition{
				Name:    u.Condition.Name,
				Context: &c,
			}
		}

		ans = append(ans, client.ClientTupleKey{
			User:      u.Actor,
			Relation:  string(u.Relation),
			Object:    u.Target,
			Condition: condition,
		})
	}

	return ans
}

func contextEntriesToMap(entries ...dto.ContextEntry) (ans *map[string]any) {
	ans = &map[string]any{}
	c := *ans
	for _, v := range entries {
		switch v.Type {
		case dto.CETBool:
			c[v.Name], _ = strconv.ParseBool(v.Value)
		case dto.CETTimestamp:
			c[v.Name], _ = time.Parse(time.RFC3339, v.Value)
		case dto.CETDuration:
			t, err := time.ParseDuration(v.Value)
			if err != nil {
				continue
			}
			c[v.Name] = t.String()
		default:
			c[v.Name] = v.Value
		}
	}
	return
}

func (s *Service) doPermissionCheck(ctx context.Context, actor string, relation dto.PermissionName, target string, condition *dto.RelationCondition) (ans *dto.RelationCheckResponse, err error) {
	request := client.ClientCheckRequest{
		User:     actor,
		Relation: string(relation),
		Object:   target,
	}

	if condition != nil {
		request.Context = contextEntriesToMap(condition.Context...)
	}

	res, err := s.fgaClient.
		Check(ctx).
		Body(request).
		Execute()

	if err != nil {
		return
	}

	ans = &dto.RelationCheckResponse{Allowed: *res.Allowed}
	return
}
