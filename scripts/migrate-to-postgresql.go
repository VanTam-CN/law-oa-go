//go:build ignore

package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
)

// MySQL 数据结构
type MySQLUser struct {
	ID        uint      `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Username  string    `json:"username"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Password  string    `json:"password"`
	Role      string    `json:"role"`
	Phone     string    `json:"phone"`
	Avatar    string    `json:"avatar"`
	Status    string    `json:"status"`
}

type MySQLClient struct {
	ID           uint      `json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Name         string    `json:"name"`
	Type         string    `json:"type"`
	Email        string    `json:"email"`
	Phone        string    `json:"phone"`
	Address      string    `json:"address"`
	Company      string    `json:"company"`
	IDCard       string    `json:"id_card"`
	Industry     string    `json:"industry"`
	ContactPerson string    `json:"contact_person"`
	ContactPhone string    `json:"contact_phone"`
	Source       string    `json:"source"`
	Notes        string    `json:"notes"`
	Status       string    `json:"status"`
}

type MySQLCase struct {
	ID          uint      `json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	ClientID    uint      `json:"client_id"`
	LawyerID    uint      `json:"lawyer_id"`
	CaseType    string    `json:"case_type"`
	Priority    string    `json:"priority"`
	Status      string    `json:"status"`
	StartDate   *time.Time `json:"start_date"`
	EndDate     *time.Time `json:"end_date"`
}

// 配置结构
type Config struct {
	MySQLHost     string
	MySQLPort     string
	MySQLUser     string
	MySQLPassword string
	MySQLDatabase string

	PostgreSQLHost     string
	PostgreSQLPort     string
	PostgreSQLUser     string
	PostgreSQLPassword string
	PostgreSQLDatabase string
}

func main() {
	// 加载配置
	config := loadConfig()

	// 连接MySQL
	mysqlDB, err := connectMySQL(config)
	if err != nil {
		log.Fatalf("连接MySQL失败: %v", err)
	}
	defer mysqlDB.Close()

	// 连接PostgreSQL
	postgresDB, err := connectPostgreSQL(config)
	if err != nil {
		log.Fatalf("连接PostgreSQL失败: %v", err)
	}
	defer postgresDB.Close()

	fmt.Println("🚀 开始数据迁移...")
	fmt.Println("================================")

	// 开始迁移
	if err := migrateData(mysqlDB, postgresDB); err != nil {
		log.Fatalf("数据迁移失败: %v", err)
	}

	fmt.Println("✅ 数据迁移完成!")
}

func loadConfig() *Config {
	return &Config{
		MySQLHost:         getEnv("MYSQL_HOST", "localhost"),
		MySQLPort:         getEnv("MYSQL_PORT", "3306"),
		MySQLUser:         getEnv("MYSQL_USER", "root"),
		MySQLPassword:     getEnv("MYSQL_PASSWORD", "password"),
		MySQLDatabase:     getEnv("MYSQL_DATABASE", "law_oa_go"),

		PostgreSQLHost:     getEnv("POSTGRES_HOST", "localhost"),
		PostgreSQLPort:     getEnv("POSTGRES_PORT", "5432"),
		PostgreSQLUser:     getEnv("POSTGRES_USER", "law_oa_user"),
		PostgreSQLPassword: getEnv("POSTGRES_PASSWORD", "law_oa_password"),
		PostgreSQLDatabase: getEnv("POSTGRES_DATABASE", "law_oa_db"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func connectMySQL(config *Config) (*sql.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		config.MySQLUser, config.MySQLPassword, config.MySQLHost, config.MySQLPort, config.MySQLDatabase)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}

func connectPostgreSQL(config *Config) (*sql.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		config.PostgreSQLHost, config.PostgreSQLPort, config.PostgreSQLUser, config.PostgreSQLPassword, config.PostgreSQLDatabase)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}

func migrateData(mysqlDB, postgresDB *sql.DB) error {
	// 迁移用户数据
	if err := migrateUsers(mysqlDB, postgresDB); err != nil {
		return fmt.Errorf("迁移用户数据失败: %w", err)
	}

	// 迁移客户数据
	if err := migrateClients(mysqlDB, postgresDB); err != nil {
		return fmt.Errorf("迁移客户数据失败: %w", err)
	}

	// 迁移案件数据
	if err := migrateCases(mysqlDB, postgresDB); err != nil {
		return fmt.Errorf("迁移案件数据失败: %w", err)
	}

	return nil
}

func migrateUsers(mysqlDB, postgresDB *sql.DB) error {
	fmt.Println("📋 迁移用户数据...")

	// 查询MySQL用户数据
	rows, err := mysqlDB.Query(`
		SELECT id, created_at, updated_at, username, name, email, password, role, phone, avatar, status
		FROM users
		WHERE deleted_at IS NULL
		ORDER BY id
	`)
	if err != nil {
		return err
	}
	defer rows.Close()

	var users []MySQLUser
	for rows.Next() {
		var user MySQLUser
		err := rows.Scan(
			&user.ID, &user.CreatedAt, &user.UpdatedAt, &user.Username, &user.Name,
			&user.Email, &user.Password, &user.Role, &user.Phone, &user.Avatar, &user.Status,
		)
		if err != nil {
			return err
		}
		users = append(users, user)
	}

	// 插入PostgreSQL
	stmt, err := postgresDB.Prepare(`
		INSERT INTO users (id, username, name, email, password, role, phone, avatar, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT (id) DO UPDATE SET
			username = EXCLUDED.username,
			name = EXCLUDED.name,
			email = EXCLUDED.email,
			password = EXCLUDED.password,
			role = EXCLUDED.role,
			phone = EXCLUDED.phone,
			avatar = EXCLUDED.avatar,
			status = EXCLUDED.status,
			updated_at = EXCLUDED.updated_at
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, user := range users {
		_, err := stmt.Exec(
			user.ID, user.Username, user.Name, user.Email, user.Password,
			user.Role, user.Phone, user.Avatar, user.Status, user.CreatedAt, user.UpdatedAt,
		)
		if err != nil {
			return err
		}
	}

	fmt.Printf("✅ 成功迁移 %d 个用户\n", len(users))
	return nil
}

func migrateClients(mysqlDB, postgresDB *sql.DB) error {
	fmt.Println("📋 迁移客户数据...")

	// 查询MySQL客户数据
	rows, err := mysqlDB.Query(`
		SELECT id, created_at, updated_at, name, type, email, phone, address, company,
		       id_card, industry, contact_person, contact_phone, source, notes, status
		FROM clients
		WHERE deleted_at IS NULL
		ORDER BY id
	`)
	if err != nil {
		return err
	}
	defer rows.Close()

	var clients []MySQLClient
	for rows.Next() {
		var client MySQLClient
		err := rows.Scan(
			&client.ID, &client.CreatedAt, &client.UpdatedAt, &client.Name, &client.Type,
			&client.Email, &client.Phone, &client.Address, &client.Company,
			&client.IDCard, &client.Industry, &client.ContactPerson, &client.ContactPhone,
			&client.Source, &client.Notes, &client.Status,
		)
		if err != nil {
			return err
		}
		clients = append(clients, client)
	}

	// 插入PostgreSQL
	stmt, err := postgresDB.Prepare(`
		INSERT INTO clients (id, name, type, email, phone, address, company, id_card,
			industry, contact_person, contact_phone, source, notes, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
		ON CONFLICT (id) DO UPDATE SET
			name = EXCLUDED.name,
			type = EXCLUDED.type,
			email = EXCLUDED.email,
			phone = EXCLUDED.phone,
			address = EXCLUDED.address,
			company = EXCLUDED.company,
			id_card = EXCLUDED.id_card,
			industry = EXCLUDED.industry,
			contact_person = EXCLUDED.contact_person,
			contact_phone = EXCLUDED.contact_phone,
			source = EXCLUDED.source,
			notes = EXCLUDED.notes,
			status = EXCLUDED.status,
			updated_at = EXCLUDED.updated_at
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, client := range clients {
		_, err := stmt.Exec(
			client.ID, client.Name, client.Type, client.Email, client.Phone, client.Address,
			client.Company, client.IDCard, client.Industry, client.ContactPerson,
			client.ContactPhone, client.Source, client.Notes, client.Status,
			client.CreatedAt, client.UpdatedAt,
		)
		if err != nil {
			return err
		}
	}

	fmt.Printf("✅ 成功迁移 %d 个客户\n", len(clients))
	return nil
}

func migrateCases(mysqlDB, postgresDB *sql.DB) error {
	fmt.Println("📋 迁移案件数据...")

	// 查询MySQL案件数据
	rows, err := mysqlDB.Query(`
		SELECT id, created_at, updated_at, title, description, client_id, lawyer_id,
		       case_type, priority, status, start_date, end_date
		FROM cases
		WHERE deleted_at IS NULL
		ORDER BY id
	`)
	if err != nil {
		return err
	}
	defer rows.Close()

	var cases []MySQLCase
	for rows.Next() {
		var c MySQLCase
		err := rows.Scan(
			&c.ID, &c.CreatedAt, &c.UpdatedAt, &c.Title, &c.Description,
			&c.ClientID, &c.LawyerID, &c.CaseType, &c.Priority, &c.Status,
			&c.StartDate, &c.EndDate,
		)
		if err != nil {
			return err
		}
		cases = append(cases, c)
	}

	// 插入PostgreSQL
	stmt, err := postgresDB.Prepare(`
		INSERT INTO cases (id, title, description, client_id, lawyer_id, case_type, priority, status,
			start_date, end_date, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		ON CONFLICT (id) DO UPDATE SET
			title = EXCLUDED.title,
			description = EXCLUDED.description,
			client_id = EXCLUDED.client_id,
			lawyer_id = EXCLUDED.lawyer_id,
			case_type = EXCLUDED.case_type,
			priority = EXCLUDED.priority,
			status = EXCLUDED.status,
			start_date = EXCLUDED.start_date,
			end_date = EXCLUDED.end_date,
			updated_at = EXCLUDED.updated_at
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, c := range cases {
		_, err := stmt.Exec(
			c.ID, c.Title, c.Description, c.ClientID, c.LawyerID,
			c.CaseType, c.Priority, c.Status, c.StartDate, c.EndDate,
			c.CreatedAt, c.UpdatedAt,
		)
		if err != nil {
			return err
		}
	}

	fmt.Printf("✅ 成功迁移 %d 个案件\n", len(cases))
	return nil
}

// 辅助函数：检查字符串是否为空
func isEmpty(s string) bool {
	return strings.TrimSpace(s) == ""
}