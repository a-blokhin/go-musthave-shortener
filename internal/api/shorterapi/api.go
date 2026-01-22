package shorterapi

import (
	"github.com/gin-gonic/gin"

	"go-musthave-shortener/internal/usecase/createshortlinkusecase"
	"go-musthave-shortener/internal/usecase/redirectfromshortlinkusecase"
	"go-musthave-shortener/pkg/createshortlinkpkg"
)

type ShortAPI struct {
	baseURL                string
	createShortLinkUseCase *createshortlinkusecase.Usecase
	redirectUseCase        *redirectfromshortlinkusecase.Usecase
}

func New(
	baseURL string,
	createShortLinkUseCase *createshortlinkusecase.Usecase,
	redirectUseCase *redirectfromshortlinkusecase.Usecase,
) *ShortAPI {
	return &ShortAPI{
		baseURL:                baseURL,
		createShortLinkUseCase: createShortLinkUseCase,
		redirectUseCase:        redirectUseCase,
	}
}

func (api *ShortAPI) RegisterHandlers(router *gin.Engine) {

	router.POST(createshortlinkpkg.MethodPath, api.createShortLinkUseCase.Execute)

	router.GET("/:id", api.redirectUseCase.Execute)
}
