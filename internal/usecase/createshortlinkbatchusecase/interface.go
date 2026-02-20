package createshortlinkbatchusecase

type LinkRepo interface {
	AddBatch(urls []string) ([]string, error)
}
