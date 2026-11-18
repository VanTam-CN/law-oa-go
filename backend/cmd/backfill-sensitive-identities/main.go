package main

import (
	"flag"
	"fmt"
	"log"
	"strings"

	"law-oa-go/internal/config"
	"law-oa-go/internal/database"
	"law-oa-go/internal/security"

	"gorm.io/gorm"
)

// This command is intentionally separate from application startup. It first
// supports a dry run, and writes only after the operator explicitly passes
// --apply. No identity value is ever printed.
func main() {
	apply := flag.Bool("apply", false, "确认执行加密回填；默认只盘点")
	flag.Parse()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}
	if _, err := security.SubjectDataKey(); err != nil {
		log.Fatalf("身份信息密钥不可用: %v", err)
	}
	db, err := database.InitWithConfig(cfg)
	if err != nil {
		log.Fatalf("连接数据库失败: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("获取数据库连接失败: %v", err)
	}
	defer sqlDB.Close()

	if *apply {
		log.Println("开始执行身份信息加密回填；日志不会输出身份原文")
	} else {
		log.Println("当前为只读盘点模式；如需写入请显式传入 --apply")
	}
	entityCount, clientCount, err := backfill(db, *apply)
	if err != nil {
		log.Fatalf("身份信息回填失败: %v", err)
	}
	log.Printf("盘点完成: 主体记录=%d, 客户记录=%d", entityCount, clientCount)

	var entityPlaintext, clientPlaintext int64
	if err := db.Raw("SELECT COUNT(*) FROM entities WHERE identity_number IS NOT NULL AND TRIM(identity_number) <> ''").Scan(&entityPlaintext).Error; err != nil {
		log.Fatalf("复核主体明文失败: %v", err)
	}
	if err := db.Raw("SELECT COUNT(*) FROM clients WHERE id_card IS NOT NULL AND TRIM(id_card) <> ''").Scan(&clientPlaintext).Error; err != nil {
		log.Fatalf("复核客户明文失败: %v", err)
	}
	log.Printf("回填后复核: 主体明文=%d, 客户明文=%d", entityPlaintext, clientPlaintext)
	if *apply && (entityPlaintext != 0 || clientPlaintext != 0) {
		log.Fatal("回填后仍存在明文身份字段，生产门禁不能通过")
	}
}

func backfill(db *gorm.DB, apply bool) (int64, int64, error) {
	entities, err := backfillTable(db, "entities", "identity_number", "identity_number_digest", "identity_number_ciphertext", apply)
	if err != nil {
		return 0, 0, err
	}
	clients, err := backfillTable(db, "clients", "id_card", "id_card_digest", "id_card_ciphertext", apply)
	if err != nil {
		return entities, 0, err
	}
	return entities, clients, nil
}

func backfillTable(db *gorm.DB, table, plaintextColumn, digestColumn, ciphertextColumn string, apply bool) (int64, error) {
	query := fmt.Sprintf("SELECT id, %s FROM %s WHERE %s IS NOT NULL AND TRIM(%s) <> ''", plaintextColumn, table, plaintextColumn, plaintextColumn)
	rows, err := db.Raw(query).Rows()
	if err != nil {
		return 0, fmt.Errorf("读取%s明文记录失败: %w", table, err)
	}
	defer rows.Close()

	var count int64
	for rows.Next() {
		var id uint
		var plaintext string
		if err := rows.Scan(&id, &plaintext); err != nil {
			return count, fmt.Errorf("读取%s记录失败: %w", table, err)
		}
		ciphertext, digest, err := security.ProtectIdentityNumber(strings.TrimSpace(plaintext))
		if err != nil {
			return count, fmt.Errorf("生成%s保护值失败: %w", table, err)
		}
		count++
		if !apply {
			continue
		}
		update := fmt.Sprintf("UPDATE %s SET %s = '', %s = ?, %s = ? WHERE id = ? AND %s = ?", table, plaintextColumn, digestColumn, ciphertextColumn, plaintextColumn)
		result := db.Exec(update, digest, ciphertext, id, plaintext)
		if result.Error != nil {
			return count, fmt.Errorf("更新%s记录失败: %w", table, result.Error)
		}
		if result.RowsAffected != 1 {
			return count, fmt.Errorf("更新%s记录时检测到并发变化，已停止回填", table)
		}
	}
	if err := rows.Err(); err != nil {
		return count, fmt.Errorf("遍历%s记录失败: %w", table, err)
	}
	return count, nil
}
