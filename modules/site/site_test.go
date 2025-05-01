package site

import (
	"fmt"
	"net/http"
)

// //go:embed static/*
// var manageStatic embed.FS

func StartTest() {
	// Example usage
	config := SiteOptions{
		Port:     8080,
		TLSCert:  "path/to/cert.pem",
		TLSKey:   "path/to/key.pem",
		UseHTTPS: false,
		BaseRoot: "static",
		UseEmbed: true,
		// EmbedFiles:     manageStatic,
		ForceIndexHTML: true,
		// JWTOptions: auth.JWTOptions{
		// 	SecretKey:     "RxyiJcD8O19/GE9GL/V2sn0b/MOSWTWoygN77e7RNSI=",
		// 	TokenDuration: 24 * 365, // 365 days
		// },
	}

	mysite := NewSite(config)
	if err := mysite.Start(); err != nil {
		fmt.Println("Error starting site:", err)
	}

	mysite.AddRoute("/example", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello, this is an example route!")

	})

	// Register a new route to output a JSON response
	mysite.AddRoute("/json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message": "This is a JSON response"}`))
	})
}
