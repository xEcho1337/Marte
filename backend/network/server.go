package network

import (
	"backend/api"
	"backend/auth"
	"backend/config"
	"backend/database"
	"backend/middleware"
	"backend/submitter"
	"backend/web"
	"fmt"
	"net/http"
	"shared/golog"
)

var log = golog.New("HTTP")

func Start() {
	log.Info("Loading config...")

	err := config.Ensure("config.yml")
	if err != nil {
		log.Fatal(err.Error())
	}

	cfg, err := config.Load("config.yml")
	if err != nil {
		log.Fatal(err.Error())
	}

	if cfg.TeamIpFile != "" {
		ips, err := config.LoadIPs(cfg.TeamIpFile)
		if err != nil {
			log.Fatal(fmt.Sprintf("Failed to load IPs from %s: %v", cfg.TeamIpFile, err))
		}
		cfg.TeamIPs = ips
		log.Info("Loaded %d IPs from %s", len(ips), cfg.TeamIpFile)
	}

	db := &database.Manager{}
	db.LoadDatabase()
	db.CreateTable()

	authManager := &auth.AuthManager{
		Password:        cfg.Password,
		Sessions:        make(map[string]string),
		UsernameSession: make(map[string]string),
	}
	flagManager := submitter.NewFlagManager(1024, db)

	if records, err := db.LoadAllFlags(); err != nil {
		log.Errorf("Failed to load flags from database: %v", err)
	} else if len(records) > 0 {
		flagManager.LoadCache(records)
		log.Info("Restored %d flags from database", len(records))
	}

	server := &middleware.Server{
		AuthManager: authManager,
		FlagManager: flagManager,
		Config:      cfg,
	}

	mux := http.NewServeMux()
	renderer := web.NewRenderer("../frontend/templates")

	registerEndpoints(server, mux)

	mux.HandleFunc("GET /login", web.LoginHandler(renderer))
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		http.Redirect(w, r, "/login", http.StatusFound)
	})

	mux.HandleFunc("GET /dashboard", web.DashboardHandler(renderer))
	mux.HandleFunc("GET /dashboard/grafici", web.GraficiHandler(renderer))
	mux.HandleFunc("GET /dashboard/flags", web.FlagsHandler(renderer))
	mux.HandleFunc("GET /dashboard/docs", web.DocsHandler(renderer))
	mux.HandleFunc("GET /dashboard/top", web.TopHandler(renderer))

	fs := http.FileServer(http.Dir("../frontend"))
	mux.Handle("GET /static/", http.StripPrefix("/static/", fs))

	go func() {
		log.Info("Binding to port %d", cfg.ApiPort)
		err := http.ListenAndServe(fmt.Sprintf(":%d", cfg.ApiPort), mux)

		if err != nil {
			log.Fatal(err.Error())
		}
	}()

	submitter.Start(cfg, flagManager, authManager)

	flagManager.Start(cfg.Submitter, cfg.UrlFlagSubmit, cfg.TeamToken, cfg.FlagBatch, cfg.FlagSubmitRate)

	log.Info("All servers started")
	select {}
}

func registerEndpoints(server *middleware.Server, mux *http.ServeMux) {
	mux.HandleFunc("POST /api/login", chain(
		api.Login(server),
		middleware.CreateRateLimiter(1, 1).RateLimit(server),
	))
	mux.HandleFunc("GET /api/get_services", chain(
		api.GetService(server),
		middleware.RequireAuth(server),
	))
	mux.HandleFunc("GET /api/profile", chain(
		api.Profile(server),
		middleware.RequireAuth(server),
	))
	mux.HandleFunc("GET /api/farm", chain(
		api.Farm(server),
		middleware.RequireAuth(server),
	))
	mux.HandleFunc("GET /api/flags", chain(
		api.Flags(server),
		middleware.RequireAuth(server),
	))
	mux.HandleFunc("GET /api/top", chain(
		api.Top(server),
		middleware.RequireAuth(server),
	))
	mux.HandleFunc("GET /api/stats", chain(
		api.Stats(server),
		middleware.RequireAuth(server),
	))
}

func chain(h http.HandlerFunc, mws ...middleware.Middleware) http.HandlerFunc {
	for i := len(mws) - 1; i >= 0; i-- {
		h = mws[i](h)
	}
	return h
}
