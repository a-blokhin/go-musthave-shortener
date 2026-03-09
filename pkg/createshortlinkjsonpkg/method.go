package createshortlinkjsonpkg

import "net/http"

const MethodPath = "/api/shorten"
const Method = http.MethodPost

type Request struct {
	URL string `json:"url"`
}

type Response struct {
	Result string `json:"result"`
}
