package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"law-oa-go/internal/config"
	"law-oa-go/internal/database"
)

func main() {
	var (
		_              = flag.String("config", "config.yaml", "配置文件路径")
		migrationsPath = flag.String("migrations", "./migrations", "迁移文件目录")
		command        = flag.String("command", "up", "迁移命令: bootstrap, up, down, steps, goto, version, force, drop, status")
		steps          = flag.Int("steps", 0, "迁移步数（用于steps命令）")
		version        = flag.Int("version", 0, "目标版本（用于goto和force命令）")
	)
	flag.Parse()

	if err := run(*command, *migrationsPath, *steps, *version); err != nil {
		log.Fatalf("迁移命令执行失败: %v", err)
	}
}

// run 承载命令完整流程；返回 error 交由 main 统一退出，
// 确保 defer 的连接/迁移器释放先于进程退出（gocritic exitAfterDefer 门禁）。
func run(command, migrationsPath string, steps, version int) error {
	// 加载配置
	// Bootstrap only needs the database connection. The application runtime
	// keeps the stricter JWT/OnlyOffice/production checks in config.Load().
	cfg, err := config.LoadForMigration()
	if err != nil {
		return fmt.Errorf("加载配置失败: %w", err)
	}
	if cfg.IsProduction() && command != "bootstrap" {
		return fmt.Errorf("生产环境只允许执行 -command bootstrap；历史迁移、回滚、drop 和状态命令必须在非生产环境或离线副本执行")
	}

	// Production uses an explicit, transactional PostgreSQL bootstrap. The
	// historical SQL directory contains mixed dialects and is intentionally not
	// used as the production installation path.
	if command == "bootstrap" {
		if driver := strings.ToLower(strings.TrimSpace(cfg.Database.Driver)); driver != "postgres" && driver != "postgresql" {
			return fmt.Errorf("生产 schema bootstrap 仅支持 PostgreSQL，当前驱动为 %q", cfg.Database.Driver)
		}
		db, bootstrapErr := database.InitWithConfig(cfg)
		if bootstrapErr != nil {
			return fmt.Errorf("初始化数据库连接失败: %w", bootstrapErr)
		}
		sqlDB, dbErr := db.DB()
		if dbErr != nil {
			return fmt.Errorf("获取数据库连接失败: %w", dbErr)
		}
		defer sqlDB.Close()
		if bootstrapErr = database.BootstrapProductionSchema(db); bootstrapErr != nil {
			return fmt.Errorf("生产 schema bootstrap 失败: %w", bootstrapErr)
		}
		fmt.Printf("生产 PostgreSQL schema bootstrap 完成，版本: %s\n", database.ProductionSchemaVersion)
		return nil
	}

	// 创建迁移器
	migrator, err := database.NewMigrator(&cfg.Database, migrationsPath)
	if err != nil {
		return fmt.Errorf("创建迁移器失败: %w", err)
	}
	defer migrator.Close()

	// 验证数据库
	if err := migrator.ValidateDatabase(); err != nil {
		return fmt.Errorf("数据库验证失败: %w", err)
	}

	// 执行命令
	switch command {
	case "up":
		err = migrator.Up()
	case "down":
		err = migrator.Down()
	case "steps":
		if steps == 0 {
			return fmt.Errorf("steps命令需要指定步数")
		}
		err = migrator.Steps(steps)
	case "goto":
		if version == 0 {
			return fmt.Errorf("goto命令需要指定版本号")
		}
		err = migrator.Goto(uint(version))
	case "version":
		var v uint
		var dirty bool
		v, dirty, err = migrator.Version()
		if err == nil {
			fmt.Printf("当前版本: %d\n", v)
			if dirty {
				fmt.Println("状态: 脏数据（迁移可能失败）")
			} else {
				fmt.Println("状态: 正常")
			}
		}
	case "force":
		if version < 0 {
			return fmt.Errorf("force命令需要指定有效的版本号")
		}
		err = migrator.Force(version)
	case "drop":
		fmt.Print("警告：此操作将删除所有数据库表！确认请输入 'yes': ")
		var confirm string
		fmt.Scanln(&confirm)
		if confirm != "yes" {
			fmt.Println("操作已取消")
			return nil
		}
		err = migrator.Drop()
	case "status":
		status, statusErr := migrator.Status()
		if statusErr != nil {
			return fmt.Errorf("获取状态失败: %w", statusErr)
		}
		fmt.Printf("数据库迁移状态:\n")
		fmt.Printf("  版本: %d\n", status.Version)
		fmt.Printf("  状态: %s\n", func() string {
			if status.Dirty {
				return "脏数据"
			}
			return "正常"
		}())

		// 显示已应用的迁移
		migrations, migrationsErr := migrator.GetAppliedMigrations()
		if migrationsErr != nil {
			log.Printf("获取迁移列表失败: %v", migrationsErr)
		} else {
			fmt.Printf("  已应用的迁移: %v\n", migrations)
		}
	case "create":
		if len(flag.Args()) == 0 {
			return fmt.Errorf("create命令需要指定迁移名称")
		}
		migrationName := flag.Args()[0]
		err = createMigrationFiles(migrationsPath, migrationName)
	default:
		return fmt.Errorf("未知命令: %s（可用命令: bootstrap, up, down, steps, goto, version, force, drop, status, create）", command)
	}

	if err != nil {
		return fmt.Errorf("执行命令失败: %w", err)
	}

	fmt.Println("命令执行成功")
	return nil
}

// createMigrationFiles 创建迁移文件
func createMigrationFiles(migrationsPath, name string) error {
	// 获取下一个版本号
	files, err := os.ReadDir(migrationsPath)
	if err != nil {
		return fmt.Errorf("读取迁移目录失败: %w", err)
	}

	maxVersion := 0
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		fileName := file.Name()
		if len(fileName) >= 6 {
			versionStr := fileName[:6]
			if version, err := strconv.Atoi(versionStr); err == nil && version > maxVersion {
				maxVersion = version
			}
		}
	}

	nextVersion := maxVersion + 1
	versionStr := fmt.Sprintf("%06d", nextVersion)

	// 创建up文件
	upFile := fmt.Sprintf("%s/%s_%s.up.sql", migrationsPath, versionStr, name)
	upContent := fmt.Sprintf("-- %s up migration\n\n-- Add your SQL statements here\n", name)

	if err := os.WriteFile(upFile, []byte(upContent), 0o644); err != nil {
		return fmt.Errorf("创建up文件失败: %w", err)
	}

	// 创建down文件
	downFile := fmt.Sprintf("%s/%s_%s.down.sql", migrationsPath, versionStr, name)
	downContent := fmt.Sprintf("-- %s down migration\n\n-- Add your rollback SQL statements here\n", name)

	if err := os.WriteFile(downFile, []byte(downContent), 0o644); err != nil {
		return fmt.Errorf("创建down文件失败: %w", err)
	}

	fmt.Printf("迁移文件已创建:\n")
	fmt.Printf("  %s\n", upFile)
	fmt.Printf("  %s\n", downFile)

	return nil
}
