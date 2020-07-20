package main

import (
	"flag"
	"log"

	"github.com/NewGlad/xsolla-be/internal/app/newsapi"
	"github.com/gorilla/sessions"
)

var (
	configPath string
)

func init() {
	flag.StringVar(&configPath, "config-path", "configs/newsapi.yaml", "Path to the API Config yaml")
}

func main() {
	flag.Parse()
	config, err := newsapi.NewConfig(configPath)
	if err != nil {
		log.Fatal(err)
	}
	sessionsStore := sessions.NewCookieStore([]byte(config.SessionKey))
	server := newsapi.New(config, sessionsStore)
	if err := server.Start(); err != nil {
		log.Fatal(err)
	}
}
