package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"law-oa-go/internal/models"
	"law-oa-go/internal/repositories"
	"law-oa-go/internal/security"
	"law-oa-go/internal/services"
)

const (
	c06MaskLawyer1UID uint = 3101
	c06MaskLawyer2UID uint = 3102
	c06MaskManagerUID uint = 3103
)
const c06MaskTechAdminUID uint = 3104
const c06MaskLawyer2Text string = "3102"

func newC06MaskingTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(t.TempDir()+"/c06mask.db"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&models.Client{}, &models.Case{}, &models.CaseEthicalWallWhitelist{}, &models.ClientContact{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

func seedC06MaskingClient(t *testing.T, db *gorm.DB, createdBy uint) *models.Client {
	t.Helper()
	client := &models.Client{
		Name:          "C06掩码验证客户",
		Type:          "企业",
		Email:         "ops@mask-test.example",
		Phone:         "13800138000",
		ContactPerson: "王建国",
		ContactPhone:  "13900139000",
		CreatedBy:     createdBy,
		Status:        "active",
		Version:       1,
	}
	if err := db.Create(client).Error; err != nil {
		t.Fatalf("seed client: %v", err)
	}
	return client
}

func newC06MaskingRouter(t *testing.T, db *gorm.DB, userID uint, role string) *gin.Engine {
	t.Helper()
	clientService := services.NewClientService(repositories.NewClientRepository(db), db)
	authz := services.NewAuthorizationService(
		repositories.NewCaseRepository(db),
		repositories.NewClientRepository(db),
		repositories.NewUserRepository(db),
		repositories.NewDocumentRepository(db),
	)
	clientHandler := NewClientHandler(clientService, authz)
	clientHandler.SetContactService(services.NewClientContactService(db))
	aggregate := NewDemoAggregateHandler(db)
	aggregate.SetAuthorizationService(authz)
	aggregate.SetClientContactService(services.NewClientContactService(db))

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", userID)
		c.Set("role", role)
		c.Next()
	})
	router.GET("/clients/:id", clientHandler.GetClient)
	router.GET("/clients", clientHandler.ListClients)
	router.PUT("/clients/:id", clientHandler.UpdateClient)
	router.GET("/clients/:id/primary-contact", clientHandler.GetPrimaryContact)
	router.PUT("/clients/:id/primary-contact", clientHandler.SavePrimaryContact)
	router.GET("/clients/:id/master-profile", aggregate.ClientMasterProfile)
	router.GET("/case-intakes/:id", aggregate.IntakeWorkbench)
	return router
}

func c06MaskingGetJSON(t *testing.T, router *gin.Engine, path string) map[string]interface{} {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("%s status = %d, want 200; body = %s", path, w.Code, w.Body.String())
	}
	var payload struct {
		Success bool                   `json:"success"`
		Data    map[string]interface{} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
		t.Fatalf("%s decode: %v; body = %s", path, err, w.Body.String())
	}
	if !payload.Success {
		t.Fatalf("%s success=false; body = %s", path, w.Body.String())
	}
	return payload.Data
}

func strconvUint(id uint) string { return strconv.FormatUint(uint64(id), 10) }

func mustC06Seal(t *testing.T, value string) string {
	t.Helper()
	t.Setenv("SUBJECT_DATA_KEY", "0123456789abcdef0123456789abcdef")
	sealed, err := security.ProtectSensitiveValue("client-contact-v1", value)
	if err != nil {
		t.Fatalf("seal: %v", err)
	}
	return sealed
}

func TestC06Masking_ManagerSeesMaskedClientMaster(t *testing.T) {
	db := newC06MaskingTestDB(t)
	client := seedC06MaskingClient(t, db, c06MaskLawyer1UID)
	router := newC06MaskingRouter(t, db, c06MaskManagerUID, "director")

	detail := c06MaskingGetJSON(t, router, "/clients/"+strconvUint(client.ID))
	if got, _ := detail["contact_person"].(string); got != "王**" {
		t.Fatalf("detail contact_person = %q, want 王**", got)
	}
	if got, _ := detail["contact_phone"].(string); got != "139****9000" {
		t.Fatalf("detail contact_phone = %q, want 139****9000", got)
	}

	list := c06MaskingGetJSON(t, router, "/clients")
	rows, _ := list["clients"].([]interface{})
	if len(rows) != 1 {
		t.Fatalf("list rows = %d, want 1", len(rows))
	}
	row, _ := rows[0].(map[string]interface{})
	if got, _ := row["contact_person"].(string); got != "王**" {
		t.Fatalf("list contact_person = %q, want 王**", got)
	}
}

func TestC06Masking_OwningLawyerKeepsExistingReadSemantics(t *testing.T) {
	db := newC06MaskingTestDB(t)
	client := seedC06MaskingClient(t, db, c06MaskLawyer1UID)
	router := newC06MaskingRouter(t, db, c06MaskLawyer1UID, "lawyer")

	detail := c06MaskingGetJSON(t, router, "/clients/"+strconvUint(client.ID))
	if got, _ := detail["contact_person"].(string); got != "王建国" {
		t.Fatalf("detail contact_person = %q, want plaintext 王建国", got)
	}
	// Phone/ContactPhone were already statically masked for every role before
	// C06; plan A keeps that pre-existing lawyer-visible behavior.
	if got, _ := detail["phone"].(string); got != "138****8000" {
		t.Fatalf("detail phone = %q, want pre-existing static mask 138****8000", got)
	}
	if got, _ := detail["contact_phone"].(string); got != "139****9000" {
		t.Fatalf("detail contact_phone = %q, want pre-existing static mask 139****9000", got)
	}

	list := c06MaskingGetJSON(t, router, "/clients")
	rows, _ := list["clients"].([]interface{})
	if len(rows) != 1 {
		t.Fatalf("list rows = %d, want 1", len(rows))
	}
	row, _ := rows[0].(map[string]interface{})
	if got, _ := row["contact_person"].(string); got != "王建国" {
		t.Fatalf("list contact_person = %q, want plaintext", got)
	}
}

func TestC06Masking_TechAdminAndEthicalWallDenied(t *testing.T) {
	db := newC06MaskingTestDB(t)
	client := seedC06MaskingClient(t, db, c06MaskLawyer1UID)

	techRouter := newC06MaskingRouter(t, db, c06MaskTechAdminUID, "admin")
	req := httptest.NewRequest(http.MethodGet, "/clients/"+strconvUint(client.ID), nil)
	w := httptest.NewRecorder()
	techRouter.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Fatalf("tech admin status = %d, want 403; body = %s", w.Code, w.Body.String())
	}

	wallCase := models.Case{
		CaseNumber:         "C06WALL-001",
		Title:              "wall case",
		ClientID:           client.ID,
		LawyerID:           c06MaskLawyer1UID,
		CaseType:           "civil",
		Status:             "active",
		CreatedBy:          c06MaskLawyer2Text,
		EthicalWallEnabled: true,
	}
	if err := db.Create(&wallCase).Error; err != nil {
		t.Fatalf("seed wall case: %v", err)
	}
	wallRouter := newC06MaskingRouter(t, db, c06MaskManagerUID, "director")
	req2 := httptest.NewRequest(http.MethodGet, "/clients/"+strconvUint(client.ID), nil)
	w2 := httptest.NewRecorder()
	wallRouter.ServeHTTP(w2, req2)
	if w2.Code != http.StatusForbidden {
		t.Fatalf("ethical wall status = %d, want 403; body = %s", w2.Code, w2.Body.String())
	}
}

func TestC06Masking_PrimaryContactMaskedForManagerPlaintextForLawyer(t *testing.T) {
	db := newC06MaskingTestDB(t)
	client := seedC06MaskingClient(t, db, c06MaskLawyer1UID)
	contact := models.ClientContact{
		ClientID: client.ID, Name: "王建国", Position: "法务总监",
		PhoneCiphertext: mustC06Seal(t, "13900139000"),
		EmailCiphertext: mustC06Seal(t, "wjg@mask-test.example"),
		IsPrimary:       true, Version: 1,
		CreatedBy: c06MaskLawyer1UID, UpdatedBy: c06MaskLawyer1UID,
	}
	if err := db.Create(&contact).Error; err != nil {
		t.Fatalf("seed contact: %v", err)
	}

	managerRouter := newC06MaskingRouter(t, db, c06MaskManagerUID, "director")
	masked := c06MaskingGetJSON(t, managerRouter, "/clients/"+strconvUint(client.ID)+"/primary-contact")
	if got, _ := masked["name"].(string); got != "王**" {
		t.Fatalf("primary contact name = %q, want 王**", got)
	}
	if got, _ := masked["phone"].(string); got != "139****9000" {
		t.Fatalf("primary contact phone = %q, want masked", got)
	}
	if got, _ := masked["email"].(string); got != "w***@mask-test.example" {
		t.Fatalf("primary contact email = %q, want masked", got)
	}

	lawyerRouter := newC06MaskingRouter(t, db, c06MaskLawyer1UID, "lawyer")
	plain := c06MaskingGetJSON(t, lawyerRouter, "/clients/"+strconvUint(client.ID)+"/primary-contact")
	if got, _ := plain["name"].(string); got != "王建国" {
		t.Fatalf("lawyer contact name = %q, want plaintext", got)
	}
	if got, _ := plain["phone"].(string); got != "13900139000" {
		t.Fatalf("lawyer contact phone = %q, want plaintext", got)
	}
	if got, _ := plain["email"].(string); got != "wjg@mask-test.example" {
		t.Fatalf("lawyer contact email = %q, want plaintext", got)
	}
}

func TestC06Masking_MasterProfileMasksManagerContact(t *testing.T) {
	db := newC06MaskingTestDB(t)
	client := seedC06MaskingClient(t, db, c06MaskLawyer1UID)
	router := newC06MaskingRouter(t, db, c06MaskManagerUID, "director")

	profile := c06MaskingGetJSON(t, router, "/clients/"+strconvUint(client.ID)+"/master-profile")
	clientRow, _ := profile["client"].(map[string]interface{})
	if got, _ := clientRow["contact_person"].(string); got != "王**" {
		t.Fatalf("profile client contact_person = %q, want 王**", got)
	}
	primary, _ := profile["primary_contact"].(map[string]interface{})
	if primary == nil {
		t.Fatalf("profile primary_contact missing; body = %v", profile)
	}
	if got, _ := primary["name"].(string); got != "王**" {
		t.Fatalf("profile primary name = %q, want 王**", got)
	}
}

func c06MaskingPutJSON(t *testing.T, router *gin.Engine, path string, body map[string]interface{}) (int, map[string]interface{}) {
	t.Helper()
	raw, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal body: %v", err)
	}
	req := httptest.NewRequest(http.MethodPut, path, bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	var payload struct {
		Success bool                   `json:"success"`
		Data    map[string]interface{} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
		t.Fatalf("%s decode status=%d: %v; body = %s", path, w.Code, err, w.Body.String())
	}
	return w.Code, payload.Data
}

func TestC06Masking_ManagerMaskedEchoPutKeepsStoredClient(t *testing.T) {
	db := newC06MaskingTestDB(t)
	client := seedC06MaskingClient(t, db, c06MaskLawyer1UID)
	router := newC06MaskingRouter(t, db, c06MaskManagerUID, "director")

	status, data := c06MaskingPutJSON(t, router, "/clients/"+strconvUint(client.ID), map[string]interface{}{
		"version":        1,
		"phone":          "138****8000",
		"contact_person": "王**",
		"contact_phone":  "139****9000",
	})
	if status != http.StatusOK {
		t.Fatalf("masked echo put status = %d, want 200; data = %v", status, data)
	}
	var raw models.Client
	if err := db.First(&raw, client.ID).Error; err != nil {
		t.Fatalf("reload client: %v", err)
	}
	if raw.Phone != "13800138000" || raw.ContactPerson != "王建国" || raw.ContactPhone != "13900139000" {
		t.Fatalf("masked echo overwrote stored client: phone=%q contact_person=%q contact_phone=%q", raw.Phone, raw.ContactPerson, raw.ContactPhone)
	}

	status, data = c06MaskingPutJSON(t, router, "/clients/"+strconvUint(client.ID), map[string]interface{}{
		"version":        raw.Version,
		"phone":          "13911112222",
		"contact_person": "李四光",
		"contact_phone":  "13755556666",
	})
	if status != http.StatusOK {
		t.Fatalf("real edit put status = %d, want 200; data = %v", status, data)
	}
	if err := db.First(&raw, client.ID).Error; err != nil {
		t.Fatalf("reload client after edit: %v", err)
	}
	if raw.Phone != "13911112222" || raw.ContactPerson != "李四光" || raw.ContactPhone != "13755556666" {
		t.Fatalf("real edit not applied: phone=%q contact_person=%q contact_phone=%q", raw.Phone, raw.ContactPerson, raw.ContactPhone)
	}
}

func TestC06Masking_ManagerMaskedEchoPutKeepsPrimaryContact(t *testing.T) {
	db := newC06MaskingTestDB(t)
	client := seedC06MaskingClient(t, db, c06MaskLawyer1UID)
	contact := models.ClientContact{
		ClientID: client.ID, Name: "王建国", Position: "法务总监",
		PhoneCiphertext: mustC06Seal(t, "13900139000"),
		EmailCiphertext: mustC06Seal(t, "wjg@mask-test.example"),
		IsPrimary:       true, Version: 1, CreatedBy: c06MaskLawyer1UID, UpdatedBy: c06MaskLawyer1UID,
	}
	if err := db.Create(&contact).Error; err != nil {
		t.Fatalf("seed contact: %v", err)
	}
	router := newC06MaskingRouter(t, db, c06MaskManagerUID, "director")

	status, data := c06MaskingPutJSON(t, router, "/clients/"+strconvUint(client.ID)+"/primary-contact", map[string]interface{}{
		"version":  1,
		"name":     "王**",
		"position": "法务总监",
		"phone":    "139****9000",
		"email":    "w***@mask-test.example",
	})
	if status != http.StatusOK {
		t.Fatalf("masked echo put status = %d, want 200; data = %v", status, data)
	}
	var raw models.ClientContact
	if err := db.Where("client_id = ? AND is_primary = ?", client.ID, true).First(&raw).Error; err != nil {
		t.Fatalf("reload contact: %v", err)
	}
	if raw.Name != "王建国" {
		t.Fatalf("contact name = %q, want 王建国", raw.Name)
	}
}

func TestC06Masking_ManagerPrimaryContactPutMaskedEchoAndResponse(t *testing.T) {
	db := newC06MaskingTestDB(t)
	client := seedC06MaskingClient(t, db, c06MaskLawyer1UID)
	contact := models.ClientContact{
		ClientID: client.ID, Name: "王建国", Position: "法务总监",
		PhoneCiphertext: mustC06Seal(t, "13900139000"),
		EmailCiphertext: mustC06Seal(t, "wjg@mask-test.example"),
		IsPrimary:       true, Version: 1, CreatedBy: c06MaskLawyer1UID, UpdatedBy: c06MaskLawyer1UID,
	}
	if err := db.Create(&contact).Error; err != nil {
		t.Fatalf("seed contact: %v", err)
	}
	router := newC06MaskingRouter(t, db, c06MaskManagerUID, "director")

	status, data := c06MaskingPutJSON(t, router, "/clients/"+strconvUint(client.ID)+"/primary-contact", map[string]interface{}{
		"version":  1,
		"name":     "王**",
		"position": "法务总监",
		"phone":    "139****9000",
		"email":    "w***@mask-test.example",
	})
	if status != http.StatusOK {
		t.Fatalf("masked echo put status = %d, want 200; data = %v", status, data)
	}
	// The management-role PUT response itself must be masked, not plaintext.
	if got, _ := data["name"].(string); got != "王**" {
		t.Fatalf("PUT response name = %q, want 王**", got)
	}
	if got, _ := data["phone"].(string); got != "139****9000" {
		t.Fatalf("PUT response phone = %q, want masked", got)
	}
	if got, _ := data["email"].(string); got != "w***@mask-test.example" {
		t.Fatalf("PUT response email = %q, want masked", got)
	}
	// Stored plaintext and sealed ciphertext survive the masked echo.
	var raw models.ClientContact
	if err := db.Where("client_id = ? AND is_primary = ?", client.ID, true).First(&raw).Error; err != nil {
		t.Fatalf("reload contact: %v", err)
	}
	if raw.Name != "王建国" || raw.Position != "法务总监" {
		t.Fatalf("stored name/position = %q/%q, want 王建国/法务总监", raw.Name, raw.Position)
	}
	if got := c06MaskingUnseal(t, raw.PhoneCiphertext); got != "13900139000" {
		t.Fatalf("stored phone = %q, want 13900139000", got)
	}
	if got := c06MaskingUnseal(t, raw.EmailCiphertext); got != "wjg@mask-test.example" {
		t.Fatalf("stored email = %q, want wjg@mask-test.example", got)
	}
	if raw.Version != 2 {
		t.Fatalf("stored version = %d, want 2 after accepted save", raw.Version)
	}
}

func c06MaskingUnseal(t *testing.T, sealed string) string {
	t.Helper()
	t.Setenv("SUBJECT_DATA_KEY", "0123456789abcdef0123456789abcdef")
	value, err := security.DecryptSensitiveValue("client-contact-v1", sealed)
	if err != nil {
		t.Fatalf("unseal: %v", err)
	}
	return value
}

func TestC06Masking_BasicEmailVisibleInAggregates(t *testing.T) {
	db := newC06MaskingTestDB(t)
	client := seedC06MaskingClient(t, db, c06MaskLawyer1UID)
	// IntakeWorkbench reads case_intakes through map-backed queries; the
	// sqlite test schema only needs the columns the workbench actually uses.
	if err := db.Exec("CREATE TABLE case_intakes (id text PRIMARY KEY, intake_code text, title text, client_id integer, status text, created_by integer, metadata text, created_at datetime, updated_at datetime)").Error; err != nil {
		t.Fatalf("create intake table: %v", err)
	}
	intake := map[string]interface{}{
		"id": "c06mask-intake-0001", "intake_code": "C06MASK-I1",
		"title": "掩码验证接案", "client_id": client.ID,
		"status": "draft", "created_by": uint64(c06MaskLawyer1UID),
		"metadata": "{}", "created_at": time.Now(), "updated_at": time.Now(),
	}
	if err := db.Table("case_intakes").Create(intake).Error; err != nil {
		t.Fatalf("seed intake: %v", err)
	}

	managerRouter := newC06MaskingRouter(t, db, c06MaskManagerUID, "director")
	profile := c06MaskingGetJSON(t, managerRouter, "/clients/"+strconvUint(client.ID)+"/master-profile")
	profileClient, _ := profile["client"].(map[string]interface{})
	if got, _ := profileClient["email"].(string); got != "ops@mask-test.example" {
		t.Fatalf("master-profile email = %q, want ops@mask-test.example (basic metadata stays visible)", got)
	}
	workbench := c06MaskingGetJSON(t, managerRouter, "/case-intakes/c06mask-intake-0001")
	workbenchClient, _ := workbench["client"].(map[string]interface{})
	if got, _ := workbenchClient["email"].(string); got != "ops@mask-test.example" {
		t.Fatalf("intake workbench email = %q, want visible", got)
	}
	if got, _ := workbenchClient["contact_person"].(string); got != "王**" {
		t.Fatalf("intake workbench contact_person = %q, want 王**", got)
	}
	if got, _ := workbenchClient["phone"].(string); got != "138****8000" {
		t.Fatalf("intake workbench phone = %q, want masked", got)
	}
}

func TestC06Masking_CaseLinkedLawyerKeepsPlaintextContact(t *testing.T) {
	db := newC06MaskingTestDB(t)
	client := seedC06MaskingClient(t, db, c06MaskLawyer1UID)
	linkedCase := models.Case{
		CaseNumber: "C06LINK-001", Title: "关联案件", ClientID: client.ID,
		LawyerID: c06MaskLawyer2UID, CaseType: "civil", Status: "active",
		CreatedBy: c06MaskLawyer2Text,
	}
	if err := db.Create(&linkedCase).Error; err != nil {
		t.Fatalf("seed linked case: %v", err)
	}
	contact := models.ClientContact{
		ClientID: client.ID, Name: "王建国", Position: "法务总监",
		PhoneCiphertext: mustC06Seal(t, "13900139000"),
		EmailCiphertext: mustC06Seal(t, "wjg@mask-test.example"),
		IsPrimary:       true, Version: 1, CreatedBy: c06MaskLawyer1UID, UpdatedBy: c06MaskLawyer1UID,
	}
	if err := db.Create(&contact).Error; err != nil {
		t.Fatalf("seed contact: %v", err)
	}
	router := newC06MaskingRouter(t, db, c06MaskLawyer2UID, "lawyer")

	detail := c06MaskingGetJSON(t, router, "/clients/"+strconvUint(client.ID))
	if got, _ := detail["contact_person"].(string); got != "王建国" {
		t.Fatalf("case-linked lawyer contact_person = %q, want plaintext 王建国", got)
	}
	pc := c06MaskingGetJSON(t, router, "/clients/"+strconvUint(client.ID)+"/primary-contact")
	if got, _ := pc["name"].(string); got != "王建国" {
		t.Fatalf("case-linked lawyer primary-contact name = %q, want plaintext", got)
	}
	if got, _ := pc["phone"].(string); got != "13900139000" {
		t.Fatalf("case-linked lawyer primary-contact phone = %q, want plaintext", got)
	}
}
