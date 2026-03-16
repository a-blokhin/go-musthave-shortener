package app

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"go-musthave-shortener/internal/api/shorterapi"
	"go-musthave-shortener/internal/audit"
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
	"go-musthave-shortener/internal/usecase/deleteurlsusecase"
	"go-musthave-shortener/internal/usecase/getstatsusecase"
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
	audit  *audit.Service

	usecases struct {
		createShortLink       *createshortlinkusecase.Usecase
		createShortLinkJSON   *createshortlinkjsonusecase.Usecase
		createShortLinkBatch  *createshortlinkbatchusecase.Usecase
		redirectFromShortLink *redirectfromshortlinkusecase.Usecase
		getUserURLs           *getuserurlsusecase.Usecase
		pingDatabase          *pingdatabaseusecase.Usecase
		deleteURLs            *deleteurlsusecase.Usecase
		getStats              *getstatsusecase.Usecase
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
			return err
		}
		d.db = db
		d.logger.Info("Connected to PostgreSQL database")
	}

	err = d.initRepos()
	if err != nil {
		return err
	}
	err = d.initAudit()
	if err != nil {
		return err
	}
	d.initUsecases()
	d.initMux()
	d.initAPI()

	return nil
}

func (d *DI) initRepos() error {
	if d.db != nil {
		d.logger.Info("Using PostgreSQL database storage")

		migrator := migration.New(d.logger, "migrations")
		if err := migrator.Up(d.config.DatabaseDSN); err != nil {
			return err
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
	return nil
}

func (d *DI) initAudit() error {
	d.audit = audit.NewService(d.logger)

	if d.config.AuditFile != "" {
		fileReceiver, err := audit.NewFileReceiver(d.config.AuditFile)
		if err != nil {
			return err
		}
		d.audit.AddReceiver(fileReceiver)
		d.logger.Info("File audit receiver enabled", zap.String("path", d.config.AuditFile))
	}

	if d.config.AuditURL != "" {
		httpReceiver, err := audit.NewHTTPReceiver(d.config.AuditURL)
		if err != nil {
			return err
		}
		d.audit.AddReceiver(httpReceiver)
		d.logger.Info("HTTP audit receiver enabled", zap.String("url", d.config.AuditURL))
	}
	return nil
}

func (d *DI) initUsecases() {
	d.usecases.createShortLink = createshortlinkusecase.New(d.repos.shorterRepo, d.logger, d.config.BaseURL, d.audit)
	d.usecases.createShortLinkJSON = createshortlinkjsonusecase.New(d.repos.shorterRepo, d.logger, d.config.BaseURL, d.audit)
	d.usecases.createShortLinkBatch = createshortlinkbatchusecase.New(d.repos.shorterRepo, d.logger, d.config.BaseURL)
	d.usecases.redirectFromShortLink = redirectfromshortlinkusecase.New(d.repos.shorterRepo, d.logger, d.audit)
	d.usecases.getUserURLs = getuserurlsusecase.New(d.repos.shorterRepo, d.logger, d.config.BaseURL)
	d.usecases.pingDatabase = pingdatabaseusecase.New(d.db, d.logger)
	d.usecases.deleteURLs = deleteurlsusecase.New(d.repos.shorterRepo, d.logger, d.config.DeleteURLs)
	d.usecases.getStats = getstatsusecase.New(d.repos.shorterRepo, d.logger)
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
		d.usecases.deleteURLs,
		d.usecases.getStats,
		d.config.TrustedSubnet,
		d.logger,
	)
	d.api.RegisterHandlers(d.router)
}

func (d *DI) StartServer() error {
	d.httpServer = &http.Server{
		Addr:    d.config.ServerAddress,
		Handler: d.router,
	}

	if d.config.EnableHTTPS {
		if d.config.SSLCertPath == "" {
			return errors.New("HTTPS is enabled but SSL certificate is not specified")
		}
		if d.config.SSLKeyPath == "" {
			return errors.New("HTTPS is enabled but key path is not specified")
		}
		return d.httpServer.ListenAndServeTLS(d.config.SSLCertPath, d.config.SSLKeyPath)
	}

	return d.httpServer.ListenAndServe()
}

func (d *DI) StopServer(ctx context.Context) error {
	if d.usecases.deleteURLs != nil {
		d.usecases.deleteURLs.Close()
	}

	if d.audit != nil {
		d.audit.Close()
	}

	if d.db != nil {
		d.db.Close()
	}

	if d.httpServer != nil {
		return d.httpServer.Shutdown(ctx)
	}
	return nil
}
