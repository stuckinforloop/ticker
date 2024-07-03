package server

import (
	"log"
	"net/http"

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

func (s *Server) Start() {
	api := api.New(s.DAO)
	api.RegisterRoutes()

	s.DAO.Logger.Info("starting server", zap.String("port", "9000"))
	if err := http.ListenAndServe(":9000", api.Mux()); err != nil {
		log.Fatal(err)
	}
}
