package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	metrics "github.com/slok/go-http-metrics/metrics/prometheus"
)

var (
	version = "0.0.0"
)

func main() {
	var apiListen, lifecycleListen, corsDomain, auth string
	var showVersion bool

	flag.StringVar(&apiListen, "apiListen", ":80", "Listen for API requests on this host/port.")
	flag.StringVar(&lifecycleListen, "lifecycleListen", ":8888", "Listen for lifecycle requests (health, metrics) on this host/port")
	flag.StringVar(&corsDomain, "cors", "*", "The 'Access-Control-Allow-Origin' value to be returned.")
	flag.StringVar(&auth, "auth", "", "A list of comma separated user=passwords for basic auth creds")
	flag.BoolVar(&showVersion, "version", false, "Display the version")
	flag.Parse()

	if showVersion {
		handleVersionCommand()
		return
	}

	cfg := apiRouterConfig{
		corsDomain: corsDomain,
		accounts:   processAuthConfig(auth),
	}

	runServers(cfg, apiListen, lifecycleListen)
}

func runServers(cfg apiRouterConfig, apiListen string, lifecycleListen string) {
	sigChannel := make(chan os.Signal, 1)
	signal.Notify(sigChannel, syscall.SIGTERM, syscall.SIGINT)

	agg := newAggregate()

	promMetricsConfig := metrics.Config{
		Registry: promRegistry,
	}

	apiRouter := setupAPIRouter(cfg, agg, promMetricsConfig)
	go runServer("api", apiRouter, apiListen)

	lifecycleRouter := setupLifecycleRouter(promRegistry)
	go runServer("lifecycle", lifecycleRouter, lifecycleListen)

	// Block until an interrupt or term signal is sent
	<-sigChannel
}

func runServer(label string, r *gin.Engine, listen string) {
	log.Printf("%s server listening at %s", label, listen)
	if err := r.Run(listen); err != nil {
		log.Panicf("error while serving %s: %v", label, err)
	}
}

func handleVersionCommand() {
	fmt.Printf("%s\n", version)
}
