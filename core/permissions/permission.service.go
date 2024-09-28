package permissions

import (
	"context"
	"fmt"

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
}

func initService() (*Service, error) {
	rlog.Debug("starting permissions service...")
	var err error
	fgaClient, err := client.NewSdkClient(&client.ClientConfiguration{
		ApiUrl:  secrets.FgaUrl,
		StoreId: secrets.FgaStoreId,
	})
	if err != nil {
		return nil, err
	}
	rlog.Debug(fmt.Sprintf("Connected to OpenFGA Server on %s", secrets.FgaUrl))

	return &Service{
		fgaClient: fgaClient,
	}, nil
}

//encore:api private method=POST path=/permissions
func (s *Service) SetPermissions(ctx context.Context, req dto.UpdatePermissionsRequest) error {
	s.fgaClient.Write(ctx).Body(client.ClientWriteRequest{
		Writes: toOpenFgaUpdates(req.Updates),
	}).
		Execute()
	return nil
}

func toOpenFgaUpdates(updates []dto.PermissionUpdate) []openfga.TupleKey {
	ans := make([]client.ClientTupleKey, 0)

	for _, u := range updates {
		ans = append(ans, client.ClientTupleKey{
			User:     u.User,
			Relation: u.Relation,
			Object:   u.Object,
		})
	}

	return ans
}
