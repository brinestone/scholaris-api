package permissions

import (
	"context"
	"strconv"
	"strings"
	"time"

	"encore.dev"
	"encore.dev/rlog"
	"github.com/brinestone/scholaris/dto"
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
//encore:api private method=POST path=/permissions/related
func (s *Service) ListRelations(ctx context.Context, req dto.ListRelationsRequest) (*dto.ListRelationsResponse, error) {
	reqBody := client.ClientListObjectsRequest{
		User:     req.Actor,
		Relation: req.Relation,
		Type:     string(req.Type),
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

	return &dto.ListRelationsResponse{
		Relations: resultMap,
	}, nil
}

// Checks whether a permission is valid or not.
//
//encore:api private method=POST path=/permissions/check
func (s *Service) CheckPermission(ctx context.Context, req dto.RelationCheckRequest) (*dto.RelationCheckResponse, error) {
	request := client.ClientCheckRequest{
		User:     req.Actor,
		Relation: req.Relation,
		Object:   req.Target,
	}

	if req.Condition != nil {
		c := make(map[string]interface{})
		request.Context = &c

		for _, v := range req.Condition.Context {
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
	}

	res, err := s.fgaClient.
		Check(ctx).
		Body(request).
		Execute()

	if err != nil {
		rlog.Error(err.Error())
	}

	return &dto.RelationCheckResponse{Allowed: *res.Allowed}, nil
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
			Relation: u.Relation,
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
			Relation:  u.Relation,
			Object:    u.Target,
			Condition: condition,
		})
	}

	return ans
}
