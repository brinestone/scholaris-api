package institutions

import (
	"context"

	"github.com/brinestone/scholaris/dto"
)

// Creates a new Enrollment
//
//encore:api method=POST auth path=/institutions/:institution/enroll
func NewEnrollment(ctx context.Context, institution uint64) (*dto.GetFormQuestionsResponse, error) {
	return nil, nil
}
