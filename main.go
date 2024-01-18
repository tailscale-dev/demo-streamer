package main

import (
	"context"
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"os"

	"github.com/google/uuid"
	"tailscale.com/client/tailscale"
	"tailscale.com/client/tailscale/apitype"
	"tailscale.com/tailcfg"
)

//go:embed static/*
var static embed.FS

func main() {
	if devMode := os.Getenv("DEV"); devMode == "" {
		uiAssets, _ := fs.Sub(static, "static")
		http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(uiAssets))))
	} else {
		http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		serveTemplate("./ui/index.html", getPageData(r.Context(), r.RemoteAddr), w, r)
	})

	http.HandleFunc("/api/uuid", func(w http.ResponseWriter, r *http.Request) {
		uuid := uuid.New().String()
		fmt.Fprintf(w, "%s\n", uuid) // write to http response
		fmt.Printf("%s\n", uuid)     // write to stdout
	})

	port := "8080"
	if portEnv := os.Getenv("PORT"); portEnv != "" {
		port = portEnv
	}

	fmt.Printf("Starting server: http://localhost:%s/\n", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%s", port), nil); err != nil {
		log.Fatal(err)
	}
}

func serveTemplate(templatePath string, p *Page, w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles(templatePath)
	t.Execute(w, p)
}

type Page struct {
	UserProfile *tailcfg.UserProfile
}

func getPageData(ctx context.Context, remoteAddr string) *Page {
	whois, err := whois(ctx, remoteAddr)
	if err != nil {
		return &Page{UserProfile: nil}
	}
	return &Page{UserProfile: whois.UserProfile}
}

func whois(ctx context.Context, remoteAddr string) (*apitype.WhoIsResponse, error) {
	localClient := &tailscale.LocalClient{}
	whois, err := localClient.WhoIs(ctx, remoteAddr)

	if err != nil {
		return nil, fmt.Errorf("failed to identify remote host: %w", err)
	}
	if whois.Node.IsTagged() {
		return nil, fmt.Errorf("tagged nodes do not have a user identity")
	}
	if whois.UserProfile == nil || whois.UserProfile.LoginName == "" {
		return nil, fmt.Errorf("failed to identify remote user")
	}

	return whois, nil
}
