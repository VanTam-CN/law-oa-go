//go:build ignore

package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

const DSN = "host=localhost port=5432 user=law_oa_user password=law_oa_password dbname=law_oa_db sslmode=disable"

func main() {
	db, err := sql.Open("postgres", DSN)
	if err != nil {
		log.Fatal("连接数据库失败:", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal("数据库连接测试失败:", err)
	}

	fmt.Println("✅ 数据库连接成功")
	fmt.Println("🔍 查询现有用户...")

	// 查询所有用户
	rows, err := db.Query("SELECT id, username, name, email, role, phone, status FROM users ORDER BY id")
	if err != nil {
		log.Fatal("查询用户失败:", err)
	}
	defer rows.Close()

	var users []struct {
		ID       int
		Username string
		Name     string
		Email    string
		Role     string
		Phone    string
		Status   string
	}

	fmt.Println("\n📋 数据库中的用户:")
	fmt.Println("ID\t用户名\t姓名\t邮箱\t角色\t手机\t状态")
	fmt.Println("-------------------------------------------------------------------")

	hasUsers := false
	for rows.Next() {
		var user struct {
			ID       int
			Username string
			Name     string
			Email    string
			Role     string
			Phone    string
			Status   string
		}

		err := rows.Scan(&user.ID, &user.Username, &user.Name, &user.Email, &user.Role, &user.Phone, &user.Status)
		if err != nil {
			log.Printf("扫描用户数据失败: %v", err)
			continue
		}

		hasUsers = true
		fmt.Printf("%d\t%s\t%s\t%s\t%s\t%s\t%s\n", user.ID, user.Username, user.Name, user.Email, user.Role, user.Phone, user.Status)
		users = append(users, user)
	}

	if !hasUsers {
		fmt.Println("❌ 数据库中没有找到任何用户")
	} else {
		fmt.Printf("\n✅ 找到 %d 个用户\n", len(users))
	}

	// 查询审批相关数据
	fmt.Println("\n📋 查询审批数据...")

	var approvalCount, recordCount int
	db.QueryRow("SELECT COUNT(*) FROM approval_requests").Scan(&approvalCount)
	db.QueryRow("SELECT COUNT(*) FROM approval_records").Scan(&recordCount)

	fmt.Printf("审批申请数量: %d\n", approvalCount)
	fmt.Printf("审批记录数量: %d\n", recordCount)

	// 如果有用户，查看每个用户的审批统计
	if len(users) > 0 {
		fmt.Println("\n📊 用户审批统计:")
		for _, user := range users {
			var myApprovals, pendingApprovals int
			db.QueryRow("SELECT COUNT(*) FROM approval_requests WHERE applicant_id = $1", user.ID).Scan(&myApprovals)
			db.QueryRow("SELECT COUNT(*) FROM approval_requests WHERE current_approver_id = $1", user.ID).Scan(&pendingApprovals)
			fmt.Printf("用户 %s (ID:%d): 我的申请=%d, 待我审批=%d\n", user.Name, user.ID, myApprovals, pendingApprovals)
		}
	}
}