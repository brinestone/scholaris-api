package institutions_test

import (
	"fmt"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/brinestone/scholaris/dto"
	"github.com/brinestone/scholaris/institutions"
	"github.com/stretchr/testify/assert"
)

func TestLookup(t *testing.T) {
	cnt := gofakeit.UintRange(1, 10)
	for i := uint(0); i < cnt; i++ {
		_, err := makeInstitution()
		if err != nil {
			t.Error(err)
			return
		}
	}

	res, err := institutions.Lookup(mainContext, &dto.PageBasedPaginationParams{
		Page: 0,
		Size: 10,
	})

	if err != nil {
		t.Error(err)
		return
	}

	assert.NotNil(t, res)
	assert.GreaterOrEqual(t, res.Meta.Total, cnt)
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
