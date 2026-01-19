package shorterapi

import (
	"github.com/gin-gonic/gin"

	createshortlinkusecase "go-musthave-shortener/internal/handler/createShortLinkUsecase"
	redirectfromshortlinkusecase "go-musthave-shortener/internal/handler/redirectFromShortLinkUsecase"
	createshortlinkpkg "go-musthave-shortener/pkg/createShortLinkPkg"
)

type ShorterAPI struct {
	baseURL                string
	createShortLinkUseCase *createshortlinkusecase.Usecase
	redirectUseCase        *redirectfromshortlinkusecase.Usecase
}

func New(
	baseURL string,
	createShortLinkUseCase *createshortlinkusecase.Usecase,
	redirectUseCase *redirectfromshortlinkusecase.Usecase,
) *ShorterAPI {
	return &ShorterAPI{
		baseURL:                baseURL,
		createShortLinkUseCase: createShortLinkUseCase,
		redirectUseCase:        redirectUseCase,
	}
}

func (api *ShorterAPI) RegisterHandlers(router *gin.Engine) {

	router.POST(createshortlinkpkg.MethodPath, api.createShortLinkUseCase.Execute)

	router.GET("/:id", api.redirectUseCase.Execute)
}
