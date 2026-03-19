// Package shorterapi provides HTTP handlers for the URL shortener service.
// It registers routes for creating short links, redirecting to original URLs,
// and managing user URLs.
package shorterapi

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"go-musthave-shortener/internal/middleware"
	"go-musthave-shortener/internal/usecase/createshortlinkbatchusecase"
	"go-musthave-shortener/internal/usecase/createshortlinkjsonusecase"
	"go-musthave-shortener/internal/usecase/createshortlinkusecase"
	"go-musthave-shortener/internal/usecase/deleteurlsusecase"
	"go-musthave-shortener/internal/usecase/getstatsusecase"
	"go-musthave-shortener/internal/usecase/getuserurlsusecase"
	"go-musthave-shortener/internal/usecase/pingdatabaseusecase"
	"go-musthave-shortener/internal/usecase/redirectfromshortlinkusecase"
	"go-musthave-shortener/pkg/createshortlinkbatchpkg"
	"go-musthave-shortener/pkg/createshortlinkjsonpkg"
	"go-musthave-shortener/pkg/createshortlinkpkg"
)

type ShortAPI struct {
	baseURL                     string
	createShortLinkUseCase      *createshortlinkusecase.Usecase
	createShortLinkJSONUseCase  *createshortlinkjsonusecase.Usecase
	createShortLinkBatchUseCase *createshortlinkbatchusecase.Usecase
	redirectUseCase             *redirectfromshortlinkusecase.Usecase
	pingDatabase                *pingdatabaseusecase.Usecase
	getUserURLs                 *getuserurlsusecase.Usecase
	deleteURLs                  *deleteurlsusecase.Usecase
	getStatsHandler             *getstatsusecase.Usecase
	trustedSubnet               string
	logger                      *zap.Logger
}

func New(
	baseURL string,
	createShortLinkUseCase *createshortlinkusecase.Usecase,
	createShortLinkJSONUseCase *createshortlinkjsonusecase.Usecase,
	createShortLinkBatchUseCase *createshortlinkbatchusecase.Usecase,
	redirectUseCase *redirectfromshortlinkusecase.Usecase,
	pingDatabase *pingdatabaseusecase.Usecase,
	getUserURLs *getuserurlsusecase.Usecase,
	deleteURLs *deleteurlsusecase.Usecase,
	getStatsHandler *getstatsusecase.Usecase,
	trustedSubnet string,
	logger *zap.Logger,
) *ShortAPI {
	return &ShortAPI{
		baseURL:                     baseURL,
		createShortLinkUseCase:      createShortLinkUseCase,
		createShortLinkJSONUseCase:  createShortLinkJSONUseCase,
		createShortLinkBatchUseCase: createShortLinkBatchUseCase,
		redirectUseCase:             redirectUseCase,
		pingDatabase:                pingDatabase,
		getUserURLs:                 getUserURLs,
		deleteURLs:                  deleteURLs,
		getStatsHandler:             getStatsHandler,
		trustedSubnet:               trustedSubnet,
		logger:                      logger,
	}
}

// RegisterHandlers registers all HTTP handlers for the URL shortener service.
//
// It registers the following routes:
//   - POST / - create a short link from plain text
//   - POST /api/shorten - create a short link from JSON
//   - POST /api/shorten/batch - create multiple short links from JSON
//   - GET /:id - redirect from short link to original URL
//   - GET /ping - check database connectivity
//   - GET /api/user/urls - get all URLs for the authenticated user
//   - DELETE /api/user/urls - delete URLs for the authenticated user
//   - GET /api/internal/stats - get service statistics (requires trusted subnet)
//
// Parameters:
//   - router: the Gin router to register handlers on
func (api *ShortAPI) RegisterHandlers(router *gin.Engine) {
	router.POST(createshortlinkpkg.MethodPath, api.createShortLinkUseCase.Execute)
	router.POST(createshortlinkjsonpkg.MethodPath, api.createShortLinkJSONUseCase.Execute)
	router.POST(createshortlinkbatchpkg.MethodPath, api.createShortLinkBatchUseCase.Execute)
	router.GET("/:id", api.redirectUseCase.Execute)
	router.GET("/ping", api.pingDatabase.Execute)
	router.GET("/api/user/urls", api.getUserURLs.Execute)
	router.DELETE("/api/user/urls", api.deleteURLs.Execute)

	if api.getStatsHandler != nil {
		if api.trustedSubnet != "" {
			router.GET("/api/internal/stats", middleware.TrustedSubnetMiddleware(api.trustedSubnet, api.logger), api.getStatsHandler.Execute)
		} else {
			router.GET("/api/internal/stats", middleware.TrustedSubnetMiddleware("", api.logger), api.getStatsHandler.Execute)
		}
	}
}
