package dao

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/stuckinforloop/ticker/internal/timeutils"
)

var testULID = "000000000006AFVGQT5ZYC0GEK"
var testTime int64 = 1704102552000

func TestDAO(t *testing.T) {
	t.Run("default-test-dao", func(t *testing.T) {
		dao, err := NewTestDAO()
		assert.NoError(t, err)

		ulid := dao.ULIDSource.New(0)
		assert.Equal(t, testULID, ulid)
		assert.Equal(t, timeutils.FoundingTime, dao.TimeNow())
	})

	t.Run("test-dao-modified-timenow", func(t *testing.T) {
		dao, err := NewTestDAO(WithTimeNow(func() int64 {
			return testTime
		}))
		assert.NoError(t, err)

		ulid := dao.ULIDSource.New(0)
		assert.Equal(t, testULID, ulid)
		assert.NotEqual(t, timeutils.FoundingTime, dao.TimeNow())
	})
}
