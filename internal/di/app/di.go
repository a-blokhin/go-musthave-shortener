package app

import (
	"context"
	"net/http"
	"net/http/pprof"

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
		} else {
			d.audit.AddReceiver(fileReceiver)
			d.logger.Info("File audit receiver enabled", zap.String("path", d.config.AuditFile))
		}
	}

	if d.config.AuditURL != "" {
		httpReceiver, err := audit.NewHTTPReceiver(d.config.AuditURL)
		if err != nil {
			return err
		} else {
			d.audit.AddReceiver(httpReceiver)
			d.logger.Info("HTTP audit receiver enabled", zap.String("url", d.config.AuditURL))
		}
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
}

func (d *DI) initMux() {
	gin.SetMode(gin.ReleaseMode)
	d.router = gin.New()
	d.router.Use(gin.Recovery())
	d.router.Use(middleware.AuthMiddleware(d.logger))
	d.router.Use(middleware.GzipMiddleware())
	d.router.Use(middleware.LoggingMiddleware(d.logger))

	// Add pprof endpoints for profiling
	pprofGroup := d.router.Group("/debug/pprof")
	{
		pprofGroup.GET("/", gin.WrapF(http.HandlerFunc(pprof.Index)))
		pprofGroup.GET("/cmdline", gin.WrapF(http.HandlerFunc(pprof.Cmdline)))
		pprofGroup.GET("/profile", gin.WrapF(http.HandlerFunc(pprof.Profile)))
		pprofGroup.POST("/symbol", gin.WrapF(http.HandlerFunc(pprof.Symbol)))
		pprofGroup.GET("/symbol", gin.WrapF(http.HandlerFunc(pprof.Symbol)))
		pprofGroup.GET("/trace", gin.WrapF(http.HandlerFunc(pprof.Trace)))
		pprofGroup.GET("/allocs", gin.WrapF(http.HandlerFunc(pprof.Handler("allocs").ServeHTTP)))
		pprofGroup.GET("/block", gin.WrapF(http.HandlerFunc(pprof.Handler("block").ServeHTTP)))
		pprofGroup.GET("/goroutine", gin.WrapF(http.HandlerFunc(pprof.Handler("goroutine").ServeHTTP)))
		pprofGroup.GET("/heap", gin.WrapF(http.HandlerFunc(pprof.Handler("heap").ServeHTTP)))
		pprofGroup.GET("/mutex", gin.WrapF(http.HandlerFunc(pprof.Handler("mutex").ServeHTTP)))
		pprofGroup.GET("/threadcreate", gin.WrapF(http.HandlerFunc(pprof.Handler("threadcreate").ServeHTTP)))
	}
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
