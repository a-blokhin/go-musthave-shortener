package shorterapi

import (
	createshortlinkusecase "go-musthave-shortener/internal/handler/createShortLinkUsecase"
	redirectfromshortlinkusecase "go-musthave-shortener/internal/handler/redirectFromShortLinkUsecase"
	createshortlinkpkg "go-musthave-shortener/pkg/createShortLinkPkg"
	redirectfromshortlinkpkg "go-musthave-shortener/pkg/redirectFromShortLinkPkg"
	"io"
	"net/http"
	"strings"
)

type ShorterAPI struct {
	baseURL                      string
	createshortlinkusecase       *createshortlinkusecase.Usecase
	redirectfromshortlinkusecase *redirectfromshortlinkusecase.Usecase
}

func New(
	baseURL string,
	createshortlinkusecase *createshortlinkusecase.Usecase,
	redirectfromshortlinkusecase *redirectfromshortlinkusecase.Usecase,
) *ShorterAPI {
	return &ShorterAPI{
		baseURL:                      baseURL,
		createshortlinkusecase:       createshortlinkusecase,
		redirectfromshortlinkusecase: redirectfromshortlinkusecase,
	}
}

func (api *ShorterAPI) RegisterHandlers(mux *http.ServeMux) {

	mux.HandleFunc("/", api.handleRequest)
}

func (api *ShorterAPI) handleRequest(w http.ResponseWriter, r *http.Request) {

	switch {
	case r.Method == createshortlinkpkg.Method && r.URL.Path == createshortlinkpkg.MethodPath:
		api.createShortURL(w, r)
	case r.Method == redirectfromshortlinkpkg.Method && r.URL.Path != "/":
		api.redirectToOriginalURL(w, r)
	default:
		http.Error(w, "Bad Request", http.StatusBadRequest)
	}
}

func (api *ShorterAPI) createShortURL(w http.ResponseWriter, r *http.Request) {

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	url := strings.TrimSpace(string(body))
	if url == "" {
		http.Error(w, "Bad Request: URL is required", http.StatusBadRequest)
		return
	}

	request := createshortlinkpkg.Request{
		URL: url,
	}

	response, err := api.createshortlinkusecase.Execute(request)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(api.baseURL + "/" + response.ShortURL))
}

func (api *ShorterAPI) redirectToOriginalURL(w http.ResponseWriter, r *http.Request) {

	alias := r.URL.Path[1:]
	if alias == "" {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	request := redirectfromshortlinkpkg.Request{
		Alias: alias,
	}

	response, err := api.redirectfromshortlinkusecase.Execute(request)
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	w.Header().Set("Location", response.OriginalURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
}
