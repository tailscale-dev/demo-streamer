package main

import (
	"context"
	"embed"
	"flag"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net/http"

	"github.com/google/uuid"
	"tailscale.com/client/tailscale"
)

//go:embed ui/*
var ui embed.FS

var (
	port = flag.String("port", "80", "the port to listen on")
	dev  = flag.Bool("dev", false, "enable dev mode")
)

func main() {
	flag.Parse()

	var templateFn func() *template.Template
	if *dev {
		// load assets from local filesystem
		http.Handle("/ui/", http.StripPrefix("/ui/", http.FileServer(http.Dir("ui"))))

		templateFn = func() *template.Template {
			t, _ := template.ParseFiles("./ui/index.html")
			return t
		}
	} else {
		// load assets from embedded filesystem
		uiAssets, _ := fs.Sub(ui, "ui")
		http.Handle("/ui/", http.StripPrefix("/ui/", http.FileServer(http.FS(uiAssets))))

		t, _ := template.ParseFS(uiAssets, "index.html")
		templateFn = func() *template.Template {
			return t
		}
	}

	http.HandleFunc("/api/uuid", func(w http.ResponseWriter, r *http.Request) {
		uuid := uuid.New().String()
		fmt.Fprintf(w, "%s\n", uuid) // write to http response
		fmt.Printf("%s\n", uuid)     // write to stdout - TODO: maybe only in dev mode?
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		var pageData *page
		whois, err := tailscaleWhois(r.Context(), r)
		if err != nil {
			pageData = &page{
				UserProfile: nil,
			}
		} else if whois != nil {
			var firstInitial string
			if whois.DisplayName != "" {
				firstInitial = string(whois.DisplayName[0])
			} else {
				firstInitial = string(whois.LoginName[0])
			}
			pageData = &page{
				UserProfile:  whois,
				FirstInitial: firstInitial,
			}
		}

		fmt.Printf("user info [%+v] for [%+v] \n", pageData, r.RemoteAddr)

		err = templateFn().Execute(w, pageData)
		if err != nil {
			fmt.Printf("error rendering template [%+v] \n", err)
			// TODO: re-render template with nil pageData to not interrupt demo?
		}
	})

	fmt.Printf("Starting server: http://localhost:%s/\n", *port)
	if err := http.ListenAndServe(fmt.Sprintf(":%s", *port), nil); err != nil {
		log.Fatal(err)
	}
}

type page struct {
	UserProfile  *whoisData
	FirstInitial string
}

type whoisData struct {
	LoginName   string
	DisplayName string
}

func tailscaleWhois(ctx context.Context, r *http.Request) (*whoisData, error) {
	var u *whoisData

	localClient := &tailscale.LocalClient{}
	whois, err := localClient.WhoIs(ctx, r.RemoteAddr)

	if err != nil {
		if r.Header.Get("Tailscale-User-Login") != "" {
			// https://tailscale.com/kb/1312/serve#identity-headers
			u = &whoisData{
				LoginName:   r.Header.Get("Tailscale-User-Login"),
				DisplayName: r.Header.Get("Tailscale-User-Name"),
			}
		} else {
			return nil, fmt.Errorf("failed to identify remote host: %w", err)
		}
	} else {
		if whois.Node.IsTagged() {
			return nil, fmt.Errorf("tagged nodes do not have a user identity")
		} else if whois.UserProfile == nil || whois.UserProfile.LoginName == "" {
			return nil, fmt.Errorf("failed to identify remote user")
		}
		u = &whoisData{
			LoginName:   whois.UserProfile.LoginName,
			DisplayName: whois.UserProfile.DisplayName,
		}
	}

	return u, nil
}
