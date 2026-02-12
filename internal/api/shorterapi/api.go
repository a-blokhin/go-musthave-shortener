package shorterapi

import (
	"github.com/gin-gonic/gin"

	"go-musthave-shortener/internal/usecase/createshortlinkbatchusecase"
	"go-musthave-shortener/internal/usecase/createshortlinkjsonusecase"
	"go-musthave-shortener/internal/usecase/createshortlinkusecase"
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
}

func New(
	baseURL string,
	createShortLinkUseCase *createshortlinkusecase.Usecase,
	createShortLinkJSONUseCase *createshortlinkjsonusecase.Usecase,
	createShortLinkBatchUseCase *createshortlinkbatchusecase.Usecase,
	redirectUseCase *redirectfromshortlinkusecase.Usecase,
	pingDatabase *pingdatabaseusecase.Usecase,
) *ShortAPI {
	return &ShortAPI{
		baseURL:                     baseURL,
		createShortLinkUseCase:      createShortLinkUseCase,
		createShortLinkJSONUseCase:  createShortLinkJSONUseCase,
		createShortLinkBatchUseCase: createShortLinkBatchUseCase,
		redirectUseCase:             redirectUseCase,
		pingDatabase:                pingDatabase,
	}
}

func (api *ShortAPI) RegisterHandlers(router *gin.Engine) {
	router.POST(createshortlinkpkg.MethodPath, api.createShortLinkUseCase.Execute)
	router.POST(createshortlinkjsonpkg.MethodPath, api.createShortLinkJSONUseCase.Execute)
	router.POST(createshortlinkbatchpkg.MethodPath, api.createShortLinkBatchUseCase.Execute)
	router.GET("/:id", api.redirectUseCase.Execute)
	router.GET("/ping", api.pingDatabase.Execute)
}
