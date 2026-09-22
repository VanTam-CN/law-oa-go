//go:build ignore
// +build ignore

package repositories

import (
	"context"
	"testing"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"law-oa-go/internal/models"
)

// setupPostgreSQLTestDB 创建PostgreSQL测试数据库连接
func setupPostgreSQLTestDB(t *testing.T) *gorm.DB {
	dsn := "host=localhost port=5432 user=law_oa_user password=law_oa_password dbname=law_oa_db_test sslmode=disable TimeZone=UTC"

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to PostgreSQL test database: %v", err)
	}

	// 自动迁移模型
	err = db.AutoMigrate(&models.User{}, &models.Client{}, &models.Case{})
	if err != nil {
		t.Fatalf("Failed to migrate models: %v", err)
	}

	// 清理测试数据
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("Failed to get underlying sql.DB: %v", err)
	}

	// 删除所有测试数据（重置序列）
	tables := []string{"cases", "clients", "users"}
	for _, table := range tables {
		_, err := sqlDB.Exec("DELETE FROM " + table)
		if err != nil {
			t.Logf("Warning: Failed to clean table %s: %v", table, err)
		}
		// 重置自增序列
		_, err = sqlDB.Exec("ALTER SEQUENCE " + table + "_id_seq RESTART WITH 1")
		if err != nil {
			t.Logf("Warning: Failed to reset sequence for %s: %v", table, err)
		}
	}

	return db
}

// teardownPostgreSQLTestDB 清理PostgreSQL测试数据库
func teardownPostgreSQLTestDB(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// TestPostgreSQL_ClientRepository_Create 测试PostgreSQL下的客户创建
func TestPostgreSQL_ClientRepository_Create(t *testing.T) {
	db := setupPostgreSQLTestDB(t)
	defer teardownPostgreSQLTestDB(db)

	repo := NewClientRepository(db)
	ctx := context.Background()

	client := &models.Client{
		Name:      "PostgreSQL测试客户",
		Email:     "pgsql_test@example.com",
		Phone:     "13900139001",
		Address:   "PostgreSQL测试地址",
		Status:    "active",
		Type:      "个人",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := repo.Create(ctx, client)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	if client.ID == 0 {
		t.Error("Client ID should not be zero after creation")
	}

	// 验证客户确实被创建
	var foundClient models.Client
	err = db.First(&foundClient, client.ID).Error
	if err != nil {
		t.Fatalf("Failed to find created client: %v", err)
	}

	if foundClient.Name != client.Name {
		t.Errorf("Expected client name %s, got %s", client.Name, foundClient.Name)
	}

	if foundClient.Email != client.Email {
		t.Errorf("Expected client email %s, got %s", client.Email, foundClient.Email)
	}
}

// TestPostgreSQL_ClientRepository_FindByID 测试PostgreSQL下的客户查找
func TestPostgreSQL_ClientRepository_FindByID(t *testing.T) {
	db := setupPostgreSQLTestDB(t)
	defer teardownPostgreSQLTestDB(db)

	repo := NewClientRepository(db)
	ctx := context.Background()

	// 创建测试客户
	testClient := createTestClientPostgreSQL(db)

	// 测试查找存在的客户
	foundClient, err := repo.FindByID(ctx, testClient.ID)
	if err != nil {
		t.Fatalf("Failed to find client: %v", err)
	}

	if foundClient == nil {
		t.Fatal("Found client should not be nil")
	}

	if foundClient.Name != testClient.Name {
		t.Errorf("Expected client name %s, got %s", testClient.Name, foundClient.Name)
	}

	// 测试查找不存在的客户
	notFoundClient, err := repo.FindByID(ctx, 999999)
	if err != nil {
		t.Errorf("FindByID should not return error for non-existent client: %v", err)
	}

	if notFoundClient != nil {
		t.Error("Non-existent client should return nil")
	}
}

// TestPostgreSQL_ClientRepository_List 测试PostgreSQL下的客户列表查询
func TestPostgreSQL_ClientRepository_List(t *testing.T) {
	db := setupPostgreSQLTestDB(t)
	defer teardownPostgreSQLTestDB(db)

	repo := NewClientRepository(db)
	ctx := context.Background()

	// 创建多个测试客户
	clients := []*models.Client{
		{Name: "客户A", Email: "clienta@example.com", Status: "active", Type: "个人"},
		{Name: "客户B", Email: "clientb@example.com", Status: "active", Type: "企业"},
		{Name: "客户C", Email: "clientc@example.com", Status: "inactive", Type: "个人"},
	}

	for _, client := range clients {
		client.CreatedAt = time.Now()
		client.UpdatedAt = time.Now()
		if err := db.Create(client).Error; err != nil {
			t.Fatalf("Failed to create test client: %v", err)
		}
	}

	// 测试列表查询
	params := &ClientListParams{
		Page:     1,
		PageSize: 10,
	}

	result, total, err := repo.List(ctx, params)
	if err != nil {
		t.Fatalf("Failed to list clients: %v", err)
	}

	if total != 3 {
		t.Errorf("Expected total count 3, got %d", total)
	}

	if len(result) != 3 {
		t.Errorf("Expected 3 clients, got %d", len(result))
	}

	// 测试状态过滤
	params.Status = "active"
	result, total, err = repo.List(ctx, params)
	if err != nil {
		t.Fatalf("Failed to list active clients: %v", err)
	}

	if total != 2 {
		t.Errorf("Expected 2 active clients, got %d", total)
	}

	// 测试类型过滤
	params.Status = ""
	params.Type = "企业"
	result, total, err = repo.List(ctx, params)
	if err != nil {
		t.Fatalf("Failed to list enterprise clients: %v", err)
	}

	if total != 1 {
		t.Errorf("Expected 1 enterprise client, got %d", total)
	}
}

// TestPostgreSQL_ClientRepository_Update 测试PostgreSQL下的客户更新
func TestPostgreSQL_ClientRepository_Update(t *testing.T) {
	db := setupPostgreSQLTestDB(t)
	defer teardownPostgreSQLTestDB(db)

	repo := NewClientRepository(db)
	ctx := context.Background()

	// 创建测试客户
	testClient := createTestClientPostgreSQL(db)

	// 更新客户信息
	testClient.Name = "更新后的客户名称"
	testClient.Address = "更新后的地址"
	testClient.UpdatedAt = time.Now()

	err := repo.Update(ctx, testClient)
	if err != nil {
		t.Fatalf("Failed to update client: %v", err)
	}

	// 验证更新是否成功
	var updatedClient models.Client
	err = db.First(&updatedClient, testClient.ID).Error
	if err != nil {
		t.Fatalf("Failed to find updated client: %v", err)
	}

	if updatedClient.Name != "更新后的客户名称" {
		t.Errorf("Expected updated name '更新后的客户名称', got '%s'", updatedClient.Name)
	}

	if updatedClient.Address != "更新后的地址" {
		t.Errorf("Expected updated address '更新后的地址', got '%s'", updatedClient.Address)
	}
}

// TestPostgreSQL_ClientRepository_Delete 测试PostgreSQL下的客户删除
func TestPostgreSQL_ClientRepository_Delete(t *testing.T) {
	db := setupPostgreSQLTestDB(t)
	defer teardownPostgreSQLTestDB(db)

	repo := NewClientRepository(db)
	ctx := context.Background()

	// 创建测试客户
	testClient := createTestClientPostgreSQL(db)

	// 删除客户
	err := repo.Delete(ctx, testClient.ID)
	if err != nil {
		t.Fatalf("Failed to delete client: %v", err)
	}

	// 验证客户是否被删除（软删除）
	var deletedClient models.Client
	err = db.First(&deletedClient, testClient.ID).Error
	if err == nil {
		t.Error("Soft deleted client should not be found with First()")
	}

	// 检查软删除标记
	err = db.Unscoped().First(&deletedClient, testClient.ID).Error
	if err != nil {
		t.Errorf("Soft deleted client should be found with Unscoped(): %v", err)
	}

	if deletedClient.DeletedAt.Time.IsZero() {
		t.Error("DeletedAt should not be zero for soft deleted client")
	}
}

// createTestClientPostgreSQL 创建PostgreSQL测试客户
func createTestClientPostgreSQL(db *gorm.DB) *models.Client {
	client := &models.Client{
		Name:      "PostgreSQL测试客户",
		Email:     "pgsql_client@example.com",
		Phone:     "13900139002",
		Address:   "PostgreSQL测试地址",
		Status:    "active",
		Type:      "个人",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := db.Create(client).Error; err != nil {
		panic(err)
	}

	return client
}
