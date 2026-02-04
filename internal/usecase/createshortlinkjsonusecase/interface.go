package createshortlinkjsonusecase

type LinkRepo interface {
	Add(url string) (alias string, err error)
}