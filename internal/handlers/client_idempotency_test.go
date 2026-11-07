package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"law-oa-go/internal/common"
	"law-oa-go/internal/models"
	"law-oa-go/internal/repositories"
	"law-oa-go/internal/services"
)

func clientCreationHTTPFixture(t *testing.T) (*gorm.DB, func(uint, string) *gin.Engine) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	t.Setenv("SUBJECT_DATA_KEY", strings.Repeat("h", 32))
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "clients.db")), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	conn, err := db.DB()
	require.NoError(t, err)
	t.Cleanup(func() { _ = conn.Close() })
	require.NoError(t, db.AutoMigrate(&models.Client{}, &models.ClientCreationRequest{}))
	// Only columns used by the actual authorization/list queries are needed.
	require.NoError(t, db.Exec(`CREATE TABLE cases (id INTEGER PRIMARY KEY, client_id BIGINT,
		deleted_at TIMESTAMP, ethical_wall_enabled BOOLEAN, lawyer_id BIGINT, created_by TEXT)`).Error)
	require.NoError(t, db.Exec(`CREATE TABLE case_ethical_wall_whitelist (case_id BIGINT, user_id BIGINT)`).Error)
	makeRouter := func(actor uint, role string) *gin.Engine {
		repo := repositories.NewClientRepository(db)
		h := NewClientHandler(services.NewClientService(repo), services.NewAuthorizationService(repositories.NewCaseRepository(db), repo, nil, nil))
		h.SetEthicalWallListFilterEnabled(true)
		router := gin.New()
		router.Use(func(c *gin.Context) { c.Set("user_id", actor); c.Set("role", role); c.Next() })
		router.POST("/api/v1/clients", h.CreateClient)
		router.GET("/api/v1/clients", h.ListClients)
		router.GET("/api/v1/clients/:id", h.GetClient)
		return router
	}
	return db, makeRouter
}

func postClientCreation(t *testing.T, router *gin.Engine, payload map[string]interface{}, keys ...string) *httptest.ResponseRecorder {
	t.Helper()
	body, err := json.Marshal(payload)
	require.NoError(t, err)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/clients", strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")
	for _, key := range keys {
		req.Header.Add("Idempotency-Key", key)
	}
	response := httptest.NewRecorder()
	router.ServeHTTP(response, req)
	return response
}

func clientCreationPayload() map[string]interface{} {
	return map[string]interface{}{"name": "虚构同名客户", "type": "个人", "identity_number": "310110199001010011", "created_by": 999}
}

func clientCreationResult(t *testing.T, response *httptest.ResponseRecorder) services.ClientResponse {
	t.Helper()
	require.Equal(t, http.StatusOK, response.Code, response.Body.String())
	var envelope struct {
		Success bool
		Data    services.ClientResponse
	}
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &envelope))
	require.True(t, envelope.Success)
	require.NotZero(t, envelope.Data.ID)
	return envelope.Data
}

func requireClientCreationError(t *testing.T, response *httptest.ResponseRecorder, status int, code string) {
	t.Helper()
	require.Equal(t, status, response.Code, response.Body.String())
	var envelope common.APIResponse
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &envelope))
	require.False(t, envelope.Success)
	require.Nil(t, envelope.Data)
	if code != "" {
		require.Equal(t, code, envelope.Error.Code)
	}
}

func TestClientCreationHTTPContract(t *testing.T) {
	db, routerFor := clientCreationHTTPFixture(t)
	router := routerFor(7, "lawyer")
	payload := clientCreationPayload()
	firstResponse := postClientCreation(t, router, payload, "http-key")
	first := clientCreationResult(t, firstResponse)
	require.NotContains(t, firstResponse.Body.String(), payload["identity_number"])
	var stored models.Client
	require.NoError(t, db.First(&stored, first.ID).Error)
	require.Equal(t, uint(7), stored.CreatedBy)
	// Simulate a lost response by constructing a new handler/service instance.
	again := clientCreationResult(t, postClientCreation(t, routerFor(7, "lawyer"), payload, "http-key"))
	require.Equal(t, first.ID, again.ID)
	payload["identity_number"] = "310110199001010022"
	requireClientCreationError(t, postClientCreation(t, router, payload, "http-key"), 409, "CLIENT_CREATE_KEY_CONFLICT")
	payload["identity_number"] = "310110199001010011"
	for _, keys := range [][]string{{""}, {" "}, {"with space"}, {"中文"}, {strings.Repeat("k", 121)}, {"key-one", "key-two"}} {
		requireClientCreationError(t, postClientCreation(t, router, payload, keys...), 400, "CLIENT_CREATE_KEY_INVALID")
	}
	for _, role := range []string{"assistant", "intake_assistant", "admin", "user", "finance", "conflict_officer"} {
		requireClientCreationError(t, postClientCreation(t, routerFor(7, role), payload, "http-key"), 403, "")
	}
	requireClientCreationError(t, postClientCreation(t, routerFor(0, "lawyer"), payload, "http-key"), 401, "")
	var count int64
	require.NoError(t, db.Model(&models.Client{}).Count(&count).Error)
	require.EqualValues(t, 1, count)

	// A later ethical wall must take effect even though this actor created it.
	require.NoError(t, db.Exec("INSERT INTO cases (id, client_id, ethical_wall_enabled, lawyer_id, created_by) VALUES (1, ?, TRUE, 8, '8')", first.ID).Error)
	denied := postClientCreation(t, router, payload, "http-key")
	requireClientCreationError(t, denied, 403, "")
	require.NotContains(t, denied.Body.String(), "虚构同名客户")
	require.NoError(t, db.Exec("INSERT INTO case_ethical_wall_whitelist (case_id, user_id) VALUES (1, 7)").Error)
	restored := clientCreationResult(t, postClientCreation(t, router, payload, "http-key"))
	require.Equal(t, first.ID, restored.ID)
}

func TestClientCreationHTTPFailuresAndLegacyCompatibility(t *testing.T) {
	db, routerFor := clientCreationHTTPFixture(t)
	router := routerFor(7, "lawyer")
	// Generic rejection must not reveal a restricted customer's email/name.
	existing := clientCreationPayload()
	existing["name"] = "受限客户名称"
	existing["email"] = "restricted@example.invalid"
	clientCreationResult(t, postClientCreation(t, routerFor(8, "lawyer"), existing))
	payload := clientCreationPayload()
	payload["email"] = "restricted@example.invalid"
	rejected := postClientCreation(t, router, payload, "duplicate-email")
	requireClientCreationError(t, rejected, 400, "CLIENT_CREATE_REJECTED")
	for _, forbidden := range []string{"受限客户名称", "restricted@example.invalid", "already exists", "SELECT", "INSERT"} {
		require.NotContains(t, rejected.Body.String(), forbidden)
	}
	t.Setenv("SUBJECT_DATA_KEY", "")
	requireClientCreationError(t, postClientCreation(t, router, clientCreationPayload(), "no-secret"), 503, "SUBJECT_DATA_KEY_REQUIRED")
	legacy := map[string]interface{}{"name": "原有无身份客户", "type": "个人"}
	first := clientCreationResult(t, postClientCreation(t, router, legacy))
	second := clientCreationResult(t, postClientCreation(t, router, legacy))
	require.NotEqual(t, first.ID, second.ID)
	t.Setenv("SUBJECT_DATA_KEY", strings.Repeat("h", 32))
	require.NoError(t, db.Migrator().DropTable(&models.ClientCreationRequest{}))
	requireClientCreationError(t, postClientCreation(t, router, clientCreationPayload(), "no-ledger"), 503, "CLIENT_CREATE_UNAVAILABLE")
	clientCreationResult(t, postClientCreation(t, router, legacy))
}

func TestClientCreationC03SameNameDistinctIdentityAndAuthorizedPagination(t *testing.T) {
	_, routerFor := clientCreationHTTPFixture(t)
	router := routerFor(7, "lawyer")
	ids := make(map[uint]bool)
	for i := 0; i < 65; i++ {
		payload := clientCreationPayload()
		payload["identity_number"] = fmt.Sprintf("31011019900101%04d", i)
		created := clientCreationResult(t, postClientCreation(t, router, payload, fmt.Sprintf("c03-%d", i)))
		require.False(t, ids[created.ID], "same name must not merge distinct identities")
		ids[created.ID] = true
	}
	other := clientCreationResult(t, postClientCreation(t, routerFor(8, "lawyer"), clientCreationPayload(), "other-owner"))
	seen := make(map[uint]bool)
	for page := 1; page <= 2; page++ {
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/clients?page=%d&page_size=50&search=%s", page, url.QueryEscape("虚构同名客户")), nil)
		response := httptest.NewRecorder()
		router.ServeHTTP(response, req)
		require.Equal(t, 200, response.Code, response.Body.String())
		var envelope struct {
			Data struct {
				Clients    []services.ClientResponse
				Pagination struct{ Total int }
			}
		}
		require.NoError(t, json.Unmarshal(response.Body.Bytes(), &envelope))
		require.Equal(t, 65, envelope.Data.Pagination.Total)
		if page == 1 {
			require.Len(t, envelope.Data.Clients, 50)
		} else {
			require.Len(t, envelope.Data.Clients, 15)
		}
		for _, client := range envelope.Data.Clients {
			require.NotEqual(t, other.ID, client.ID)
			require.False(t, seen[client.ID])
			seen[client.ID] = true
		}
	}
	require.Equal(t, ids, seen)
}
