package dao

import (
	"database/sql"
	"fmt"
	"math/rand"
	"time"

	"go.uber.org/zap"

	"github.com/stuckinforloop/ticker/internal/logger"
	"github.com/stuckinforloop/ticker/internal/timeutils"
	"github.com/stuckinforloop/ticker/internal/ulid"
)

type Option func(*DAO)

type DAO struct {
	ro         *sql.DB
	rw         *sql.DB
	Logger     *zap.Logger
	ULIDSource *ulid.Source
	TimeNow    timeutils.TimeNow
}

func New(opts ...Option) (*DAO, error) {
	logger, err := logger.New("dev")
	if err != nil {
		return nil, fmt.Errorf("setup logger: %w", err)
	}

	source := &ulid.Source{Rand: rand.New(rand.NewSource(time.Now().Unix()))}
	timeNow := func() int64 {
		return time.Now().Unix()
	}

	dao := DAO{
		Logger:     logger,
		ULIDSource: source,
		TimeNow:    timeNow,
	}

	for _, opt := range opts {
		opt(&dao)
	}

	return &dao, nil
}

func NewTestDAO(opts ...Option) (*DAO, error) {
	dao, err := New(
		WithLogger(zap.NewNop()),
		WithULIDSource(&ulid.Source{Rand: rand.New(rand.NewSource(0))}),
		WithTimeNow(func() int64 {
			return timeutils.FoundingTime
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("setup test dao: %w", err)
	}

	for _, opt := range opts {
		opt(dao)
	}

	return dao, nil
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

func WithDB(ro, rw *sql.DB) Option {
	return func(d *DAO) {
		d.ro = ro
		d.rw = rw
	}
}

func (dao *DAO) RO() *sql.DB {
	return dao.ro
}

func (dao *DAO) RW() *sql.DB {
	return dao.rw
}
