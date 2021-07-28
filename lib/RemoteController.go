package mdath

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

const (
	BuildVersion       int    = 31
	UserAgent          string = "" // "Mozilla/5.0 (System; OS) MangaDex@Home/2.x.x (JSON) cheetah/31.0"
	ApiServerURL       string = "https://api.mangadex.network"
	DefaultUpstreamURL string = "https://uploads.mangadex.org"
	MinCacheSize       int64  = 64_424_509_440        // 60 GB
	DefaultCacheSize   int64  = 1_125_899_906_842_624 // 1 PB
	KeepAliveInterval         = 1 * time.Minute
)

type PingRequestPayload struct {
	ClientSecret            string `json:"secret"`
	ImageServerPort         int    `json:"port"`
	ImageServerAddress      string `json:"ip_address,omitempty"`
	CacheSizeLimit          int64  `json:"disk_space"`    // in Bytes, must be at least 60 * 1024^3
	NetworkSpeed            int    `json:"network_speed"` // in Bytes/sec, use 0 for unmetered (use server side maximum)
	BuildVersion            int    `json:"build_version"`
	CertificateCreationDate string `json:"tls_created_at,omitempty"`
}

type PingResponsePayload struct {
	ClientID                     string   `json:"client_id"`
	ClientURL                    string   `json:"url"`
	Paused                       bool     `json:"paused"`
	Compromised                  bool     `json:"compromised"`
	LatestBuildVersion           int      `json:"latest_build"`
	UpstreamServer               string   `json:"image_server"`
	ExpirationTokenDecryptionKey string   `json:"token_key"`
	ExpirationTokenDisabled      bool     `json:"disable_tokens"`
	TLS                          *TLSInfo `json:"tls,omitempty"`
}

type TLSInfo struct {
	CreationDate string `json:"created_at"`
	PrivateKey   string `json:"private_key"`
	Certificate  string `json:"certificate"`
}

type StopRequestPayload struct {
	ClientSecret string `json:"secret"`
}

type StopResponsePayload struct {
}

type RemoteController struct {
	connected        bool
	config           PingRequestPayload
	upstream         string
	tlsProvider      *TLSProvider
	requestValidator *RequestValidator
}

// Instantiate a new RemoteController for interacting with the MangaDex@Home Remote API server.
// Must provide the API key for authentication with the MangaDex@Home Remote API server (aka MangaDex@Home client secret/key).
// Must provide the port on which this MangaDex@Home client is hosting/caching the images.
// Optionally provide the ip address on which this MangaDex@Home client is hosting/caching the images (leave blank to use the public IP of this host).
// Optionally provide the maximum cache size in bytes that shall be reported to the MangaDex@Home Remote API server (if set to default: 0, unlimited will be used).
// Optionally provide the maximum network speed that shall be reported to the MangaDex@Home Remote API server (if set to default: 0, unlimited will be used).
func CreateRemoteController(key string, ip string, port int, cache int64, speed int) (instance *RemoteController) {
	if cache == 0 {
		cache = DefaultCacheSize
	}
	if cache < MinCacheSize {
		cache = MinCacheSize
	}
	instance = &RemoteController{
		connected: false,
		config: PingRequestPayload{
			ClientSecret:            key,
			ImageServerPort:         port,
			ImageServerAddress:      ip,
			CacheSizeLimit:          cache,
			NetworkSpeed:            speed,
			BuildVersion:            BuildVersion,
			CertificateCreationDate: "",
		},
		upstream:         DefaultUpstreamURL,
		tlsProvider:      new(TLSProvider),
		requestValidator: new(RequestValidator),
	}
	go func() {
		for range time.Tick(KeepAliveInterval) {
			if instance.connected {
				instance.ping()
			}
		}
	}()
	return
}

func (instance *RemoteController) ping() (err error) {
	payload := instance.config
	data := new(PingResponsePayload)
	err = post("/ping", payload, data)
	if err != nil {
		return
	}
	instance.upstream = data.UpstreamServer
	if data.TLS != nil {
		instance.config.CertificateCreationDate = data.TLS.CreationDate
		instance.tlsProvider.Update(data.TLS)
	}
	instance.requestValidator.Update(data.ExpirationTokenDisabled, data.ExpirationTokenDecryptionKey)
	log.Println("[INFO]", "MangaDex@Home Remote API Server:")
	log.Println("     >", "Client =", data.ClientID, ", Version =", BuildVersion, "/", data.LatestBuildVersion, ", Compromised =", data.Compromised, ", Paused =", data.Paused, ", Cert-Included =", data.TLS != nil)
	log.Println("     >", "Validate-Token =", !data.ExpirationTokenDisabled, ", Token-Key =", data.ExpirationTokenDecryptionKey)
	log.Println("     >", "Public-URL =", data.ClientURL)
	log.Println("     >", "Upstream-URL =", data.UpstreamServer)
	return
}

// Open a connection to the MangaDex@Home Remote API server to keep-alive and exchange client information periodically.
func (instance *RemoteController) Connect() (upstreamServer *string, tlsProvider *TLSProvider, requestValidator *RequestValidator, err error) {
	if instance.connected {
		return
	}
	instance.config.CertificateCreationDate = ""
	err = instance.ping()
	if err != nil {
		log.Println("[ERROR]", "Failed to connected to MangaDex@Home Remote API Server", err)
		return
	}
	upstreamServer = &instance.upstream
	tlsProvider = instance.tlsProvider
	requestValidator = instance.requestValidator
	instance.connected = true
	log.Println("[INFO]", "Connected to MangaDex@Home Remote API Server")
	return
}

func (instance *RemoteController) Disconnect() (err error) {
	if !instance.connected {
		return
	}
	payload := &StopRequestPayload{
		ClientSecret: instance.config.ClientSecret,
	}
	data := new(StopResponsePayload)
	err = post("/stop", payload, data)
	if err != nil {
		log.Println("[ERROR]", "Failed to disconnect from MangaDex@Home Remote API Server", err)
		return
	}
	instance.connected = false
	log.Println("[INFO]", "Disconnect from MangaDex@Home Remote API Server")
	return
}

func post(endpoint string, payload interface{}, data interface{}) (err error) {
	buffer := new(bytes.Buffer)
	err = json.NewEncoder(buffer).Encode(payload)
	if err != nil {
		return
	}

	request, err := http.NewRequest("POST", ApiServerURL+endpoint, buffer)
	if err != nil {
		return
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("User-Agent", UserAgent)
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		err = fmt.Errorf("request to '%s' responded with status %d", ApiServerURL+endpoint, response.StatusCode)
		return
	}
	err = json.NewDecoder(response.Body).Decode(data)
	return
}
