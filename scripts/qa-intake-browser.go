//go:build ignore

// A loopback-only harness for real-browser acceptance against fictional data.
// Run explicitly with `go run scripts/qa-intake-browser.go`; it is not a server
// entry point or a production deployment mode.
package main

import (
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"law-oa-go/internal/config"
	"law-oa-go/internal/database"
	"law-oa-go/internal/middleware"
	"law-oa-go/internal/router"
)

func main() {
	if os.Getenv("ENVIRONMENT") != "test" || os.Getenv("QA_FIXTURE_CONFIRM") != "I_UNDERSTAND" {
		log.Fatal("requires explicit fictional test environment")
	}
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	if !strings.HasPrefix(cfg.Database.Host, "/private/tmp/lawoa-intake-collaboration-runtime/") || cfg.Database.Database != "lawoa_intake_mvp_test" {
		log.Fatal("only the isolated intake acceptance database is allowed")
	}
	db, err := database.InitWithConfig(cfg)
	if err != nil {
		log.Fatal(err)
	}
	middleware.InitJWT(cfg)
	gin.SetMode(gin.TestMode)
	app := gin.New()
	app.Use(gin.Recovery())
	router.Init(app, db, nil)
	app.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "fixture": true, "production_ready": false})
	})
	log.Println("fictional intake acceptance server ready on 127.0.0.1:18587")
	log.Fatal(http.ListenAndServe("127.0.0.1:18587", app))
}
