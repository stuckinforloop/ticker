package ulid

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
)

var testULID = "000000000006AFVGQT5ZYC0GEK"

func TestULID(t *testing.T) {
	rand := rand.New(rand.NewSource(0))
	ulidSource := Source{rand}
	ulid := ulidSource.New(0)
	assert.Equal(t, testULID, ulid)
}
