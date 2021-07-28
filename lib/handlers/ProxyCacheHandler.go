package handlers

import (
	"io"
	"log"
	mdath "mdath/lib"
	"net/http"
)

type ProxyCacheHandler struct {
	origins   []string
	validator *mdath.RequestValidator
}

func CreateProxyCacheHandler(origins []string, validator *mdath.RequestValidator) (instance *ProxyCacheHandler) {
	if len(origins) == 0 {
		// TODO: error handling ...
		//origins = []string{"https://s2.mangadex.org", "https://uploads.mangadex.org"}
	}
	return &ProxyCacheHandler{
		origins:   origins,
		validator: validator,
	}
}

func (instance *ProxyCacheHandler) ServeHTTP(destination http.ResponseWriter, request *http.Request) {
	path, _, err := instance.validator.ExtractValidatedPath(request)
	if err != nil {
		log.Println("[VERBOSE]", "Request (Blocked):", request.RemoteAddr, "=>", request.Host+request.URL.Path)
		destination.WriteHeader(http.StatusForbidden)
		return
	} else {
		log.Println("[VERBOSE]", "Request (Accepted):", request.RemoteAddr, "=>", request.Host+request.URL.Path)
	}

	// TODO: get origin from list (random, round robin, ...)
	url := instance.origins[0] + path
	source, err := http.Get(url)
	if err != nil {
		log.Println("[WARN]", "Failed to receive image from upstream server", err)
		destination.WriteHeader(http.StatusBadGateway)
		return
	}
	defer source.Body.Close()

	for key, values := range source.Header {
		for _, value := range values {
			destination.Header().Add(key, value)
		}
	}
	destination.WriteHeader(source.StatusCode)
	io.Copy(destination, source.Body)
	log.Println("[VERBOSE]", "Response (Proxied):", request.RemoteAddr, "<=", url)
}
