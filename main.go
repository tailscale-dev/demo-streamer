package main

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"

	"github.com/google/uuid"
)

//go:embed ui/*
var static embed.FS

func main() {
	uiAssets, _ := fs.Sub(static, "ui")
	http.Handle("/", http.FileServer(http.FS(uiAssets)))

	http.HandleFunc("/api/uuid", func(w http.ResponseWriter, r *http.Request) {
		uuid := uuid.New().String()
		// write to http response
		fmt.Fprintf(w, "%s\n", uuid)
		// write to stdout
		fmt.Printf("%s\n", uuid)
	})

	port := "80"
	if portEnv := os.Getenv("PORT"); portEnv != "" {
		port = portEnv
	}

	fmt.Printf("Server running: http://localhost:%s/\n", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%s", port), nil); err != nil {
		log.Fatal(err)
	}
}
