package mdath

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"regexp"
	"time"

	"golang.org/x/crypto/nacl/box"
)

const (
	KeySize   int = 32
	NonceSize int = 24
)

var expression = regexp.MustCompile(`^\/?([^\/]*)(\/data(?:-saver)?\/[a-zA-Z0-9]{32}\/[^\/\-]+\-([a-zA-Z0-9]{64}\.[a-z]{3,4}))$`)

type Token struct {
	ClientID string    `json:"client_id"`
	Expires  time.Time `json:"expires"`
	Hash     string    `json:"hash"`
}

type RequestValidator struct {
	disabled  bool
	keyBase64 string
	keyBytes  [KeySize]byte
}

func (instance *RequestValidator) Update(disabled bool, key string) (err error) {
	instance.disabled = disabled
	if instance.keyBase64 == key {
		return
	}
	instance.keyBase64 = key
	bytes, err := base64.StdEncoding.DecodeString(key)
	if err != nil || len(bytes) != KeySize {
		return
	}
	copy(instance.keyBytes[:], bytes[:KeySize])
	return
}

// Verify that the path and the token are valid and returns the path without the token.
func (instance *RequestValidator) ExtractValidatedPath(request *http.Request) (path string, file string, err error) {
	token, path, file, err := instance.verifyPath(request.URL.Path)
	if err != nil {
		log.Println("[DEBUG]", "Failed path verification:", err)
		return
	}
	err = instance.verifyReferer(request.Referer())
	if err != nil {
		log.Println("[DEBUG]", "Failed referer verification:", err)
		return
	}
	err = instance.verifyToken(token)
	if err != nil {
		log.Println("[DEBUG]", "Failed token verification:", err)
		return
	}
	return
}

func (instance *RequestValidator) verifyReferer(referer string) error {
	return nil
}

func (instance *RequestValidator) verifyPath(path string) (token string, segment string, file string, err error) {
	segments := expression.FindStringSubmatch(path)
	if len(segments) != 4 {
		err = errors.New("invalid path pattern")
		return
	}
	token = segments[1]
	segment = segments[2]
	file = segments[3]
	return
}

func (instance *RequestValidator) verifyToken(token string) (err error) {
	if instance.disabled {
		return
	}
	bytes, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil {
		return
	}
	if len(bytes) < NonceSize {
		err = errors.New("invalid length of token")
		return
	}
	var nonce [NonceSize]byte
	copy(nonce[:], bytes[:NonceSize])
	decrypted, success := box.OpenAfterPrecomputation(nil, bytes[NonceSize:], &nonce, &instance.keyBytes)
	if !success {
		err = errors.New("decryption of token failed")
		return
	}
	data := &Token{}
	err = json.Unmarshal(decrypted, data)
	if err != nil {
		return
	}
	if time.Now().After(data.Expires) {
		err = errors.New("token expired")
		return
	}
	return
}
