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

    // 查询所有客户
    rows, err := db.Query("SELECT id, name, type FROM clients ORDER BY id")
    if err != nil {
        log.Printf("查询客户失败: %v", err)
        return
    }
    defer rows.Close()

    fmt.Println("所有客户列表:")
    for rows.Next() {
        var id uint
        var name, clientType string
        rows.Scan(&id, &name, &clientType)
        fmt.Printf("ID: %d, 名称: %s, 类型: %s\n", id, name, clientType)
    }
}