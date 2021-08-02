package main

import (
	"flag"
	"fmt"
	"log"
	mdath "mdath/lib"
	"mdath/lib/handlers"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"
)

const (
	GigaByte                             = 1073741824
	GracefulShutdownPeriod               = 30 * time.Second
	GracefulShutdownNotificationInterval = 5 * time.Second
)

var (
	key             string
	ip              string
	port            int
	noTokenCheck    bool
	upstreamServer  string
	upstreamServers []string
	cacheDirectory  string
	cacheSize       int64
	loglevel        int
)

func main() {
	if len(os.Args) < 2 {
		// TODO: print help?
		log.Println("[ERROR]", "Missing commandline arguments")
		os.Exit(1)
	}
	switch os.Args[1] {
	case "proxy":
		startClusterProxy()
	case "cache":
		startClusterCache()
	default:
		startStandAlone()
	}
}

func run() {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	<-signals
	close(signals)
	fmt.Println()
}

func startStandAlone() {
	cmd := flag.NewFlagSet("", flag.ExitOnError)
	cmd.StringVar(&key, "key", "", "Client secret required to connect to the MangaDex@Home Remote API Server.")
	cmd.StringVar(&ip, "ip", "", "...")
	cmd.IntVar(&port, "port", 443, "Port on which the client will listen to incoming requests and serve the cached images.")
	cmd.BoolVar(&noTokenCheck, "no-token-check", false, "Disable token verification ...")
	cmd.StringVar(&cacheDirectory, "cache", "./cache", "Directory where images are cached.")
	cmd.Int64Var(&cacheSize, "size", 256, "Max. cache size (in GB) which is reported to the MangaDex@Home Remote API Server (used for shard assignment).")
	flag.IntVar(&loglevel, "loglevel", 2, "The log level ...")

	cmd.Parse(os.Args[1:])

	remote := mdath.CreateRemoteController(key, ip, port, cacheSize*GigaByte, 0)
	upstream, tls, validator, err := remote.Connect()
	if err != nil {
		log.Panic(err)
	}

	if noTokenCheck {
		validator = new(mdath.RequestValidator)
		validator.Update(true, "")
	}

	server := mdath.CreateImageServer(tls, handlers.CreateFileCacheHandler(cacheDirectory, upstream, validator))
	err = server.Start(port, runtime.NumCPU(), false)
	if err != nil {
		log.Panic(err)
	}

	run()

	err = remote.Disconnect()
	if err != nil {
		os.Exit(1)
	}
	err = server.Stop(GracefulShutdownPeriod, GracefulShutdownNotificationInterval)
	if err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}

func startClusterProxy() {
	cmd := flag.NewFlagSet("proxy", flag.ExitOnError)
	cmd.StringVar(&key, "key", "", "Client secret required to connect to the MangaDex@Home Remote API Server.")
	cmd.StringVar(&ip, "ip", "", "...")
	cmd.IntVar(&port, "port", 443, "The port on which the client will listen to incoming requests and serve the cached images.")
	cmd.BoolVar(&noTokenCheck, "no-token-check", false, "Disable token verification ...")
	cmd.StringVar(&upstreamServer, "origins", "https://uploads.mangadex.org", "Comma separated list of ...")
	flag.IntVar(&loglevel, "loglevel", 2, "The log level ...")

	cmd.Parse(os.Args[2:])

	// TODO: introduce new type for flag that parses []string
	upstreamServers = strings.Split(upstreamServer, ",")

	remote := mdath.CreateRemoteController(key, ip, port, 0*GigaByte, 0)
	_, tls, validator, err := remote.Connect()
	if err != nil {
		log.Panic(err)
	}

	if noTokenCheck {
		validator = new(mdath.RequestValidator)
		validator.Update(true, "")
	}

	server := mdath.CreateImageServer(tls, handlers.CreateProxyCacheHandler(upstreamServers, validator))
	err = server.Start(port, runtime.NumCPU(), false)
	if err != nil {
		log.Panic(err)
	}

	run()

	err = remote.Disconnect()
	if err != nil {
		os.Exit(1)
	}
	err = server.Stop(GracefulShutdownPeriod, GracefulShutdownNotificationInterval)
	if err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}

func startClusterCache() {
	cmd := flag.NewFlagSet("cache", flag.ExitOnError)
	cmd.IntVar(&port, "port", 80, "Port on which the client will listen to incoming requests and serve the cached images.")
	cmd.StringVar(&upstreamServer, "upstream", "https://uploads.mangadex.org", "...")
	cmd.StringVar(&cacheDirectory, "cache", "./cache", "")
	cmd.Int64Var(&cacheSize, "size", 256, "The max. size (in GB) used for cached images.")
	flag.IntVar(&loglevel, "loglevel", 2, "The log level ...")

	cmd.Parse(os.Args[2:])

	tls := new(mdath.TLSProvider)
	validator := new(mdath.RequestValidator)
	validator.Update(true, "")

	server := mdath.CreateImageServer(tls, handlers.CreateFileCacheHandler(cacheDirectory, &upstreamServer, validator))
	err := server.Start(port, runtime.NumCPU(), true)
	if err != nil {
		log.Panic(err)
	}

	run()

	err = server.Stop(GracefulShutdownPeriod, GracefulShutdownNotificationInterval)
	if err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}
