package app

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"go-musthave-shortener/internal/api/shorterapi"
	"go-musthave-shortener/internal/config"
	"go-musthave-shortener/internal/database"
	"go-musthave-shortener/internal/middleware"
	"go-musthave-shortener/internal/migration"
	"go-musthave-shortener/internal/repository"
	"go-musthave-shortener/internal/repository/postgresrepository"
	"go-musthave-shortener/internal/repository/shorterfilerepository"
	"go-musthave-shortener/internal/repository/shorterrepository"
	"go-musthave-shortener/internal/usecase/createshortlinkbatchusecase"
	"go-musthave-shortener/internal/usecase/createshortlinkjsonusecase"
	"go-musthave-shortener/internal/usecase/createshortlinkusecase"
	"go-musthave-shortener/internal/usecase/getuserurlsusecase"
	"go-musthave-shortener/internal/usecase/pingdatabaseusecase"
	"go-musthave-shortener/internal/usecase/redirectfromshortlinkusecase"
)

type DI struct {
	router *gin.Engine
	api    *shorterapi.ShortAPI
	config *config.Config
	logger *zap.Logger
	db     *database.DB

	usecases struct {
		createShortLink       *createshortlinkusecase.Usecase
		createShortLinkJSON   *createshortlinkjsonusecase.Usecase
		createShortLinkBatch  *createshortlinkbatchusecase.Usecase
		redirectFromShortLink *redirectfromshortlinkusecase.Usecase
		getUserURLs           *getuserurlsusecase.Usecase
		pingDatabase          *pingdatabaseusecase.Usecase
	}

	repos struct {
		shorterRepo repository.LinkRepository
	}

	httpServer *http.Server
}

func (d *DI) Init(config *config.Config) error {
	d.config = config

	loggerConfig := zap.NewProductionConfig()
	loggerConfig.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	logger, err := loggerConfig.Build()
	if err != nil {
		return err
	}
	d.logger = logger

	if config.DatabaseDSN != "" {
		db, err := database.New(config.DatabaseDSN)
		if err != nil {
			d.logger.Error("Failed to connect to database", zap.Error(err))
			return err
		}
		d.db = db
		d.logger.Info("Connected to PostgreSQL database")
	}

	d.initRepos()
	d.initUsecases()
	d.initMux()
	d.initAPI()

	return nil
}

func (d *DI) initRepos() {
	if d.db != nil {
		d.logger.Info("Using PostgreSQL database storage")

		migrator := migration.New(d.logger, "migrations")
		if err := migrator.Up(d.config.DatabaseDSN); err != nil {
			d.logger.Fatal("Failed to run database migrations", zap.Error(err))
		}

		postgresRepo := postgresrepository.New(d.db.Pool(), d.logger)
		d.repos.shorterRepo = postgresRepo
	} else if d.config.FileStoragePath != "" {
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
	d.usecases.createShortLinkBatch = createshortlinkbatchusecase.New(d.repos.shorterRepo, d.logger, d.config.BaseURL)
	d.usecases.redirectFromShortLink = redirectfromshortlinkusecase.New(d.repos.shorterRepo, d.logger)
	d.usecases.getUserURLs = getuserurlsusecase.New(d.repos.shorterRepo, d.logger, d.config.BaseURL)
	d.usecases.pingDatabase = pingdatabaseusecase.New(d.db, d.logger)
}

func (d *DI) initMux() {
	gin.SetMode(gin.ReleaseMode)
	d.router = gin.New()
	d.router.Use(gin.Recovery())
	d.router.Use(middleware.AuthMiddleware(d.logger))
	d.router.Use(middleware.GzipMiddleware())
	d.router.Use(middleware.LoggingMiddleware(d.logger))
}

func (d *DI) initAPI() {
	d.api = shorterapi.New(
		d.config.BaseURL,
		d.usecases.createShortLink,
		d.usecases.createShortLinkJSON,
		d.usecases.createShortLinkBatch,
		d.usecases.redirectFromShortLink,
		d.usecases.pingDatabase,
		d.usecases.getUserURLs,
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
	// Close database connection
	if d.db != nil {
		d.db.Close()
	}

	if d.httpServer != nil {
		return d.httpServer.Shutdown(ctx)
	}
	return nil
}
