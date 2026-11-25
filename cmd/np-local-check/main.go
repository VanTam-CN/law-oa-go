//go:build np_local_check

// Serve only the synthetic browser fixture. Seed and inspect its rows separately.
package main

import (
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"law-oa-go/internal/config"
	"law-oa-go/internal/database"
	"law-oa-go/internal/middleware"
	"law-oa-go/internal/router"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	dbConfig := cfg.Database
	if os.Getenv("NP_LOCAL_CHECK") != "synthetic-only" || cfg.IsProduction() ||
		dbConfig.Driver != "postgres" || dbConfig.Host != "127.0.0.1" || dbConfig.Port != "18790" ||
		dbConfig.Database != "np_first_slice_dev" || dbConfig.Username != "np_m1" {
		log.Fatal("isolated synthetic PostgreSQL fixture required")
	}
	var seed struct {
		Actor string `json:"actor_id"`
	}
	manifest, err := os.ReadFile(os.Getenv("NP_BROWSER_SEED"))
	if err != nil || json.Unmarshal(manifest, &seed) != nil {
		log.Fatal("browser seed manifest required")
	}
	actor, err := strconv.ParseUint(seed.Actor, 10, 64)
	if err != nil || actor == 0 {
		log.Fatal("invalid synthetic actor")
	}
	if err := os.Setenv("NP_ASSISTANT_ACTORS", strconv.FormatUint(actor, 10)); err != nil {
		log.Fatal(err)
	}
	db, err := database.InitWithConfig(cfg)
	if err != nil {
		log.Fatal(err)
	}
	middleware.InitJWT(cfg)
	gin.SetMode(gin.ReleaseMode)
	app := gin.New()
	app.Use(middleware.LoggerWithFormatter(), middleware.Recovery())
	if err := router.Init(app, db, nil); err != nil {
		log.Fatal(err)
	}
	srv := &http.Server{Addr: "127.0.0.1:18789", Handler: dropAfterCommit(app), ReadHeaderTimeout: 5 * time.Second}
	log.Fatal(srv.ListenAndServe())
}

func dropAfterCommit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut || r.Header.Get("X-NP-Drop-Response") != "after-commit" {
			next.ServeHTTP(w, r)
			return
		}
		record := httptest.NewRecorder()
		next.ServeHTTP(record, r)
		if record.Code == http.StatusOK {
			if conn, _, err := http.NewResponseController(w).Hijack(); err == nil {
				_ = conn.Close()
				return
			}
		}
		for k, values := range record.Header() {
			w.Header()[k] = values
		}
		w.WriteHeader(record.Code)
		_, _ = w.Write(record.Body.Bytes())
	})
}
