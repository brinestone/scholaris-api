package institutions_test

import (
	"context"
	"fmt"
	"testing"

	"encore.dev/et"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/brinestone/scholaris/core/permissions"
	"github.com/brinestone/scholaris/dto"
	"github.com/brinestone/scholaris/institutions"
	"github.com/stretchr/testify/assert"
)

func TestLookup(t *testing.T) {
	cnt := gofakeit.UintRange(1, 10)
	ids := make([]uint64, cnt)
	for i := uint(0); i < cnt; i++ {
		created, err := makeInstitution()
		if err != nil {
			t.Error(err)
			return
		}
		ids[i] = created.Id
	}

	et.MockEndpoint(permissions.ListRelations, func(ctx context.Context, p dto.ListRelationsRequest) (ans *dto.ListRelationsResponse, err error) {
		ans = &dto.ListRelationsResponse{
			Relations: map[dto.PermissionType][]uint64{
				dto.PTInstitution: ids,
			},
		}
		return
	})
	defer mockEndpoints()

	t.Run("all", func(t *testing.T) {
		res, err := institutions.Lookup(mainContext, dto.LookupInstitutionsRequest{
			Page: 0,
			Size: 10,
		})

		if err != nil {
			t.Error(err)
			return
		}

		assert.NotNil(t, res)
		assert.NotEmpty(t, res.Institutions)
	})

	t.Run("subscribedOnly", func(t *testing.T) {
		res, err := institutions.Lookup(mainContext, dto.LookupInstitutionsRequest{
			Page:           0,
			Size:           10,
			SubscribedOnly: true,
		})

		if err != nil {
			t.Error(err)
			return
		}

		assert.NotNil(t, res)
		assert.GreaterOrEqual(t, uint(len(res.Institutions)), cnt)
	})
}

func TestNewIntitution(t *testing.T) {
	lookup, err := makeInstitution()

	assert.Nil(t, err)
	assert.NotNil(t, lookup)
}

func TestGetInstitution(t *testing.T) {
	i, err := makeInstitution()
	if err != nil {
		t.Error(err)
		return
	}

	assert.NotNil(t, i)

	t.Run("UsingSlugIdentifier", func(t *testing.T) {
		res, err := institutions.GetInstitution(mainContext, i.Slug)
		if err != nil {
			t.Error(err)
			return
		}

		assert.NotNil(t, res)
		assert.Equal(t, i.Id, res.Id)
		assert.Equal(t, i.Slug, res.Slug)
	})

	t.Run("UsingIdIdentifier", func(t *testing.T) {
		res, err := institutions.GetInstitution(mainContext, fmt.Sprintf("%d", i.Id))
		if err != nil {
			t.Error(err)
			return
		}

		assert.NotNil(t, res)
		assert.Equal(t, i.Id, res.Id)
		assert.Equal(t, i.Slug, res.Slug)
	})
}
