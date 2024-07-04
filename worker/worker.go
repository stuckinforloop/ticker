package worker

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/stuckinforloop/ticker/database"
	"github.com/stuckinforloop/ticker/internal/dao"
)

type Worker struct {
	DAO *dao.DAO
}

func New() *Worker {
	db, err := database.New()
	if err != nil {
		log.Fatal(err)
	}

	dao, err := dao.New(
		dao.WithDB(db.RO, db.RW),
	)
	if err != nil {
		log.Fatal(err)
	}

	return &Worker{dao}
}

func (s *Worker) Run(handler func(ctx context.Context) error) {
	s.DAO.Logger.Info("starting worker")

	ctx := context.Background()
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := handler(ctx); err != nil {
			panic(err)
		}
	}()

	c := make(chan os.Signal, 1)

	// accept graceful shutdowns when quit via SIGINT (Ctrl+C) or SIGTERM.
	// SIGKILL, SIGQUIT will not be caught.
	signal.Notify(c, syscall.SIGTERM, syscall.SIGINT)

	// Block until signal is received.
	<-c
	wg.Wait()
}
