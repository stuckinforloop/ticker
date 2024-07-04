package server

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/viper"
	"github.com/stuckinforloop/ticker/database"
	"github.com/stuckinforloop/ticker/internal/dao"
	"github.com/stuckinforloop/ticker/server/api"
	"go.uber.org/zap"
)

type Server struct {
	DAO *dao.DAO
}

func New() *Server {
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

	return &Server{dao}
}

func (s *Server) Start() (err error) {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, os.Interrupt)
	defer stop()

	api := api.New(s.DAO)
	api.RegisterRoutes()

	port := viper.GetViper().GetInt("port")
	server := &http.Server{
		Addr: fmt.Sprintf(":%d", port),
		BaseContext: func(_ net.Listener) context.Context {
			return ctx
		},
		Handler:      api.Mux(),
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	serverErr := make(chan error, 1)

	go func() {
		s.DAO.Logger.Info("starting server", zap.Int("port", port))
		serverErr <- server.ListenAndServe()
	}()

	select {
	case err = <-serverErr:

	case <-ctx.Done():
		stop()
	}

	s.DAO.Logger.Warn("received sigterm/interrupt signal, shutting down server...")
	err = server.Shutdown(context.Background())
	return
}
