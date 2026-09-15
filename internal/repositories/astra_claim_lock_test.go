package repositories

import (
 "context"
 "strings"
 "testing"

 "gorm.io/driver/postgres"
 "gorm.io/gorm"
)

func TestAstraPostgresClaimEmitsRowLock(t *testing.T) {
 db, err := gorm.Open(postgres.Open("host=127.0.0.1 user=qa dbname=qa sslmode=disable"), &gorm.Config{DryRun:true, DisableAutomaticPing:true, SkipDefaultTransaction:true})
 if err != nil {t.Fatal(err)}
 var query string
 db.Callback().Query().After("gorm:query").Register("astra:capture",func(tx *gorm.DB){query=tx.Statement.SQL.String()})
 r := &IntegrationRepository{db:db}
 _, err = r.ClaimCaseCreationOnTx(context.Background(),db,"APR-QA","qa")
 if err != nil {t.Fatal(err)}
 t.Log(query)
 if !strings.Contains(query,"FOR UPDATE") {t.Fatal("PostgreSQL approval read lacks the claimed FOR UPDATE lock")}
}
