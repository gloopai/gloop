package main

import (
	"fmt"
	"net/http"

	"github.com/gloopai/gloop/site"
)

func main() {
	// Example usage
	config := site.SiteConfig{
		Port:     8080,
		TLSCert:  "path/to/cert.pem",
		TLSKey:   "path/to/key.pem",
		UseHTTPS: false,
		BaseRoot: "static",
	}

	domains := []string{"localhost:8080"}

	site := site.NewSite(config, domains)
	if err := site.Start(); err != nil {
		fmt.Println("Error starting site:", err)
	}

	site.AddRoute("/example", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello, this is an example route!")

	})

	// Register a new route to output a JSON response
	site.AddRoute("/json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message": "This is a JSON response"}`))
	})

	for {
	}
}
