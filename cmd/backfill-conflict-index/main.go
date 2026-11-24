package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"strings"

	"law-oa-go/internal/config"
	"law-oa-go/internal/database"
	"law-oa-go/internal/repositories"
)

// This command is the reviewed operator path for historical conflict-index
// backfill. It is read-only by default; --apply is required to write index
// rows and durable reconciliation evidence. Logs contain counts and hashes,
// never names, case numbers, identifiers or evidence payloads.
func main() {
	apply := flag.Bool("apply", false, "确认写入历史主体索引和对账运行记录；默认只读盘点")
	actorID := flag.Uint("actor-id", 0, "执行人用户 ID；用于审计，可选")
	evidenceReference := flag.String("evidence-reference", "", "律所档案核对凭证引用；--apply 时必填")
	flag.Parse()

	if err := run(*apply, *actorID, *evidenceReference); err != nil {
		log.Fatalf("历史主体索引回填失败: %v", err)
	}
}

// run 承载命令完整流程；返回 error 交由 main 统一退出，确保 defer 的连接释放先于进程退出。
func run(apply bool, actorID uint, evidenceReference string) error {
	if apply && strings.TrimSpace(evidenceReference) == "" {
		return fmt.Errorf("--apply 必须同时提供 --evidence-reference")
	}
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("加载配置失败: %w", err)
	}
	db, err := database.InitWithConfig(cfg)
	if err != nil {
		return fmt.Errorf("连接数据库失败: %w", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("获取数据库连接失败: %w", err)
	}
	defer sqlDB.Close()

	repo := repositories.NewConflictRepository(db, nil)
	reconciler, ok := repo.(repositories.ConflictP0SubjectIndexReconciler)
	if !ok {
		return fmt.Errorf("当前冲突仓储未提供历史主体索引对账能力")
	}
	if apply {
		log.Printf("开始写入历史主体索引对账；actor_id=%d，日志不会输出敏感业务数据", actorID)
	} else {
		log.Println("当前为只读盘点模式；如需写入请显式传入 --apply")
	}

	runs, err := reconciler.ReconcileConflictSubjectIndex(context.Background(), actorID, evidenceReference, apply)
	if err != nil {
		return fmt.Errorf("历史主体索引对账失败: %w", err)
	}
	for _, entry := range runs {
		log.Printf("scope=%s status=%s source_records=%d indexed_records=%d missing_records=%d source_version=%s reconciliation_hash=%s run_id=%s",
			entry.ScopeType, entry.Status, entry.SourceRecordCount, entry.IndexedRecordCount, entry.MissingRecordCount,
			entry.SourceVersion, entry.ReconciliationHash, entry.ID)
	}
	if apply {
		log.Println("历史主体索引回填和对账完成；请由独立冲突核查人将每个 run_id 绑定到对应档案覆盖 scope，再执行 readiness 检查")
	}
	return nil
}
