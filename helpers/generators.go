package helpers

import (
	"math/rand"
	"time"

	"github.com/oklog/ulid"
)

func NewUlid() string {
	t := time.Unix(1000000, 0)
	entropy := ulid.Monotonic(rand.New(rand.NewSource(t.UnixNano())), 0)
	return ulid.MustNew(ulid.Timestamp(t), entropy).String()
}
