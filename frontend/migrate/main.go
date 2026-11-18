package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	var (
		command       = flag.String("command", "up", "Migration command (up, down, force, version)")
		version       = flag.Int("version", 0, "Target version for force command")
		dsn           = flag.String("dsn", "", "Database DSN")
		migrationsDir = flag.String("dir", "./migrations", "Migrations directory")
	)
	flag.Parse()

	if *dsn == "" {
		// 从环境变量或配置文件读取DSN
		*dsn = os.Getenv("DB_DSN")
		if *dsn == "" {
			// 构建默认DSN
			host := getEnv("DB_HOST", "localhost")
			port := getEnv("DB_PORT", "3306")
			user := getEnv("DB_USER", "root")
			password := getEnv("DB_PASSWORD", "password")
			dbname := getEnv("DB_NAME", "law_oa")

			*dsn = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
				user, password, host, port, dbname)
		}
	}

	// 确保迁移目录存在
	if _, err := os.Stat(*migrationsDir); os.IsNotExist(err) {
		log.Printf("创建迁移目录: %s", *migrationsDir)
		if err := os.MkdirAll(*migrationsDir, 0750); err != nil {
			log.Fatalf("创建迁移目录失败: %v", err)
		}
	}

	// 执行迁移命令
	switch *command {
	case "up":
		if err := migrateUp(*dsn, *migrationsDir); err != nil {
			log.Fatalf("向上迁移失败: %v", err)
		}
	case "down":
		if err := migrateDown(*dsn, *migrationsDir); err != nil {
			log.Fatalf("向下迁移失败: %v", err)
		}
	case "force":
		if *version == 0 {
			log.Fatal("force 命令需要指定版本号")
		}
		if err := migrateForce(*dsn, *migrationsDir, *version); err != nil {
			log.Fatalf("强制迁移失败: %v", err)
		}
	case "version":
		if err := showVersion(*dsn, *migrationsDir); err != nil {
			log.Fatalf("获取版本失败: %v", err)
		}
	case "create":
		if len(flag.Args()) == 0 {
			log.Fatal("create 命令需要指定迁移名称")
		}
		name := flag.Args()[0]
		if err := createMigration(*migrationsDir, name); err != nil {
			log.Fatalf("创建迁移文件失败: %v", err)
		}
	default:
		log.Fatalf("未知命令: %s", *command)
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func migrateUp(dsn, migrationsDir string) error {
	log.Println("开始向上迁移...")

	// 检查数据库连接
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("连接数据库失败: %w", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		return fmt.Errorf("数据库连接测试失败: %w", err)
	}

	// 如果迁移目录为空，先运行初始化脚本
	if isEmpty, _ := isDirEmpty(migrationsDir); isEmpty {
		log.Println("检测到空迁移目录，运行初始化脚本...")
		if err := runInitScript(dsn); err != nil {
			return fmt.Errorf("运行初始化脚本失败: %w", err)
		}
	}

	// 创建迁移实例
	driver, err := mysql.WithInstance(db, &mysql.Config{})
	if err != nil {
		return fmt.Errorf("创建数据库驱动失败: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", migrationsDir),
		"mysql",
		driver,
	)
	if err != nil {
		return fmt.Errorf("创建迁移实例失败: %w", err)
	}
	defer m.Close()

	// 执行迁移
	if err := m.Up(); err != nil {
		if err == migrate.ErrNoChange {
			log.Println("数据库已经是最新版本")
			return nil
		}
		return fmt.Errorf("迁移失败: %w", err)
	}

	log.Println("向上迁移完成")
	return nil
}

func migrateDown(dsn, migrationsDir string) error {
	log.Println("开始向下迁移...")

	// 创建迁移实例
	m, err := getMigrateInstance(dsn, migrationsDir)
	if err != nil {
		return err
	}
	defer m.Close()

	// 执行向下迁移
	if err := m.Down(); err != nil {
		if err == migrate.ErrNoChange {
			log.Println("数据库已经是初始版本")
			return nil
		}
		return fmt.Errorf("向下迁移失败: %w", err)
	}

	log.Println("向下迁移完成")
	return nil
}

func migrateForce(dsn, migrationsDir string, version int) error {
	log.Printf("强制迁移到版本: %d", version)

	// 创建迁移实例
	m, err := getMigrateInstance(dsn, migrationsDir)
	if err != nil {
		return err
	}
	defer m.Close()

	// 执行强制迁移
	if err := m.Force(version); err != nil {
		return fmt.Errorf("强制迁移失败: %w", err)
	}

	log.Printf("强制迁移到版本 %d 完成", version)
	return nil
}

func showVersion(dsn, migrationsDir string) error {
	// 创建迁移实例
	m, err := getMigrateInstance(dsn, migrationsDir)
	if err != nil {
		return err
	}
	defer m.Close()

	// 获取当前版本
	version, dirty, err := m.Version()
	if err != nil {
		return fmt.Errorf("获取版本失败: %w", err)
	}

	log.Printf("当前版本: %d, Dirty: %t", version, dirty)
	return nil
}

func createMigration(migrationsDir, name string) error {
	// 生成时间戳
	timestamp := fmt.Sprintf("%d", time.Now().Unix())

	// 创建向上迁移文件
	upFile := filepath.Join(migrationsDir, fmt.Sprintf("%s_%s.up.sql", timestamp, name))
	if err := os.WriteFile(upFile, []byte("-- Up migration\n"), 0644); err != nil {
		return fmt.Errorf("创建向上迁移文件失败: %w", err)
	}

	// 创建向下迁移文件
	downFile := filepath.Join(migrationsDir, fmt.Sprintf("%s_%s.down.sql", timestamp, name))
	if err := os.WriteFile(downFile, []byte("-- Down migration\n"), 0644); err != nil {
		return fmt.Errorf("创建向下迁移文件失败: %w", err)
	}

	log.Printf("创建迁移文件成功:\n%s\n%s", upFile, downFile)
	return nil
}

func getMigrateInstance(dsn, migrationsDir string) (*migrate.Migrate, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("连接数据库失败: %w", err)
	}

	driver, err := mysql.WithInstance(db, &mysql.Config{})
	if err != nil {
		return nil, fmt.Errorf("创建数据库驱动失败: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", migrationsDir),
		"mysql",
		driver,
	)
	if err != nil {
		return nil, fmt.Errorf("创建迁移实例失败: %w", err)
	}

	return m, nil
}

func isDirEmpty(dir string) (bool, error) {
	f, err := os.Open(dir)
	if err != nil {
		return false, err
	}
	defer f.Close()

	_, err = f.Readdirnames(1)
	if err == nil {
		return false, nil
	}
	if err.Error() == "EOF" {
		return true, nil
	}
	return false, err
}

func runInitScript(dsn string) error {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return err
	}
	defer db.Close()

	// 读取初始化脚本
	scriptPath := "./scripts/init.sql"
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		log.Printf("初始化脚本不存在: %s", scriptPath)
		return nil
	}

	script, err := os.ReadFile(scriptPath)
	if err != nil {
		return fmt.Errorf("读取初始化脚本失败: %w", err)
	}

	// 执行初始化脚本
	if _, err := db.Exec(string(script)); err != nil {
		return fmt.Errorf("执行初始化脚本失败: %w", err)
	}

	log.Println("初始化脚本执行完成")
	return nil
}
