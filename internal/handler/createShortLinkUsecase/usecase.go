package createshortlinkusecase

import (
	"fmt"
	createshortlinkpkg "go-musthave-shortener/pkg/createShortLinkPkg"
)


type Usecase struct {
	linkRepo LinkRepo
}


func New(linkRepo LinkRepo) *Usecase {
	return &Usecase{
		linkRepo: linkRepo,
	}
}


func (u *Usecase) Execute(request createshortlinkpkg.Request) (createshortlinkpkg.Response, error) {
	alias, err := u.linkRepo.Add(request.URL)
	if err != nil {
		return createshortlinkpkg.Response{}, fmt.Errorf("failed to create short link: %w", err)
	}

	return createshortlinkpkg.Response{
		ShortURL: alias,
	}, nil
}
