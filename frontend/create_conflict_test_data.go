package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"
)

func main() {
	// PostgreSQL数据库连接
	db, err := sql.Open("postgres", "host=localhost port=5432 user=law_oa_user password=law_oa_password dbname=law_oa_db sslmode=disable")
	if err != nil {
		log.Fatal("数据库连接失败:", err)
	}
	defer db.Close()

	// 测试连接
	if err := db.Ping(); err != nil {
		log.Fatal("数据库连接测试失败:", err)
	}

	fmt.Println("=== 创建冲突检测测试数据 ===")

	// 1. 确保张伟律师存在
	var lawyerID int
	err = db.QueryRow("SELECT id FROM lawyers WHERE username = 'zhangwei'").Scan(&lawyerID)
	if err != nil {
		if err == sql.ErrNoRows {
			fmt.Println("张伟律师不存在，先创建...")
			err = db.QueryRow(`
				INSERT INTO lawyers (name, username, phone, email, position, department, specialty, created_at, updated_at)
				VALUES ('张伟', 'zhangwei', '13800138000', 'zhangwei@lawfirm.com', '律师', '诉讼部', '商业诉讼', NOW(), NOW())
				RETURNING id
			`).Scan(&lawyerID)
			if err != nil {
				log.Fatal("创建张伟律师失败:", err)
			}
		} else {
			log.Fatal("查询张伟律师失败:", err)
		}
	}

	fmt.Printf("张伟律师ID: %d\n", lawyerID)

	// 2. 创建阿里巴巴客户
	var alibabaClientID int
	err = db.QueryRow("SELECT id FROM clients WHERE name = '阿里巴巴集团' OR company = '阿里巴巴集团'").Scan(&alibabaClientID)
	if err != nil {
		if err == sql.ErrNoRows {
			fmt.Println("创建阿里巴巴客户...")
			err = db.QueryRow(`
				INSERT INTO clients (name, company, phone, email, client_type, address, created_at, updated_at)
				VALUES ('阿里巴巴集团', '阿里巴巴集团', '400-800-1688', 'legal@alibaba.com', '企业', '杭州市余杭区文一西路969号', NOW(), NOW())
				RETURNING id
			`).Scan(&alibabaClientID)
			if err != nil {
				log.Fatal("创建阿里巴巴客户失败:", err)
			}
		} else {
			log.Fatal("查询阿里巴巴客户失败:", err)
		}
	}

	fmt.Printf("阿里巴巴客户ID: %d\n", alibabaClientID)

	// 3. 创建阿里巴巴相关的案件（张伟律师代理）
	// 案件1: 阿里巴巴诉字节跳动不正当竞争纠纷案
	var case1ID int
	err = db.QueryRow("SELECT id FROM cases WHERE case_name = '阿里巴巴诉字节跳动不正当竞争纠纷案'").Scan(&case1ID)
	if err != nil {
		if err == sql.ErrNoRows {
			fmt.Println("创建阿里巴巴诉字节跳动案件...")
			err = db.QueryRow(`
				INSERT INTO cases (case_name, case_no, case_type, client_id, lawyer_id, opposing_party, status, description, created_at, updated_at)
				VALUES ('阿里巴巴诉字节跳动不正当竞争纠纷案', 'ZC2023-001', 'commercial', $1, $2, '字节跳动科技有限公司', 'in_progress', '阿里巴巴诉字节跳动不正当竞争案件，涉及电商业务竞争', NOW(), NOW())
				RETURNING id
			`, alibabaClientID, lawyerID).Scan(&case1ID)
			if err != nil {
				log.Fatal("创建案件1失败:", err)
			}
		} else {
			log.Fatal("查询案件1失败:", err)
		}
	}

	fmt.Printf("案件1 ID: %d\n", case1ID)

	fmt.Printf("\n=== 测试数据创建完成 ===\n")
	fmt.Printf("测试时间: %s\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Println("\n🎯 测试场景:")
	fmt.Println("1. 使用张伟律师账号登录 (zhangwei / law123456)")
	fmt.Println("2. 尝试为字节跳动创建新案件，对方填写 '阿里巴巴'")
	fmt.Println("3. 系统应该检测到商业竞争冲突（张伟已代理阿里巴巴相关案件）")
}
