package mdath

import (
	"log"
	"net"
	"net/http"
	"runtime"
	"strconv"
	"sync/atomic"
	"time"
)

type ImageServer struct {
	server      *http.Server
	handler     http.Handler
	tlsProvider *TLSProvider
	connections int64
}

func CreateImageServer(TLSProvider *TLSProvider, handler http.Handler) *ImageServer {
	return &ImageServer{
		tlsProvider: TLSProvider,
		handler:     handler,
	}
}

func (instance *ImageServer) updateConnectionCount(conn net.Conn, state http.ConnState) {
	if state == http.StateNew {
		atomic.AddInt64(&instance.connections, 1)
	}
	if state == http.StateClosed || state == http.StateHijacked {
		atomic.AddInt64(&instance.connections, -1)
	}
	log.Println("[DEBUG]", "State:", state, ", Open Connections:", atomic.LoadInt64(&instance.connections))
}

func (instance *ImageServer) Start(port int, workers int, nossl bool) (err error) {
	if instance.server != nil {
		return
	}

	runtime.GOMAXPROCS(workers)
	instance.server = &http.Server{
		Addr:      ":" + strconv.Itoa(port),
		ConnState: instance.updateConnectionCount,
		Handler:   instance.handler,
		//ErrorLog:     logger,
		ReadHeaderTimeout: 15 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      1 * time.Minute,
		IdleTimeout:       1 * time.Minute,
	}

	var listener net.Listener
	if nossl {
		listener, err = net.Listen("tcp4", instance.server.Addr)
	} else {
		listener, err = instance.tlsProvider.CreateListener("tcp4", instance.server.Addr)
	}
	if err != nil {
		log.Println("[ERROR]", "Failed to start Image Server", err)
		return
	}
	go func() {
		err = instance.server.Serve(listener)
	}()
	time.Sleep(2500 * time.Millisecond)
	if err == nil {
		log.Println("[INFO]", "Started the Image Cache Server on", port)
	}
	return
}

func (instance *ImageServer) Stop(timeout time.Duration, interval time.Duration) (err error) {
	if instance.server == nil {
		return
	}
	instance.server.SetKeepAlivesEnabled(false)
	for remaining := timeout; remaining > 0; remaining -= interval {
		log.Println("[INFO]", "Waiting for", instance.connections, "connection(s) before stopping the Image Cache Server in", remaining)
		time.Sleep(interval)
		if instance.connections == 0 {
			log.Println("[INFO]", "No open connection(s), stopping the Image Cache Server now")
			remaining = 0
		}
	}
	err = instance.server.Close()
	if err != nil {
		log.Println("[ERROR]", "Failed to stop the Image Cache Server")
		return
	}
	instance.server = nil
	log.Println("[INFO]", "Stopped the Image Cache Server")
	return
}
