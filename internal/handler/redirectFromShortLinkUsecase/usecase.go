package redirectfromshortlinkusecase

import (
	"fmt"
	redirectfromshortlinkpkg "go-musthave-shortener/pkg/redirectFromShortLinkPkg"
)

type Usecase struct {
	linkRepo LinkRepo
}

func New(linkRepo LinkRepo) *Usecase {
	return &Usecase{
		linkRepo: linkRepo,
	}
}

func (u *Usecase) Execute(request redirectfromshortlinkpkg.Request) (redirectfromshortlinkpkg.Response, error) {
	originalURL, err := u.linkRepo.Get(request.Alias)
	if err != nil {
		return redirectfromshortlinkpkg.Response{}, fmt.Errorf("failed to get URL for alias: %w", err)
	}

	return redirectfromshortlinkpkg.Response{
		OriginalURL: originalURL,
	}, nil
}
