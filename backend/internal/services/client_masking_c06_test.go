package services

import (
	"context"
	"path/filepath"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"law-oa-go/internal/models"
	"law-oa-go/internal/repositories"
	"law-oa-go/internal/security"
)

func newC06MaskSvcDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "c06svc.db")), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&models.Client{}, &models.ClientContact{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

func c06MaskSvcSeal(t *testing.T, value string) string {
	t.Helper()
	t.Setenv("SUBJECT_DATA_KEY", "0123456789abcdef0123456789abcdef")
	sealed, err := security.ProtectSensitiveValue(clientContactCipherPurpose, value)
	if err != nil {
		t.Fatalf("seal: %v", err)
	}
	return sealed
}

func c06MaskSvcOpen(t *testing.T, sealed string) string {
	t.Helper()
	value, err := security.DecryptSensitiveValue(clientContactCipherPurpose, sealed)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	return value
}

func strPtr(value string) *string {
	return &value
}

func TestC06MaskUnit_MaskFormats(t *testing.T) {
	if got := models.MaskContactPerson("王建国"); got != "王**" {
		t.Fatalf("MaskContactPerson = %q, want 王**", got)
	}
	if got := models.MaskContactPerson(""); got != "" {
		t.Fatalf("MaskContactPerson(empty) = %q, want empty", got)
	}
	if got := models.MaskEmail("wjg@mask-test.example"); got != "w***@mask-test.example" {
		t.Fatalf("MaskEmail = %q, want w***@mask-test.example", got)
	}
	if got := models.MaskPhone("13900139000"); got != "139****9000" {
		t.Fatalf("MaskPhone = %q, want 139****9000", got)
	}
}

func TestC06WriteGuard_UpdateClientKeepsMaskedEcho(t *testing.T) {
	db := newC06MaskSvcDB(t)
	client := &models.Client{
		Name: "回显守卫客户", Type: "企业", Phone: "13800138000",
		ContactPerson: "王建国", ContactPhone: "13900139000",
		CreatedBy: 3101, Status: "active", Version: 1,
	}
	if err := db.Create(client).Error; err != nil {
		t.Fatalf("seed client: %v", err)
	}
	svc := NewClientService(repositories.NewClientRepository(db), db)
	version := uint(1)
	req := &UpdateClientRequest{
		Version:       &version,
		Phone:         strPtr(models.MaskPhone(client.Phone)),
		ContactPerson: strPtr(models.MaskContactPerson(client.ContactPerson)),
		ContactPhone:  strPtr(models.MaskPhone(client.ContactPhone)),
	}
	if _, err := svc.UpdateClient(context.Background(), client.ID, req); err != nil {
		t.Fatalf("UpdateClient masked echo: %v", err)
	}
	var raw models.Client
	if err := db.First(&raw, client.ID).Error; err != nil {
		t.Fatalf("reload: %v", err)
	}
	if raw.Phone != "13800138000" || raw.ContactPerson != "王建国" || raw.ContactPhone != "13900139000" {
		t.Fatalf("masked echo overwrote stored values: phone=%q contact_person=%q contact_phone=%q", raw.Phone, raw.ContactPerson, raw.ContactPhone)
	}
	if raw.Version != 2 {
		t.Fatalf("version = %d, want 2 after accepted update", raw.Version)
	}
}

func TestC06WriteGuard_UpdateClientAppliesRealEdits(t *testing.T) {
	db := newC06MaskSvcDB(t)
	client := &models.Client{
		Name: "真实编辑客户", Type: "企业", Phone: "13800138000",
		ContactPerson: "王建国", ContactPhone: "13900139000",
		CreatedBy: 3101, Status: "active", Version: 1,
	}
	if err := db.Create(client).Error; err != nil {
		t.Fatalf("seed client: %v", err)
	}
	svc := NewClientService(repositories.NewClientRepository(db), db)
	version := uint(1)
	req := &UpdateClientRequest{
		Version:       &version,
		Phone:         strPtr("13911112222"),
		ContactPerson: strPtr("李四光"),
		ContactPhone:  strPtr("13755556666"),
	}
	if _, err := svc.UpdateClient(context.Background(), client.ID, req); err != nil {
		t.Fatalf("UpdateClient real edit: %v", err)
	}
	var raw models.Client
	if err := db.First(&raw, client.ID).Error; err != nil {
		t.Fatalf("reload: %v", err)
	}
	if raw.Phone != "13911112222" || raw.ContactPerson != "李四光" || raw.ContactPhone != "13755556666" {
		t.Fatalf("real edit not applied: phone=%q contact_person=%q contact_phone=%q", raw.Phone, raw.ContactPerson, raw.ContactPhone)
	}
}

func TestC06WriteGuard_SavePrimaryContactKeepsMaskedEcho(t *testing.T) {
	db := newC06MaskSvcDB(t)
	client := &models.Client{Name: "主联系人客户", Type: "企业", CreatedBy: 3101, Status: "active", Version: 1}
	if err := db.Create(client).Error; err != nil {
		t.Fatalf("seed client: %v", err)
	}
	contact := models.ClientContact{
		ClientID: client.ID, Name: "王建国", Position: "法务总监",
		PhoneCiphertext: c06MaskSvcSeal(t, "13900139000"),
		EmailCiphertext: c06MaskSvcSeal(t, "wjg@mask-test.example"),
		IsPrimary:       true, Version: 1, CreatedBy: 3101, UpdatedBy: 3101,
	}
	if err := db.Create(&contact).Error; err != nil {
		t.Fatalf("seed contact: %v", err)
	}
	svc := NewClientContactService(db)
	req := SavePrimaryContactRequest{
		Version: 1, Name: "王**", Position: "法务总监",
		Phone: "139****9000", Email: "w***@mask-test.example",
	}
	saved, err := svc.SavePrimaryContact(context.Background(), client.ID, 3101, req)
	if err != nil {
		t.Fatalf("SavePrimaryContact masked echo: %v", err)
	}
	if saved.Version != 2 {
		t.Fatalf("saved version = %d, want 2", saved.Version)
	}
	var raw models.ClientContact
	if err := db.Where("client_id = ? AND is_primary = ?", client.ID, true).First(&raw).Error; err != nil {
		t.Fatalf("reload contact: %v", err)
	}
	if raw.Name != "王建国" {
		t.Fatalf("contact name = %q, want 王建国", raw.Name)
	}
	if got := c06MaskSvcOpen(t, raw.PhoneCiphertext); got != "13900139000" {
		t.Fatalf("contact phone = %q, want 13900139000", got)
	}
	if got := c06MaskSvcOpen(t, raw.EmailCiphertext); got != "wjg@mask-test.example" {
		t.Fatalf("contact email = %q, want wjg@mask-test.example", got)
	}
}

func TestC06WriteGuard_SavePrimaryContactAppliesRealEdits(t *testing.T) {
	db := newC06MaskSvcDB(t)
	client := &models.Client{Name: "主联系人编辑客户", Type: "企业", CreatedBy: 3101, Status: "active", Version: 1}
	if err := db.Create(client).Error; err != nil {
		t.Fatalf("seed client: %v", err)
	}
	contact := models.ClientContact{
		ClientID: client.ID, Name: "王建国", Position: "法务总监",
		PhoneCiphertext: c06MaskSvcSeal(t, "13900139000"),
		EmailCiphertext: c06MaskSvcSeal(t, "wjg@mask-test.example"),
		IsPrimary:       true, Version: 1, CreatedBy: 3101, UpdatedBy: 3101,
	}
	if err := db.Create(&contact).Error; err != nil {
		t.Fatalf("seed contact: %v", err)
	}
	svc := NewClientContactService(db)
	req := SavePrimaryContactRequest{
		Version: 1, Name: "李四光", Position: "法务总监",
		Phone: "13755556666", Email: "lsg@mask-test.example",
	}
	if _, err := svc.SavePrimaryContact(context.Background(), client.ID, 3101, req); err != nil {
		t.Fatalf("SavePrimaryContact real edit: %v", err)
	}
	var raw models.ClientContact
	if err := db.Where("client_id = ? AND is_primary = ?", client.ID, true).First(&raw).Error; err != nil {
		t.Fatalf("reload contact: %v", err)
	}
	if raw.Name != "李四光" {
		t.Fatalf("contact name = %q, want 李四光", raw.Name)
	}
	if got := c06MaskSvcOpen(t, raw.PhoneCiphertext); got != "13755556666" {
		t.Fatalf("contact phone = %q, want 13755556666", got)
	}
	if got := c06MaskSvcOpen(t, raw.EmailCiphertext); got != "lsg@mask-test.example" {
		t.Fatalf("contact email = %q, want lsg@mask-test.example", got)
	}
}
