package main

import (
	"fmt"

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
}
