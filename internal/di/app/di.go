package app

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	shorterapi "go-musthave-shortener/internal/api/shorterApi"
	createshortlinkusecase "go-musthave-shortener/internal/handler/createShortLinkUsecase"
	redirectfromshortlinkusecase "go-musthave-shortener/internal/handler/redirectFromShortLinkUsecase"
	shorterrepository "go-musthave-shortener/internal/repository/shorterRepository"
)

type DI struct {
	router *gin.Engine
	api    *shorterapi.ShorterAPI
	config *Config
	logger *zap.Logger

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
	
	
	loggerConfig := zap.NewProductionConfig()
	loggerConfig.Level = zap.NewAtomicLevelAt(zap.ErrorLevel)
	logger, _ := loggerConfig.Build()
	d.logger = logger

	d.initRepos()
	d.initUsecases()
	d.initMux()
	d.initAPI()
}

func (d *DI) initRepos() {
	d.repos.shorterRepository = shorterrepository.New()
}

func (d *DI) initUsecases() {
	d.usecases.createShortLink = createshortlinkusecase.New(d.repos.shorterRepository, d.logger, d.config.BaseURL)
	d.usecases.redirectFromShortLink = redirectfromshortlinkusecase.New(d.repos.shorterRepository, d.logger)
}

func (d *DI) initMux() {
	gin.SetMode(gin.ReleaseMode)
	d.router = gin.New()
	d.router.Use(gin.Recovery())
}

func (d *DI) initAPI() {
	d.api = shorterapi.New(
		d.config.BaseURL,
		d.usecases.createShortLink,
		d.usecases.redirectFromShortLink,
	)
	d.api.RegisterHandlers(d.router)
}

func (d *DI) StartServer() error {
	d.httpServer = &http.Server{
		Addr:    d.config.ServerAddress,
		Handler: d.router,
	}

	return d.httpServer.ListenAndServe()
}

func (d *DI) StopServer(ctx context.Context) error {
	if d.httpServer != nil {
		return d.httpServer.Shutdown(ctx)
	}
	return nil
}
