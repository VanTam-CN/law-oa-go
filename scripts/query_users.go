//go:build ignore

package main

import (
    "database/sql"
    "fmt"
    "log"
    _ "github.com/lib/pq"
)

func main() {
    db, err := sql.Open("postgres", "host=localhost port=5432 user=law_oa_user password=law_oa_password dbname=law_oa_db sslmode=disable")
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    rows, err := db.Query("SELECT id, email, name, role FROM users LIMIT 5")
    if err != nil {
        log.Fatal(err)
    }
    defer rows.Close()

    fmt.Println("ID\tEmail\t\tName\tRole")
    for rows.Next() {
        var id int
        var email, name, role string
        rows.Scan(&id, &email, &name, &role)
        fmt.Printf("%d\t%s\t%s\t%s\n", id, email, name, role)
    }
}
