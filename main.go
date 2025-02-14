package main

import (
	"cmp"
	"log"
	"net/http"
	"os"

	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

func main() {
	proxyHandler := &Proxy{}
	port := cmp.Or(os.Getenv("PORT"), "8080")

	v2Server := &http2.Server{}

	v1Server := &http.Server{
		Addr:    ":" + port,
		Handler: h2c.NewHandler(proxyHandler, v2Server),
	}

	if err := http2.ConfigureServer(v1Server, v2Server); err != nil {
		log.Fatalf("Failed to configure http/2: %s\n", err)
	}

	log.Println("Starting proxy server on ", port)
	if err := v1Server.ListenAndServe(); err != nil {
		log.Fatalf("Could not start server: %s\n", err)
	}
}
