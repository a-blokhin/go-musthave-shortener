package app

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"go-musthave-shortener/internal/api/shorterapi"
	"go-musthave-shortener/internal/config"
	"go-musthave-shortener/internal/middleware"
	"go-musthave-shortener/internal/repository"
	"go-musthave-shortener/internal/repository/shorterfilerepository"
	"go-musthave-shortener/internal/repository/shorterrepository"
	"go-musthave-shortener/internal/usecase/createshortlinkjsonusecase"
	"go-musthave-shortener/internal/usecase/createshortlinkusecase"
	"go-musthave-shortener/internal/usecase/redirectfromshortlinkusecase"
)

type DI struct {
	router *gin.Engine
	api    *shorterapi.ShortAPI
	config *config.Config
	logger *zap.Logger

	usecases struct {
		createShortLink       *createshortlinkusecase.Usecase
		createShortLinkJSON   *createshortlinkjsonusecase.Usecase
		redirectFromShortLink *redirectfromshortlinkusecase.Usecase
	}

	repos struct {
		shorterRepo repository.LinkRepository
	}

	httpServer *http.Server
}

func (d *DI) Init(config *config.Config) {
	d.config = config

	loggerConfig := zap.NewProductionConfig()
	loggerConfig.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	logger, err := loggerConfig.Build()
	if err != nil {
		panic(err)
	}
	d.logger = logger

	d.initRepos()
	d.initUsecases()
	d.initMux()
	d.initAPI()
}

func (d *DI) initRepos() {
	if d.config.FileStoragePath != "" {
		d.logger.Info("Using file storage", zap.String("path", d.config.FileStoragePath))
		d.repos.shorterRepo = shorterfilerepository.New(d.config.FileStoragePath)
	} else {
		d.logger.Info("Using in-memory storage")
		d.repos.shorterRepo = shorterrepository.New()
	}
}

func (d *DI) initUsecases() {
	d.usecases.createShortLink = createshortlinkusecase.New(d.repos.shorterRepo, d.logger, d.config.BaseURL)
	d.usecases.createShortLinkJSON = createshortlinkjsonusecase.New(d.repos.shorterRepo, d.logger, d.config.BaseURL)
	d.usecases.redirectFromShortLink = redirectfromshortlinkusecase.New(d.repos.shorterRepo, d.logger)
}

func (d *DI) initMux() {
	gin.SetMode(gin.ReleaseMode)
	d.router = gin.New()
	d.router.Use(gin.Recovery())
	d.router.Use(middleware.GzipMiddleware())
	d.router.Use(middleware.LoggingMiddleware(d.logger))
}

func (d *DI) initAPI() {
	d.api = shorterapi.New(
		d.config.BaseURL,
		d.usecases.createShortLink,
		d.usecases.createShortLinkJSON,
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
