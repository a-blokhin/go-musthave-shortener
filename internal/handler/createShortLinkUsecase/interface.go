package createshortlinkusecase


type LinkRepo interface {
	Add(url string) (string, error)
}
