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

    // 查询admin用户
    var id uint
    var username, email string
    var name sql.NullString
    err = db.QueryRow("SELECT id, username, email, name FROM users WHERE username = 'admin' LIMIT 1").Scan(&id, &username, &email, &name)
    if err != nil {
        log.Printf("查询admin用户失败: %v", err)
        return
    }

    displayName := "未设置"
    if name.Valid {
        displayName = name.String
    }
    fmt.Printf("Admin用户信息: ID=%d, 用户名=%s, 邮箱=%s, 姓名=%s\n", id, username, email, displayName)
}