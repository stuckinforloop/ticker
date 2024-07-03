package api

import (
	"github.com/go-chi/chi/v5"
	"github.com/stuckinforloop/ticker/internal/dao"
)

type API struct {
	dao *dao.DAO
	mux *chi.Mux
}

func New(dao *dao.DAO) *API {
	return &API{
		dao: dao,
		mux: chi.NewRouter(),
	}
}

func (a *API) Mux() *chi.Mux {
	return a.mux
}
