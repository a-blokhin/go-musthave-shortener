package redirectfromshortlinkpkg

import "net/http"

const MethodPath = "/{id}"
const Method = http.MethodGet

type Request struct {
	Alias string
}

type Response struct {
	OriginalURL string
}
