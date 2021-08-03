package handlers

import (
	"io"
	"io/fs"
	mdath "mdath/lib"
	"mdath/log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
)

type FileCacheHandler struct {
	directory string
	upstream  *string
	validator *mdath.RequestValidator
}

func CreateFileCacheHandler(directory string, upstream *string, validator *mdath.RequestValidator) (instance *FileCacheHandler) {
	return &FileCacheHandler{
		directory: directory,
		upstream:  upstream,
		validator: validator,
	}
}

func (instance *FileCacheHandler) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	path, file, err := instance.validator.ExtractValidatedPath(request)
	if err != nil {
		log.Verbose("Request (Blocked):", request.RemoteAddr, "=>", request.Host+request.URL.Path, err)
		response.WriteHeader(http.StatusForbidden)
		return
	} else {
		log.Verbose("Request (Accepted):", request.RemoteAddr, "=>", request.Host+request.URL.Path)
	}

	file = filepath.Join(instance.directory, file[0:2], file[2:4], file[56:])
	_, err = os.Stat(file)
	if err == nil {
		serveFileFromCache(file, response, request)
		log.Verbose("Response (Cache HIT):", request.RemoteAddr, "<=", file)
	} else if os.IsNotExist(err) {
		url := *instance.upstream + path
		cacheFileFromUpstream(url, file, response, request)
		log.Verbose("Response (Cache MISS):", request.RemoteAddr, "<=", url)
	} else {
		response.WriteHeader(http.StatusInternalServerError)
		log.Warn("Failed to determine cached image status", err)
	}
}

func serveFileFromCache(file string, response http.ResponseWriter, request *http.Request) {
	filereader, info, err := openCacheImage(file)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer filereader.Close()

	response.Header().Set("Content-Type", getImageMimeType(file))
	response.Header().Set("Content-Length", strconv.FormatInt(info.Size(), 10))
	response.Header().Set("Access-Control-Allow-Origin", "*")
	response.Header().Set("Access-Control-Expose-Headers", "*")
	response.Header().Set("Cache-Control", "public, max-age=1209600")
	response.Header().Set("Timing-Allow-Origin", "*")
	response.Header().Set("X-Content-Type-Options", "nosniff")
	response.Header().Set("X-Cache", "HIT")

	response.WriteHeader(http.StatusOK)
	io.Copy(response, filereader)
}

func cacheFileFromUpstream(upstream string, file string, response http.ResponseWriter, request *http.Request) {
	source, err := http.Get(upstream)
	if err != nil {
		log.Warn("Failed to receive image from upstream server", err)
		response.WriteHeader(http.StatusBadGateway)
		return
	}
	defer source.Body.Close()

	var destination io.Writer = response
	if source.StatusCode == 200 {
		filewriter, err := createCacheImage(file)
		if err == nil {
			defer filewriter.Close()
			destination = io.MultiWriter(response, filewriter)
		}
	}

	for key, values := range source.Header {
		for _, value := range values {
			response.Header().Add(key, value)
		}
	}
	response.Header().Set("X-Cache", "MISS")

	response.WriteHeader(source.StatusCode)
	io.Copy(destination, source.Body)
}

func openCacheImage(file string) (filereader *os.File, fileinfo fs.FileInfo, err error) {
	filereader, err = os.Open(file)
	if err != nil {
		log.Warn("Failed to open cached image", err)
		return
	}
	fileinfo, err = filereader.Stat()
	if err != nil {
		filereader.Close()
		log.Warn("Failed to access info of cached image", err)
		return
	}
	return
}

func createCacheImage(file string) (filewriter *os.File, err error) {
	err = os.MkdirAll(filepath.Dir(file), 0755)
	if err != nil {
		log.Warn("Failed to create cache directory tree", err)
		return
	}
	filewriter, err = os.Create(file)
	if err != nil {
		log.Warn("Failed to create cached image", err)
		return
	}
	return
}

func getImageMimeType(file string) string {
	switch filepath.Ext(file) {
	case ".png":
		return "image/png"
	case ".jpg":
		return "image/jpeg"
	case ".jpeg":
		return "image/jpeg"
	case ".gif":
		return "image/gif"
	case ".webp":
		return "image/webp"
	default:
		return "application/octet-stream"
	}
}
