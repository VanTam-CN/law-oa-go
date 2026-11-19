//go:build np_local_check

// This synthetic fixture runner is excluded from normal application builds.
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"law-oa-go/internal/config"
	"law-oa-go/internal/database"
	"law-oa-go/internal/middleware"
	"law-oa-go/internal/models"
	"law-oa-go/internal/router"
)

func main() {
	if os.Getenv("NP_LOCAL_CHECK") != "synthetic-only" {
		log.Fatal("NP_LOCAL_CHECK=synthetic-only required")
	}
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	if cfg.IsProduction() || cfg.Database.Host != "127.0.0.1" || cfg.Database.Port != "18790" || cfg.Database.Database != "np_first_slice_dev" {
		log.Fatal("refusing non-isolated database")
	}
	password := os.Getenv("NP_FIXTURE_PASSWORD")
	if len(password) < 12 {
		log.Fatal("NP_FIXTURE_PASSWORD must be synthetic and at least 12 characters")
	}
	db, err := database.InitWithConfig(cfg)
	if err != nil {
		log.Fatal(err)
	}
	if err := seed(db, password); err != nil {
		log.Fatal(err)
	}
	os.Setenv("NP_ASSISTANT_ACTORS", "98001")
	middleware.InitJWT(cfg)
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(middleware.LoggerWithFormatter(), middleware.Recovery())
	router.Init(r, db, nil)
	srv := &http.Server{Addr: "127.0.0.1:18789", Handler: http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Method == "PUT" && req.Header.Get("X-NP-Drop-Response") == "after-commit" {
			// This build-tagged synthetic harness closes a real upstream socket only after the actual handler committed.
			recorded := httptest.NewRecorder()
			r.ServeHTTP(recorded, req)
			if recorded.Code == http.StatusOK {
				conn, _, err := w.(http.Hijacker).Hijack()
				if err == nil {
					_ = conn.Close()
					return
				}
			}
			for key, values := range recorded.Header() {
				for _, value := range values {
					w.Header().Add(key, value)
				}
			}
			w.WriteHeader(recorded.Code)
			_, _ = w.Write(recorded.Body.Bytes())
			return
		}
		r.ServeHTTP(w, req)
	}), ReadHeaderTimeout: 5 * time.Second}
	log.Print("NP synthetic fixture ready on loopback 18789; no external model configured")
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
	_ = srv.Shutdown(context.Background())
}
func seed(db *gorm.DB, password string) error {
	var count int64
	if err := db.Table("users").Where("id=?", 98001).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return db.Transaction(func(tx *gorm.DB) error {
		user := models.User{ID: 98001, Username: "np.fixture.lawyer", Name: "虚构试验律师", Email: "np-lawyer@np.invalid", Password: string(hash), Role: "lawyer", Status: "active"}
		if err := tx.Create(&user).Error; err != nil {
			return err
		}
		client := models.Client{Name: "虚构试验客户", Type: "企业", CreatedBy: user.ID, Email: "np-client@np.invalid"}
		client.ID = 98001
		if err := tx.Create(&client).Error; err != nil {
			return err
		}
		for i, id := range []string{"00000000-0000-4000-8000-000000009801", "00000000-0000-4000-8000-000000009802"} {
			if err := tx.Table("case_intakes").Create(map[string]interface{}{"id": id, "intake_code": fmt.Sprintf("NP-BROWSER-%d", i+1), "title": fmt.Sprintf("虚构接案·%s", []string{"甲", "乙"}[i]), "client_id": client.ID, "created_by": user.ID, "case_type": "civil", "description": "原摘要标记", "status": "draft", "metadata": fmt.Sprintf(`{"draft_revision":%q,"lawyer_id":"98001","np_unrelated":{"keep":true}}`, uuid.NewString())}).Error; err != nil {
				return err
			}
			if err := tx.Table("case_intake_parties").Create(map[string]interface{}{"intake_id": id, "entity_name": "虚构对方", "entity_type": "LEGAL_PERSON", "party_role": "opposing_party", "metadata": `{"keep":true}`}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
