package main

import (
    "database/sql"
    "fmt"
    "log"

    _ "github.com/lib/pq"
)

func main() {
    // 数据库连接
    db, err := sql.Open("postgres", "host=localhost port=5432 user=law_oa_user password=law_oa_password dbname=law_oa_db sslmode=disable")
    if err != nil {
        log.Fatal("数据库连接失败:", err)
    }
    defer db.Close()

    // 查询张伟律师信息
    var lawyerID uint
    var username, name string
    err = db.QueryRow("SELECT id, username, name FROM users WHERE username = 'zhangwei' LIMIT 1").Scan(&lawyerID, &username, &name)
    if err != nil {
        log.Printf("查询张伟律师失败: %v", err)
        return
    }

    fmt.Printf("张伟律师信息: ID=%d, 用户名=%s, 姓名=%s\n", lawyerID, username, name)

    // 查询他代理的案件
    rows, err := db.Query(`
        SELECT c.id, c.title, cl.name as client_name
        FROM cases c
        JOIN clients cl ON c.client_id = cl.id
        WHERE c.lawyer_id = $1`, lawyerID)
    if err != nil {
        log.Printf("查询案件失败: %v", err)
        return
    }
    defer rows.Close()

    fmt.Println("张伟律师代理的案件:")
    for rows.Next() {
        var caseID uint
        var title, clientName string
        rows.Scan(&caseID, &title, &clientName)
        fmt.Printf("  - 案件ID: %d, 标题: %s, 客户: %s\n", caseID, title, clientName)
    }

    // 查询美团客户信息
    var meituanID uint
    var meituanName string
    err = db.QueryRow("SELECT id, name FROM clients WHERE name LIKE '%美团%' LIMIT 1").Scan(&meituanID, &meituanName)
    if err != nil {
        log.Printf("查询美团客户失败: %v", err)
        return
    }
    fmt.Printf("美团客户信息: ID=%d, 名称=%s\n", meituanID, meituanName)

    // 测试冲突检测API数据
    fmt.Println("\n测试冲突检测所需的数据:")
    fmt.Printf("律师ID: %d\n", lawyerID)
    fmt.Printf("客户ID: %d\n", meituanID)
    fmt.Println("案件名称: 测试利益冲突案件")
    fmt.Println("案件类型: commercial")
    fmt.Println("客户类型: COMPANY")
    fmt.Println("对方当事人: 阿里巴巴集团控股有限公司")
}