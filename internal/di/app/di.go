package app

import (
	"context"
	"net/http"

	shorterapi "go-musthave-shortener/internal/api/shorterApi"
	createshortlinkusecase "go-musthave-shortener/internal/handler/createShortLinkUsecase"
	redirectfromshortlinkusecase "go-musthave-shortener/internal/handler/redirectFromShortLinkUsecase"
	shorterrepository "go-musthave-shortener/internal/repository/shorterRepository"
)

type DI struct {
	mux    *http.ServeMux
	api    *shorterapi.ShorterAPI
	config *Config

	usecases struct {
		createShortLink       *createshortlinkusecase.Usecase
		redirectFromShortLink *redirectfromshortlinkusecase.Usecase
	}

	repos struct {
		shorterRepository *shorterrepository.Repo
	}

	httpServer *http.Server
}

func (d *DI) Init(config *Config) {
	d.config = config

	d.initRepos()
	d.initUsecases()
	d.initMux()
	d.initAPI()
}

func (d *DI) initRepos() {
	d.repos.shorterRepository = shorterrepository.New()
}

func (d *DI) initUsecases() {
	d.usecases.createShortLink = createshortlinkusecase.New(d.repos.shorterRepository)
	d.usecases.redirectFromShortLink = redirectfromshortlinkusecase.New(d.repos.shorterRepository)
}

func (d *DI) initMux() {
	d.mux = http.NewServeMux()
}

func (d *DI) initAPI() {
	d.api = shorterapi.New(
		d.config.BaseURL,
		d.usecases.createShortLink,
		d.usecases.redirectFromShortLink,
	)
	d.api.RegisterHandlers(d.mux)
}

func (d *DI) StartServer() error {
	d.httpServer = &http.Server{
		Addr:    d.config.ServerAddress,
		Handler: d.mux,
	}

	return d.httpServer.ListenAndServe()
}

func (d *DI) StopServer(ctx context.Context) error {
	if d.httpServer != nil {
		return d.httpServer.Shutdown(ctx)
	}
	return nil
}
