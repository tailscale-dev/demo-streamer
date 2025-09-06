package main

import (
	"context"
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/alecthomas/kong"
	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"tailscale.com/client/local"
	"tailscale.com/tsnet"
)

//go:embed ui/*
var ui embed.FS

// Version information
var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
)

type CLI struct {
	Port     string `kong:"default='8080',env='PORT',help='The port to listen on'"`
	Dev      bool   `kong:"env='DEV',help='Enable dev mode'"`
	Tsnet    bool   `kong:"env='TSNET',help='Use tsnet for Tailscale integration'"`
	Hostname string `kong:"env='HOSTNAME',help='Hostname for tsnet registration'"`
	AuthKey  string `kong:"env='TAILSCALE_AUTHKEY',help='Tailscale auth key for tsnet'"`
	TLS      bool   `kong:"default='true',env='TLS',help='Enable TLS certificate generation'"`
	Version  bool   `kong:"help='Show version information'"`
}

var latencyHistogram = promauto.NewHistogram(prometheus.HistogramOpts{
	Name:    "tailscale_whois_latency_milliseconds",
	Help:    "The latency of Tailscale WhoIs requests in milliseconds",
	Buckets: prometheus.DefBuckets,
})

func main() {
	var cli CLI
	ctx := kong.Parse(&cli)

	// Handle version command
	if cli.Version {
		buildInfo := getBuildInfo()
		fmt.Printf("tailscale-demo-streamer %s\n", buildInfo)
		fmt.Printf("Built on: %s\n", date)
		return
	}

	var droppedCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "dropped_total",
		Help: "The total number of dropped objects",
	})

	// Setup server based on whether we're using tsnet or regular HTTP
	var server *http.Server
	var listener interface {
		Close() error
	}

	if cli.Tsnet {
		server, listener = setupTsnetServer(&cli)
	} else {
		server, listener = setupRegularServer(&cli)
	}
	defer listener.Close()

	setupRoutes(&cli, droppedCounter)

	buildInfo := getBuildInfo()
	fmt.Printf("Starting tailscale-demo-streamer %s on port %s (tsnet: %v)\n", buildInfo, cli.Port, cli.Tsnet)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		ctx.Fatalf("Server failed: %v", err)
	}
}

func setupTsnetServer(cli *CLI) (*http.Server, *tsnet.Server) {
	hostname := cli.Hostname
	if hostname == "" {
		hostname = "tailscale-demo-streamer"
	}

	ts := &tsnet.Server{
		Hostname: hostname,
		Logf:     log.Printf,
	}

	if cli.AuthKey != "" {
		ts.AuthKey = cli.AuthKey
	}

	// Start the tsnet server
	err := ts.Start()
	if err != nil {
		log.Fatalf("Failed to start tsnet server: %v", err)
	}

	var server *http.Server

	if cli.TLS {
		// Use HTTPS with Tailscale's automatic certificate management
		ln, err := ts.ListenTLS("tcp", ":https")
		if err != nil {
			log.Fatalf("Failed to listen on HTTPS: %v", err)
		}

		server = &http.Server{
			Addr: ln.Addr().String(),
		}

		go func() {
			if err := server.Serve(ln); err != nil && err != http.ErrServerClosed {
				log.Printf("HTTPS server error: %v", err)
			}
		}()
	} else {
		// Listen on HTTP
		ln, err := ts.Listen("tcp", ":80")
		if err != nil {
			log.Fatalf("Failed to listen on HTTP: %v", err)
		}

		server = &http.Server{
			Addr: ln.Addr().String(),
		}

		go func() {
			if err := server.Serve(ln); err != nil && err != http.ErrServerClosed {
				log.Printf("HTTP server error: %v", err)
			}
		}()
	}

	return server, ts
}

func setupRegularServer(cli *CLI) (*http.Server, *http.Server) {
	server := &http.Server{
		Addr: fmt.Sprintf("0.0.0.0:%s", cli.Port),
	}
	return server, server
}

func setupRoutes(cli *CLI, droppedCounter prometheus.Counter) {
	var templateFn func() *template.Template
	if cli.Dev {
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
		if cli.Dev {
			fmt.Printf("%s\n", uuid) // write to stdout - TODO: maybe only in dev mode?
		}

		droppedCounter.Inc() // Increment the counter
	})

	http.HandleFunc("/api/user", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		whois, err := tailscaleWhois(r.Context(), r, cli.Tsnet)
		if err != nil {
			// Return empty user info on error
			fmt.Fprintf(w, `{"connected": false, "error": "%s"}`, err.Error())
			return
		}

		if whois != nil {
			var firstInitial string
			if whois.DisplayName != "" {
				firstInitial = string(whois.DisplayName[0])
			} else {
				firstInitial = string(whois.LoginName[0])
			}

			fmt.Fprintf(w, `{"connected": true, "loginName": "%s", "displayName": "%s", "firstInitial": "%s"}`,
				whois.LoginName, whois.DisplayName, firstInitial)
		} else {
			fmt.Fprintf(w, `{"connected": false}`)
		}
	})

	http.Handle("/metrics", promhttp.Handler())

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		var pageData *page
		whois, err := tailscaleWhois(r.Context(), r, cli.Tsnet)
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

		fmt.Printf("user info [%+v] for [%+v] \n", *pageData, r.RemoteAddr)

		err = templateFn().Execute(w, pageData)
		if err != nil {
			fmt.Printf("error rendering template [%+v] \n", err)
			// TODO: re-render template with nil pageData to not interrupt demo?
		}
	})
}

type page struct {
	UserProfile  *whoisData
	FirstInitial string
}

type whoisData struct {
	LoginName   string
	DisplayName string
}

func tailscaleWhois(ctx context.Context, r *http.Request, useTsnet bool) (*whoisData, error) {
	var u *whoisData

	localClient := &local.Client{}
	start := time.Now() // Start measuring latency

	defer func() {
		latency := float64(time.Since(start)) / float64(time.Millisecond) // Calculate latency in milliseconds
		latencyHistogram.Observe(latency)                                 // Record latency in the histogram
	}()

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

// getBuildInfo returns version information from build context
func getBuildInfo() string {
	if version != "dev" {
		return fmt.Sprintf("v%s-%s", version, commit)
	}

	if info, ok := debug.ReadBuildInfo(); ok {
		if info.Main.Version != "(devel)" && info.Main.Version != "" {
			return info.Main.Version
		}
		// Try to get version from VCS info
		for _, setting := range info.Settings {
			if setting.Key == "vcs.revision" && len(setting.Value) >= 7 {
				return fmt.Sprintf("dev-%s", setting.Value[:7])
			}
		}
	}

	return "dev-unknown"
}
