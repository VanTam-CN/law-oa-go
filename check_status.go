package main

import (
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Case struct {
	ID     uint   `gorm:"primaryKey"`
	Title  string
	Status string
}

func main() {
	dsn := "host=localhost port=5432 user=law_oa_user password=1q2w#E$R dbname=law_oa_db sslmode=disable TimeZone=Asia/Shanghai"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("连接数据库失败:", err)
	}

	var cases []Case
	db.Select("DISTINCT status").Find(&cases)

	fmt.Println("数据库中已有的案件状态:")
	for _, c := range cases {
		fmt.Printf("- %s\n", c.Status)
	}
}