package app

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	"go-musthave-shortener/api/proto"
	"go-musthave-shortener/internal/api/grpcapi"
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
	"go-musthave-shortener/internal/usecase/expandurlusecase"
	"go-musthave-shortener/internal/usecase/getstatsusecase"
	"go-musthave-shortener/internal/usecase/getstatsusecasegeneric"
	"go-musthave-shortener/internal/usecase/getuserurlsusecase"
	"go-musthave-shortener/internal/usecase/getuserurlsusecasegeneric"
	"go-musthave-shortener/internal/usecase/grpcexpandurlusecase"
	"go-musthave-shortener/internal/usecase/grpcgetuserurlsusecase"
	"go-musthave-shortener/internal/usecase/grpcshortenurlusecase"
	"go-musthave-shortener/internal/usecase/pingdatabaseusecase"
	"go-musthave-shortener/internal/usecase/redirectfromshortlinkusecase"
	"go-musthave-shortener/internal/usecase/shortenurlusecase"
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
		getStatsHandler       *getstatsusecase.Usecase
	}

	grpcUsecases struct {
		shortenURL  *grpcshortenurlusecase.Usecase
		expandURL   *grpcexpandurlusecase.Usecase
		getUserURLs *grpcgetuserurlsusecase.Usecase
	}

	genericUsecases struct {
		shortenURL  *shortenurlusecase.ShortenURLUsecase
		expandURL   *expandurlusecase.ExpandURLUsecase
		getUserURLs *getuserurlsusecasegeneric.GetUserURLsUsecase
		getStats    *getstatsusecasegeneric.GetStatsUsecase
	}

	repos struct {
		shorterRepo repository.LinkRepository
	}

	httpServer *http.Server
	grpcServer *grpc.Server
	grpcAPI    *grpcapi.Server
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
	d.genericUsecases.shortenURL = shortenurlusecase.New(d.repos.shorterRepo, d.logger, d.config.BaseURL, d.audit)
	d.genericUsecases.expandURL = expandurlusecase.New(d.repos.shorterRepo, d.logger, d.audit)
	d.genericUsecases.getUserURLs = getuserurlsusecasegeneric.New(d.repos.shorterRepo, d.logger, d.config.BaseURL)
	d.genericUsecases.getStats = getstatsusecasegeneric.New(d.repos.shorterRepo, d.logger)

	d.usecases.createShortLink = createshortlinkusecase.New(d.genericUsecases.shortenURL, d.logger)
	d.usecases.createShortLinkJSON = createshortlinkjsonusecase.New(d.repos.shorterRepo, d.logger, d.config.BaseURL, d.audit)
	d.usecases.createShortLinkBatch = createshortlinkbatchusecase.New(d.repos.shorterRepo, d.logger, d.config.BaseURL)
	d.usecases.redirectFromShortLink = redirectfromshortlinkusecase.New(d.genericUsecases.expandURL, d.logger)
	d.usecases.getUserURLs = getuserurlsusecase.New(d.genericUsecases.getUserURLs, d.logger)
	d.usecases.pingDatabase = pingdatabaseusecase.New(d.db, d.logger)
	d.usecases.deleteURLs = deleteurlsusecase.New(d.repos.shorterRepo, d.logger, d.config.DeleteURLs)
	d.usecases.getStatsHandler = getstatsusecase.NewHTTPHandler(d.genericUsecases.getStats, d.logger)

	d.grpcUsecases.shortenURL = grpcshortenurlusecase.New(d.genericUsecases.shortenURL, d.logger)
	d.grpcUsecases.expandURL = grpcexpandurlusecase.New(d.genericUsecases.expandURL, d.logger)
	d.grpcUsecases.getUserURLs = grpcgetuserurlsusecase.New(d.genericUsecases.getUserURLs, d.logger)
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
		d.usecases.getStatsHandler,
		d.config.TrustedSubnet,
		d.logger,
	)
	d.api.RegisterHandlers(d.router)

	d.grpcAPI = grpcapi.New(
		d.grpcUsecases.shortenURL,
		d.grpcUsecases.expandURL,
		d.grpcUsecases.getUserURLs,
	)
}

func (d *DI) StartServer() error {
	d.httpServer = &http.Server{
		Addr:    d.config.ServerAddress,
		Handler: d.router,
	}

	d.grpcServer = grpc.NewServer()
	proto.RegisterShortenerServiceServer(d.grpcServer, d.grpcAPI)

	d.logger.Info("Starting gRPC server", zap.String("address", d.config.GRPCServerAddress))
	grpcLis, err := net.Listen("tcp", d.config.GRPCServerAddress)
	if err != nil {
		return fmt.Errorf("gRPC server listen error: %w", err)
	}
	errCh := make(chan error)
	go func() {
		err := d.grpcServer.Serve(grpcLis)
		errCh <- err
	}()

	d.logger.Info("Starting HTTP server", zap.String("address", d.config.ServerAddress))
	if d.config.EnableHTTPS {
		if d.config.SSLCertPath == "" {
			return errors.New("HTTPS is enabled but SSL certificate is not specified")
		}
		if d.config.SSLKeyPath == "" {
			return errors.New("HTTPS is enabled but key path is not specified")
		}
		return d.httpServer.ListenAndServeTLS(d.config.SSLCertPath, d.config.SSLKeyPath)
	}
	err = d.httpServer.ListenAndServe()
	if err != nil {
		return fmt.Errorf("HTTP server listen error: %w", err)
	}
	return <-errCh
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

	if d.grpcServer != nil {
		d.grpcServer.GracefulStop()
	}

	if d.httpServer != nil {
		return d.httpServer.Shutdown(ctx)
	}
	return nil
}
