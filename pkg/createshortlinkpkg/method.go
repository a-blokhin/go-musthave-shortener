package createshortlinkpkg

import "net/http"

const MethodPath = "/"
const Method = http.MethodPost

type Request struct {
	URL string
}

type Response struct {
	ShortURL string
}
