package redirectfromshortlinkusecase

type LinkRepo interface {
	Get(alias string) (string, error)
}
