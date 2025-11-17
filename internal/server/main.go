package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/pseudoresonance/authserver/internal/config"
	"github.com/pseudoresonance/authserver/internal/database"
)

func main() {
	// Config
	configPathFlag := flag.String("c", "config.yaml", "path to YAML config file or directory")
	flag.Parse()
	configPathEnv, exist := os.LookupEnv("CONFIG_PATH")

	if exist && len(configPathEnv) > 0 {
		configPathFlag = &configPathEnv
	}

	config, err := config.LoadConfig(*configPathFlag)
	if err != nil {
		log.Fatalf("Error while loading config\n%v\n", err)
	}

	// Database
	db := database.DatabaseManager{}
	db.Init(config)
	defer db.Close()

	// Server
	authHandler := AuthHandler{ApiIps: config.ApiIps, MonitoringIps: config.MonitoringIpRanges, PrivateIps: config.PrivateIps, QueryTokenKey: config.QueryTokenKey, Database: &db}
	authHandler.Init()
	http.Handle("/auth", authHandler)

	connectHandler := ConnectHandler{Database: &db}
	http.Handle("/connection", connectHandler)

	forwardAuthHandler := ForwardAuthHandler{PrivateIps: config.PrivateIps, QueryTokenKey: config.QueryTokenKey, Config: config.ForwardAuth, Database: &db}
	http.Handle("/forward", forwardAuthHandler)

	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "ok")
	})

	bindAddr := fmt.Sprintf("%v:%v", config.BindAddress, config.BindPort)
	log.Printf("Starting server on %v\n", bindAddr)
	log.Fatalf("HTTP server error\n%v\n", http.ListenAndServe(bindAddr, nil))
}
