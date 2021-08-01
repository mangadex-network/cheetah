package mdath

import (
	"crypto/tls"
	"net"
	"sync"
)

type TLSProvider struct {
	info  *TLSInfo
	cert  *tls.Certificate
	mutex sync.RWMutex
}

// Provide a HTTPS listener based on the underlying TLS configuration.
func (instance *TLSProvider) CreateListener(network string, address string) (listener net.Listener, err error) {
	config := &tls.Config{
		ClientAuth: tls.NoClientCert,
		//MinVersion:               tls.VersionTLS10,
		NextProtos:               []string{"h2", "http/1.1"},
		GetCertificate:           instance.GetCertificate,
		PreferServerCipherSuites: true,
		// If CipherSuites is nil, a default list of secure cipher suites is used, with a preference order based on hardware performance.
		/*
			CipherSuites: []uint16{
				// https://pkg.go.dev/crypto/tls#pkg-constants

				// TLS 1.0 - 1.2 (for RSA server certifacte)
				//tls.TLS_RSA_WITH_RC4_128_SHA,                    // 142.03 req/s
				//tls.TLS_RSA_WITH_3DES_EDE_CBC_SHA,               //  27.58 req/s
				//tls.TLS_RSA_WITH_AES_128_CBC_SHA,                // 168.58 req/s
				//tls.TLS_RSA_WITH_AES_256_CBC_SHA,                // 164.96 req/s
				//tls.TLS_RSA_WITH_AES_128_CBC_SHA256,             // 128.54 req/s
				//tls.TLS_RSA_WITH_AES_128_GCM_SHA256,             // 221.98 req/s
				//tls.TLS_RSA_WITH_AES_256_GCM_SHA384,             // 214.77 req/s
				//tls.TLS_ECDHE_RSA_WITH_RC4_128_SHA,              //      ? req/s
				//tls.TLS_ECDHE_RSA_WITH_3DES_EDE_CBC_SHA,         //      ? req/s
				//tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA,          //      ? req/s
				//tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,          //      ? req/s
				//tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA256,       //      ? req/s
				//tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,       // 189.06 req/s
				//tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,       // 183.72 req/s
				//tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256, // 130.55 req/s
				// TLS 1.0 - 1.2 (for ECC server certificate)
				//tls.TLS_ECDHE_ECDSA_WITH_RC4_128_SHA,              //      ? req/s
				//tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA,          //      ? req/s
				//tls.TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA,          //      ? req/s
				//tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA256,       //      ? req/s
				//tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,       //      ? req/s
				//tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,       //      ? req/s
				//tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256, //      ? req/s
				// TLS 1.3
				//tls.TLS_AES_128_GCM_SHA256,                        //      ? req/s
				//tls.TLS_AES_256_GCM_SHA384,                        //      ? req/s
				//tls.TLS_CHACHA20_POLY1305_SHA256,                  //      ? req/s
			},
		*/
	}
	listener, err = net.Listen(network, address)
	if err != nil {
		return
	}
	listener = tls.NewListener(listener, config)
	return
}

// Provide the certificate of the underlying TLS configuration used in the provided HTTPS listener.
func (instance *TLSProvider) GetCertificate(clientHello *tls.ClientHelloInfo) (*tls.Certificate, error) {
	instance.mutex.RLock()
	defer instance.mutex.RUnlock()
	return instance.cert, nil
}

// Update the certificate of the underlying TLS configuration used in the provided HTTPS listener.
func (instance *TLSProvider) Update(info *TLSInfo) {
	if instance.info != nil && instance.info.CreationDate == info.CreationDate {
		return
	}
	cert, err := tls.X509KeyPair([]byte(info.Certificate), []byte(info.PrivateKey))
	if err == nil {
		instance.mutex.Lock()
		defer instance.mutex.Unlock()
		instance.info = info
		instance.cert = &cert
	}
}
