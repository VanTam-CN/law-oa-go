package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

func main() {
	// 数据库连接
	db, err := sql.Open("postgres", "host=localhost port=5432 user=law_oa_user password=1q2w#E$R dbname=law_oa_db sslmode=disable")
	if err != nil {
		log.Fatal("数据库连接失败:", err)
	}
	defer db.Close()

	// 检查客户表结构和数据
	fmt.Println("=== 检查客户表结构 ===")
	rows, err := db.Query("SELECT column_name, data_type, character_maximum_length FROM information_schema.columns WHERE table_name = 'clients' ORDER BY ordinal_position")
	if err != nil {
		log.Fatal("查询表结构失败:", err)
	}
	defer rows.Close()

	for rows.Next() {
		var columnName, dataType string
		var maxLength sql.NullInt64
		err := rows.Scan(&columnName, &dataType, &maxLength)
		if err != nil {
			log.Fatal("扫描表结构失败:", err)
		}
		if maxLength.Valid {
			fmt.Printf("列名: %s, 类型: %s, 最大长度: %d\n", columnName, dataType, maxLength.Int64)
		} else {
			fmt.Printf("列名: %s, 类型: %s\n", columnName, dataType)
		}
	}

	fmt.Println("\n=== 检查客户数据 ===")
	clientRows, err := db.Query("SELECT id, name, type FROM clients ORDER BY id LIMIT 10")
	if err != nil {
		log.Fatal("查询客户数据失败:", err)
	}
	defer clientRows.Close()

	for clientRows.Next() {
		var id int
		var name, clientType string
		err := clientRows.Scan(&id, &name, &clientType)
		if err != nil {
			log.Fatal("扫描客户数据失败:", err)
		}
		fmt.Printf("ID: %d, 名称: %s, 类型: '%s'\n", id, name, clientType)
	}

	fmt.Println("\n=== 检查案件数据中的客户类型 ===")
	caseRows, err := db.Query("SELECT c.id, c.title, cl.name as client_name, cl.type as client_type FROM cases c JOIN clients cl ON c.client_id = cl.id ORDER BY c.id LIMIT 5")
	if err != nil {
		log.Fatal("查询案件数据失败:", err)
	}
	defer caseRows.Close()

	for caseRows.Next() {
		var caseID int
		var caseTitle, clientName, clientType string
		err := caseRows.Scan(&caseID, &caseTitle, &clientName, &clientType)
		if err != nil {
			log.Fatal("扫描案件数据失败:", err)
		}
		fmt.Printf("案件ID: %d, 标题: %s, 客户: %s, 客户类型: '%s'\n", caseID, caseTitle, clientName, clientType)
	}
}