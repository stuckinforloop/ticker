package ulid

import (
	"math/rand"

	"github.com/oklog/ulid/v2"
)

type Source struct {
	*rand.Rand
}

// New generates a ulid.
// If the source is not nil and ms is provided
// then that is used for generating a ulid which
// may not be random (useful for deterministic tests)
func (s *Source) New(ms uint64) string {
	if s.Rand != nil {
		return ulid.MustNew(ms, s.Rand).String()
	}

	return ulid.Make().String()
}
