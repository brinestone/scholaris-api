package permissions

import (
	"context"
	_ "embed"

	"encore.dev/rlog"
	"github.com/brinestone/scholaris/dto"
	openfga "github.com/openfga/go-sdk"
	"github.com/openfga/go-sdk/client"
)

//encore:service
type Service struct {
	fgaClient *client.OpenFgaClient
}

var secrets struct {
	FgaUrl     string `encore:"sensitive"`
	FgaStoreId string `encore:"sensitive"`
	FgaModelId string `encore:"sensitive"`
}

//go:embed system.json
var dsl string

func initService() (*Service, error) {
	var err error
	fgaClient, err := client.NewSdkClient(&client.ClientConfiguration{
		ApiUrl:               secrets.FgaUrl,
		StoreId:              secrets.FgaStoreId,
		AuthorizationModelId: secrets.FgaModelId,
	})
	if err != nil {
		return nil, err
	}

	return &Service{
		fgaClient: fgaClient,
	}, nil
}

// Checks whether a permission is valid or not.
//
//encore:api private method=GET path=/permissions/check
func (s *Service) CheckPermission(ctx context.Context, req dto.RelationCheckRequest) (*dto.RelationCheckResponse, error) {
	res, err := s.fgaClient.Check(ctx).Body(client.ClientCheckRequest{
		User:     req.Subject,
		Relation: req.Relation,
		Object:   req.Target,
	}).Execute()

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
			User:     u.Subject,
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
			User:      u.Subject,
			Relation:  u.Relation,
			Object:    u.Target,
			Condition: condition,
		})
	}

	return ans
}
