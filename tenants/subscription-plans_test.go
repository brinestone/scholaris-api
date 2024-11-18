package tenants_test

import (
	"context"
	"testing"

	"github.com/brinestone/scholaris/tenants"
	"github.com/stretchr/testify/assert"
)

func TestFindSubscriptionPlans(t *testing.T) {
	res, err := tenants.FindSubscriptionPlans(context.TODO())
	if err != nil {
		t.Error(t, err)
		return
	}

	assert.NotNil(t, res)
	assert.NotEmpty(t, res.Plans)
}
