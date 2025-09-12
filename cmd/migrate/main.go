package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"

	"law-oa-go/internal/config"
	"law-oa-go/internal/database"
)

func main() {
	var (
		_              = flag.String("config", "config.yaml", "配置文件路径")
		migrationsPath = flag.String("migrations", "./migrations", "迁移文件目录")
		command        = flag.String("command", "up", "迁移命令: up, down, steps, goto, version, force, drop, status")
		steps          = flag.Int("steps", 0, "迁移步数（用于steps命令）")
		version        = flag.Int("version", 0, "目标版本（用于goto和force命令）")
	)
	flag.Parse()

	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 创建迁移器
	migrator, err := database.NewMigrator(&cfg.Database, *migrationsPath)
	if err != nil {
		log.Fatalf("创建迁移器失败: %v", err)
	}
	defer migrator.Close()

	// 验证数据库
	if err := migrator.ValidateDatabase(); err != nil {
		log.Fatalf("数据库验证失败: %v", err)
	}

	// 执行命令
	switch *command {
	case "up":
		err = migrator.Up()
	case "down":
		err = migrator.Down()
	case "steps":
		if *steps == 0 {
			log.Fatal("steps命令需要指定步数")
		}
		err = migrator.Steps(*steps)
	case "goto":
		if *version == 0 {
			log.Fatal("goto命令需要指定版本号")
		}
		err = migrator.Goto(uint(*version))
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
		if *version < 0 {
			log.Fatal("force命令需要指定有效的版本号")
		}
		err = migrator.Force(*version)
	case "drop":
		fmt.Print("警告：此操作将删除所有数据库表！确认请输入 'yes': ")
		var confirm string
		fmt.Scanln(&confirm)
		if confirm != "yes" {
			fmt.Println("操作已取消")
			return
		}
		err = migrator.Drop()
	case "status":
		status, statusErr := migrator.Status()
		if statusErr != nil {
			log.Fatalf("获取状态失败: %v", statusErr)
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
			log.Fatal("create命令需要指定迁移名称")
		}
		migrationName := flag.Args()[0]
		err = createMigrationFiles(*migrationsPath, migrationName)
	default:
		fmt.Printf("未知命令: %s\n", *command)
		fmt.Println("可用命令: up, down, steps, goto, version, force, drop, status, create")
		os.Exit(1)
	}

	if err != nil {
		log.Fatalf("执行命令失败: %v", err)
	}

	fmt.Println("命令执行成功")
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

	if err := os.WriteFile(upFile, []byte(upContent), 0644); err != nil {
		return fmt.Errorf("创建up文件失败: %w", err)
	}

	// 创建down文件
	downFile := fmt.Sprintf("%s/%s_%s.down.sql", migrationsPath, versionStr, name)
	downContent := fmt.Sprintf("-- %s down migration\n\n-- Add your rollback SQL statements here\n", name)

	if err := os.WriteFile(downFile, []byte(downContent), 0644); err != nil {
		return fmt.Errorf("创建down文件失败: %w", err)
	}

	fmt.Printf("迁移文件已创建:\n")
	fmt.Printf("  %s\n", upFile)
	fmt.Printf("  %s\n", downFile)

	return nil
}
