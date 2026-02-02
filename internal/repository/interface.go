package repository

type LinkRepository interface {
	Add(url string) (string, error)
	Get(alias string) (string, error)
}