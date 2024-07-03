package dao

import (
	"math/rand"

	"go.uber.org/zap"

	"github.com/stuckinforloop/ticker/internal/database"
	"github.com/stuckinforloop/ticker/internal/timeutils"
	"github.com/stuckinforloop/ticker/internal/ulid"
)

type Option func(*DAO)

type DAO struct {
	DB         *database.DB
	Logger     *zap.Logger
	ULIDSource *ulid.Source
	TimeNow    timeutils.TimeNow
}

func New(opts ...Option) *DAO {
	dao := DAO{}

	for _, opt := range opts {
		opt(&dao)
	}

	return &dao
}

func NewTestDAO(opts ...Option) *DAO {
	dao := New(
		WithLogger(zap.NewNop()),
		WithULIDSource(&ulid.Source{Rand: rand.New(rand.NewSource(0))}),
		WithTimeNow(func() int64 {
			return timeutils.FoundingTime
		}),
	)

	for _, opt := range opts {
		opt(dao)
	}

	return dao
}

func WithLogger(logger *zap.Logger) Option {
	return func(d *DAO) {
		d.Logger = logger
	}
}

func WithULIDSource(source *ulid.Source) Option {
	return func(d *DAO) {
		d.ULIDSource = source
	}
}

func WithTimeNow(timeNow timeutils.TimeNow) Option {
	return func(d *DAO) {
		d.TimeNow = timeNow
	}
}
