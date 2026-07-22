// qa-fixture provisions fictional, repeatable acceptance data for a non-production
// PostgreSQL database. It is intentionally a command, not a production migration:
// test identities must never be created by application startup or schema bootstrap.
package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"law-oa-go/internal/config"
	"law-oa-go/internal/database"
	"law-oa-go/internal/models"
	"law-oa-go/internal/repositories"
	"law-oa-go/internal/services"
)

const (
	fixtureName         = "conflict-p0"
	fixtureSource       = "qa_conflict_p0_fixture_v1"
	fixtureConfirmation = "I_UNDERSTAND"
	qaPasswordEnv       = "QA_PASSWORD"

	lawyerAEmail = "qa.lawyer.a@qa.invalid"
	lawyerBEmail = "qa.lawyer.b@qa.invalid"
	officerEmail = "qa.conflict.officer@qa.invalid"
	clientAEmail = "qa.p0.client.a@qa.invalid"
	clientBEmail = "qa.p0.client.b@qa.invalid"
	caseANumber  = "QA-P0-A-2026-001"
	caseBNumber  = "QA-P0-B-2026-001"
	checkID      = "QA-P0-A-CHECK-20260719"
	approvalID   = "QA-P0-A-APPROVAL-20260719"
	approvalNo   = "APR-QA-P0-20260719"
	assignmentID = "QA-P0-A-ASSIGNMENT-20260719"
)

type userSpec struct {
	Username   string
	Name       string
	Email      string
	Role       string
	Department string
	Seniority  string
}

type clientSpec struct {
	Name    string
	Email   string
	Company string
	Type    string
	Source  string
	Notes   string
}

func main() {
	mode := flag.String("mode", "seed", "操作模式: seed 或 verify")
	flag.Parse()

	if *mode != "seed" && *mode != "verify" {
		fatal("不支持的 mode %q，必须是 seed 或 verify", *mode)
	}
	if err := requireNonProductionEnvironment(); err != nil {
		fatal("QA 夹具安全门禁失败: %v", err)
	}
	if strings.TrimSpace(os.Getenv("QA_FIXTURE_CONFIRM")) != fixtureConfirmation {
		fatal("必须设置 QA_FIXTURE_CONFIRM=%s；该确认用于防止误写入非 QA 数据库", fixtureConfirmation)
	}

	cfg, err := config.LoadForMigration()
	if err != nil {
		fatal("加载数据库配置失败: %v", err)
	}
	if cfg.IsProduction() {
		fatal("production 环境禁止运行 QA 夹具")
	}
	if !strings.EqualFold(cfg.Database.Driver, "postgres") && !strings.EqualFold(cfg.Database.Driver, "postgresql") {
		fatal("QA 夹具只支持 PostgreSQL，当前驱动为 %q", cfg.Database.Driver)
	}

	db, err := database.InitWithConfig(cfg)
	if err != nil {
		fatal("连接 QA PostgreSQL 失败: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		fatal("获取数据库连接失败: %v", err)
	}
	defer sqlDB.Close()

	if err := requireFixtureTables(db); err != nil {
		fatal("QA 夹具依赖未满足: %v", err)
	}

	ctx := context.Background()
	if *mode == "verify" {
		if err := verifyFixture(ctx, db); err != nil {
			fatal("QA 夹具核验失败: %v", err)
		}
		fmt.Printf("QA 夹具核验通过: %s (%s)\n", fixtureName, cfg.Database.Database)
		return
	}

	password := strings.TrimSpace(os.Getenv(qaPasswordEnv))
	if len(password) < 12 {
		fatal("%s 必须设置且至少 12 个字符；密码不会写入仓库或输出", qaPasswordEnv)
	}
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		fatal("生成 QA 密码哈希失败: %v", err)
	}
	if err := syncPostgreSQLSequences(db); err != nil {
		fatal("同步 QA 数据库自增序列失败: %v", err)
	}

	result, err := seedFixture(ctx, db, string(passwordHash))
	if err != nil {
		fatal("写入 QA 夹具失败: %v", err)
	}
	if err := verifyFixture(ctx, db); err != nil {
		fatal("写入后核验 QA 夹具失败: %v", err)
	}

	fmt.Printf("QA 夹具写入并核验通过: %s (%s)\n", fixtureName, cfg.Database.Database)
	fmt.Printf("律师 A: %s\n律师 B: %s\n独立核查人 C: %s\n", lawyerAEmail, lawyerBEmail, officerEmail)
	fmt.Printf("A 案件: %d/%s\nB 隔离案件: %d/%s\n检测单: %s\n审批单: %s/%s\n", result.caseA.ID, caseANumber, result.caseB.ID, caseBNumber, checkID, approvalID, approvalNo)
	fmt.Println("下一步：用 QA_PASSWORD 从 A 登录；再分别用 B、C 登录执行浏览器验收。")
}

type seedResult struct {
	lawyerA models.User
	lawyerB models.User
	officer models.User
	caseA   models.Case
	caseB   models.Case
}

func requireNonProductionEnvironment() error {
	environment := strings.ToLower(strings.TrimSpace(os.Getenv("ENVIRONMENT")))
	if environment == "" {
		environment = "development"
	}
	if environment == "production" {
		return errors.New("ENVIRONMENT=production 被拒绝")
	}
	switch environment {
	case "development", "test", "qa":
		return nil
	case "staging":
		if os.Getenv("QA_FIXTURE_ALLOW_STAGING") == "1" {
			return nil
		}
		return errors.New("staging 需要额外设置 QA_FIXTURE_ALLOW_STAGING=1")
	default:
		return fmt.Errorf("环境 %q 不在允许列表 development/test/qa 中", environment)
	}
}

func requireFixtureTables(db *gorm.DB) error {
	for _, table := range []string{
		"roles", "permissions", "role_permissions", "user_roles", "users", "clients", "cases",
		"entities", "entity_relations", "entity_name_history", "case_parties", "case_ethical_wall_whitelist",
		"conflict_checks", "conflict_details", "conflict_check_records", "conflict_cases", "conflict_search_scopes",
		"conflict_index_build_runs", "conflict_reviewer_assignments", "approval_requests",
	} {
		if !db.Migrator().HasTable(table) {
			return fmt.Errorf("缺少表 %s，请先完成 PostgreSQL schema bootstrap/迁移", table)
		}
	}
	return nil
}

// Imported development databases may contain explicit IDs while their
// PostgreSQL sequences still point at the initial value. Keep the fixture
// repeatable without changing application migrations or production startup.
func syncPostgreSQLSequences(db *gorm.DB) error {
	for _, table := range []string{
		"roles", "permissions", "user_roles", "users", "clients", "cases", "entities", "entity_relations", "entity_name_history", "case_parties", "case_ethical_wall_whitelist",
		"conflict_checks", "conflict_details", "conflict_check_records", "conflict_cases", "conflict_search_scopes", "conflict_index_build_runs", "conflict_reviewer_assignments",
	} {
		var hasIDColumn bool
		if err := db.Raw(`SELECT EXISTS (
			SELECT 1 FROM information_schema.columns
			WHERE table_schema = current_schema() AND table_name = ? AND column_name = 'id'
		)`, table).Scan(&hasIDColumn).Error; err != nil {
			return fmt.Errorf("检查 %s 主键列失败: %w", table, err)
		}
		if !hasIDColumn {
			continue
		}
		var sequence sql.NullString
		if err := db.Raw("SELECT pg_get_serial_sequence(?, 'id')", table).Scan(&sequence).Error; err != nil {
			return fmt.Errorf("读取 %s 自增序列失败: %w", table, err)
		}
		if !sequence.Valid || sequence.String == "" {
			continue
		}
		var maxID int64
		if err := db.Raw(fmt.Sprintf(`SELECT COALESCE(MAX(id), 0) FROM "%s"`, table)).Scan(&maxID).Error; err != nil {
			return fmt.Errorf("读取 %s 最大 ID 失败: %w", table, err)
		}
		sequenceValue := maxID
		called := true
		if maxID == 0 {
			sequenceValue = 1
			called = false
		}
		if err := db.Exec("SELECT setval(?, ?, ?)", sequence.String, sequenceValue, called).Error; err != nil {
			return fmt.Errorf("同步 %s 自增序列失败: %w", table, err)
		}
	}
	return nil
}

func seedFixture(ctx context.Context, db *gorm.DB, passwordHash string) (seedResult, error) {
	var result seedResult
	err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := seedRolesAndPermissions(tx); err != nil {
			return err
		}
		var err error
		if result.lawyerA, err = upsertUser(tx, userSpec{"qa_lawyer_a", "律师甲", lawyerAEmail, "lawyer", "争议解决部", "高级"}, passwordHash); err != nil {
			return err
		}
		if result.lawyerB, err = upsertUser(tx, userSpec{"qa_lawyer_b", "律师乙", lawyerBEmail, "lawyer", "争议解决部", "高级"}, passwordHash); err != nil {
			return err
		}
		if result.officer, err = upsertUser(tx, userSpec{"qa_conflict_officer", "独立冲突核查人", officerEmail, "conflict_officer", "合规风控部", "合伙人"}, passwordHash); err != nil {
			return err
		}

		clientA, err := upsertClient(tx, clientSpec{"星河智联科技有限公司", clientAEmail, "星河智联科技有限公司", "企业", "qa-p0-fixture", "虚构新客户；用于演练承办律师 A 的接案流程。"})
		if err != nil {
			return err
		}
		clientB, err := upsertClient(tx, clientSpec{"云杉数据服务有限公司", clientBEmail, "云杉数据服务有限公司", "企业", "qa-p0-fixture", "虚构历史客户；案件启用隔离墙，仅核查人可查看历史业务证据。"})
		if err != nil {
			return err
		}

		now := time.Now().UTC()
		result.caseB, err = upsertCase(tx, models.Case{
			CaseNumber: caseBNumber, Title: "云杉数据历史服务争议（隔离演练）",
			Description: "虚构历史案件：只用于验证隔离墙和受限命中，不包含真实业务信息。",
			ClientID:    clientB.ID, LawyerID: result.lawyerB.ID, CaseType: "commercial", Priority: "medium", Status: "active",
			CreatedBy: strconv.FormatUint(uint64(result.lawyerB.ID), 10), SubjectVersion: 1, SubjectState: "EFFECTIVE",
			SubjectSnapshot:        `{"source":"qa_conflict_p0_fixture_v1","role":"CLIENT","name":"云杉数据服务有限公司"}`,
			ConflictCoverageStatus: "COMPLETE", EthicalWallEnabled: true, EthicalWallDescription: "虚构验收数据：仅独立冲突核查人可复核历史事项证据。",
			EthicalWallEnabledBy: &result.officer.ID, EthicalWallEnabledAt: &now,
		})
		if err != nil {
			return err
		}
		result.caseA, err = upsertCase(tx, models.Case{
			CaseNumber: caseANumber, Title: "星河智联新案冲突复核演练",
			Description: "虚构新案：对方名称与历史登记主体候选一致，缺少唯一身份标识，必须独立复核。",
			ClientID:    clientA.ID, LawyerID: result.lawyerA.ID, CaseType: "commercial", Priority: "high", Status: "pending",
			CreatedBy: strconv.FormatUint(uint64(result.lawyerA.ID), 10), SubjectVersion: 1, SubjectState: "EFFECTIVE",
			SubjectSnapshot: `{"source":"qa_conflict_p0_fixture_v1","role":"CLIENT","name":"星河智联科技有限公司","opposing":"云杉数据服务有限公司"}`,
			ConflictCheckID: checkID, ConflictCoverageStatus: "COMPLETE", EthicalWallEnabled: false,
		})
		if err != nil {
			return err
		}

		historical, err := upsertEntity(tx, models.Entity{
			EntityType: models.EntityTypeLegalPerson, Name: "云杉数据服务有限公司",
			Alias: "云杉数据;Yunshan Data Services", IdentityType: models.IdentityTypeBusinessLicense,
			Status: models.EntityStatusActive, BusinessScope: "虚构数据服务；仅用于验收演练。", Notes: fixtureSource,
		})
		if err != nil {
			return err
		}
		current, err := upsertEntity(tx, models.Entity{
			EntityType: models.EntityTypeLegalPerson, Name: "星河智联科技有限公司",
			Alias: "星河智联;Xinghe Intelligent", IdentityType: models.IdentityTypeBusinessLicense,
			Status: models.EntityStatusActive, BusinessScope: "虚构软件服务；仅用于验收演练。", Notes: fixtureSource,
		})
		if err != nil {
			return err
		}
		former, err := upsertEntity(tx, models.Entity{
			EntityType: models.EntityTypeLegalPerson, Name: "云杉控股有限公司",
			Alias: "云杉控股;Yunshan Holdings", IdentityType: models.IdentityTypeBusinessLicense,
			Status: models.EntityStatusActive, BusinessScope: "虚构关联主体；仅用于验收演练。", Notes: fixtureSource,
		})
		if err != nil {
			return err
		}
		if err := ensureNameHistory(tx, historical.ID, "云杉云存科技有限公司", historical.Name); err != nil {
			return err
		}
		if err := ensureEntityRelation(tx, historical.ID, former.ID, models.RelationTypeParentCompany); err != nil {
			return err
		}
		if err := ensureCaseParty(tx, result.caseB.ID, historical.ID, models.PartyRole("CLIENT"), models.PartyTypeClient, "历史客户主体"); err != nil {
			return err
		}
		if err := ensureCaseParty(tx, result.caseA.ID, current.ID, models.PartyRole("CLIENT"), models.PartyTypeClient, "当前客户主体"); err != nil {
			return err
		}
		if err := ensureCaseParty(tx, result.caseA.ID, historical.ID, models.PartyRole("OPPOSING"), models.PartyTypeOpposing, "当前对方主体；与历史登记名称候选相同，尚缺唯一身份标识"); err != nil {
			return err
		}

		if err := ensureWhitelist(tx, result.caseB.ID, result.officer.ID, result.officer.ID, "独立冲突核查人复核隔离历史证据"); err != nil {
			return err
		}
		if err := ensureWhitelist(tx, result.caseB.ID, result.lawyerB.ID, result.lawyerB.ID, "历史案件承办律师查看本人办理案件"); err != nil {
			return err
		}
		if err := upsertLegacyConflictEvidence(tx, result, clientA, clientB); err != nil {
			return err
		}
		if err := upsertReviewerAssignment(tx, result); err != nil {
			return err
		}
		if err := upsertApproval(tx, result); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return result, err
	}

	// Build the immutable firm-wide index only after all source entities and
	// case parties exist. This makes the fixture exercise the same source-to-index
	// reconciliation path as a real archive import.
	repo := repositories.NewConflictRepository(db, nil)
	reconciler, ok := repo.(repositories.ConflictP0SubjectIndexReconciler)
	if !ok {
		return result, errors.New("冲突主体索引对账接口未部署")
	}
	evidenceReference := fixtureSource + ":index"
	runs, err := reconciler.ReconcileConflictSubjectIndex(ctx, result.officer.ID, evidenceReference, true)
	if err != nil {
		return result, fmt.Errorf("建立冲突主体索引失败: %w", err)
	}
	for _, run := range runs {
		from := time.Now().UTC().AddDate(-20, 0, 0)
		to := time.Now().UTC()
		_, err := services.NewConflictScopeService(db).Upsert(ctx, services.AuthActor{UserID: result.officer.ID, Role: result.officer.Role}, services.ConflictSearchScopeInput{
			ID: "qa-p0-scope-" + strings.ToLower(run.ScopeType), ScopeType: run.ScopeType, Status: services.ConflictScopeActive,
			CoverageStatus: services.ConflictCoverageComplete, SourceVersion: run.SourceVersion, EvidenceReference: evidenceReference,
			CoveredFrom: &from, CoveredTo: &to, MissingSources: []string{}, IndexRunID: run.ID,
		})
		if err != nil {
			return result, fmt.Errorf("登记 %s 覆盖范围失败: %w", run.ScopeType, err)
		}
	}
	return result, nil
}

func seedRolesAndPermissions(db *gorm.DB) error {
	role := models.Role{Name: "独立冲突核查人", Code: "conflict_officer", Description: "仅负责利益冲突证据复核，不具备一般案件管理权限", Status: "active", SortOrder: 8}
	if err := upsertByCode(db, &role); err != nil {
		return fmt.Errorf("写入 conflict_officer 角色失败: %w", err)
	}
	for _, spec := range []models.Role{
		{Name: "律师", Code: "lawyer", Description: "律师用户", Status: "active", SortOrder: 3},
	} {
		if err := upsertByCode(db, &spec); err != nil {
			return fmt.Errorf("写入 %s 角色失败: %w", spec.Code, err)
		}
	}
	for _, spec := range []models.Permission{
		{Name: "仪表盘", Code: "dashboard", Type: "menu", Status: "active", SortOrder: 1},
		{Name: "冲突检测", Code: "conflict_check", Type: "menu", Status: "active", SortOrder: 2},
		{Name: "提交冲突复核", Code: "conflict:review", Type: "button", Status: "active", SortOrder: 3},
	} {
		if err := upsertPermission(db, &spec); err != nil {
			return fmt.Errorf("写入 %s 权限失败: %w", spec.Code, err)
		}
	}
	var officerRole models.Role
	if err := db.Where("code = ?", "conflict_officer").First(&officerRole).Error; err != nil {
		return err
	}
	var lawyerRole models.Role
	if err := db.Where("code = ?", "lawyer").First(&lawyerRole).Error; err != nil {
		return err
	}
	for _, role := range []models.Role{officerRole, lawyerRole} {
		for _, code := range []string{"dashboard", "conflict_check"} {
			var permission models.Permission
			if err := db.Where("code = ?", code).First(&permission).Error; err != nil {
				return err
			}
			if err := db.Clauses(clause.OnConflict{DoNothing: true}).Create(&models.RolePermission{RoleID: role.ID, PermissionID: permission.ID}).Error; err != nil {
				return err
			}
		}
	}
	var reviewPermission models.Permission
	if err := db.Where("code = ?", "conflict:review").First(&reviewPermission).Error; err != nil {
		return err
	}
	return db.Clauses(clause.OnConflict{DoNothing: true}).Create(&models.RolePermission{RoleID: officerRole.ID, PermissionID: reviewPermission.ID}).Error
}

func upsertByCode(db *gorm.DB, role *models.Role) error {
	var existing models.Role
	err := db.Where("code = ?", role.Code).First(&existing).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return db.Create(role).Error
	}
	if err != nil {
		return err
	}
	role.ID, role.CreatedAt, role.UpdatedAt, role.DeletedAt = existing.ID, existing.CreatedAt, time.Now(), existing.DeletedAt
	return db.Model(&existing).Updates(map[string]interface{}{"name": role.Name, "description": role.Description, "status": role.Status, "sort_order": role.SortOrder, "deleted_at": nil, "updated_at": role.UpdatedAt}).Error
}

func upsertPermission(db *gorm.DB, permission *models.Permission) error {
	var existing models.Permission
	err := db.Where("code = ?", permission.Code).First(&existing).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return db.Create(permission).Error
	}
	if err != nil {
		return err
	}
	return db.Model(&existing).Updates(map[string]interface{}{"name": permission.Name, "type": permission.Type, "status": permission.Status, "sort_order": permission.SortOrder, "deleted_at": nil, "updated_at": time.Now()}).Error
}

func upsertUser(db *gorm.DB, spec userSpec, passwordHash string) (models.User, error) {
	var user models.User
	err := db.Unscoped().Where("email = ?", spec.Email).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		user = models.User{Username: spec.Username, Name: spec.Name, Email: spec.Email, Password: passwordHash, Role: spec.Role, Status: "active", Department: spec.Department, Seniority: spec.Seniority}
		if err := db.Create(&user).Error; err != nil {
			return user, err
		}
	} else if err != nil {
		return user, err
	} else {
		if err := db.Model(&user).Updates(map[string]interface{}{"username": spec.Username, "name": spec.Name, "password": passwordHash, "role": spec.Role, "status": "active", "department": spec.Department, "seniority": spec.Seniority, "deleted_at": nil, "updated_at": time.Now()}).Error; err != nil {
			return user, err
		}
		user.Username, user.Name, user.Password, user.Role, user.Status, user.Department, user.Seniority, user.DeletedAt = spec.Username, spec.Name, passwordHash, spec.Role, "active", spec.Department, spec.Seniority, gorm.DeletedAt{}
	}
	var role models.Role
	if err := db.Where("code = ?", spec.Role).First(&role).Error; err != nil {
		return user, err
	}
	if err := db.Where("user_id = ?", user.ID).Delete(&models.UserRole{}).Error; err != nil {
		return user, err
	}
	if err := db.Clauses(clause.OnConflict{DoNothing: true}).Create(&models.UserRole{UserID: user.ID, RoleID: role.ID}).Error; err != nil {
		return user, err
	}
	return user, nil
}

func upsertClient(db *gorm.DB, spec clientSpec) (models.Client, error) {
	var client models.Client
	err := db.Unscoped().Where("email = ?", spec.Email).First(&client).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		client = models.Client{Name: spec.Name, Email: spec.Email, Company: spec.Company, Type: spec.Type, Source: spec.Source, Notes: spec.Notes, Status: "active", Version: 1}
		return client, db.Create(&client).Error
	}
	if err != nil {
		return client, err
	}
	updates := map[string]interface{}{"name": spec.Name, "company": spec.Company, "type": spec.Type, "source": spec.Source, "notes": spec.Notes, "status": "active", "deleted_at": nil, "updated_at": time.Now()}
	if err := db.Model(&client).Updates(updates).Error; err != nil {
		return client, err
	}
	client.Name, client.Company, client.Type, client.Source, client.Notes, client.Status = spec.Name, spec.Company, spec.Type, spec.Source, spec.Notes, "active"
	return client, nil
}

func upsertCase(db *gorm.DB, matter models.Case) (models.Case, error) {
	var existing models.Case
	err := db.Unscoped().Where("case_number = ?", matter.CaseNumber).First(&existing).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		if err := db.Create(&matter).Error; err != nil {
			return matter, err
		}
		return matter, nil
	}
	if err != nil {
		return existing, err
	}
	if err := db.Model(&existing).Updates(map[string]interface{}{
		"title": matter.Title, "description": matter.Description, "client_id": matter.ClientID, "lawyer_id": matter.LawyerID,
		"case_type": matter.CaseType, "priority": matter.Priority, "status": matter.Status, "created_by": matter.CreatedBy,
		"subject_version": matter.SubjectVersion, "subject_state": matter.SubjectState, "subject_snapshot": matter.SubjectSnapshot,
		"pending_subject_revision_id": matter.PendingSubjectRevisionID, "conflict_check_id": matter.ConflictCheckID,
		"conflict_coverage_status": matter.ConflictCoverageStatus, "ethical_wall_enabled": matter.EthicalWallEnabled,
		"ethical_wall_description": matter.EthicalWallDescription, "ethical_wall_enabled_by": matter.EthicalWallEnabledBy,
		"ethical_wall_enabled_at": matter.EthicalWallEnabledAt, "deleted_at": nil, "updated_at": time.Now(),
	}).Error; err != nil {
		return existing, err
	}
	matter.ID, matter.CreatedAt = existing.ID, existing.CreatedAt
	return matter, nil
}

func upsertEntity(db *gorm.DB, entity models.Entity) (models.Entity, error) {
	var existing models.Entity
	err := db.Unscoped().Where("name = ? AND notes = ?", entity.Name, fixtureSource).First(&existing).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return entity, db.Create(&entity).Error
	}
	if err != nil {
		return existing, err
	}
	if err := db.Model(&existing).Updates(map[string]interface{}{"entity_type": entity.EntityType, "alias": entity.Alias, "identity_type": entity.IdentityType, "status": entity.Status, "business_scope": entity.BusinessScope, "notes": fixtureSource, "deleted_at": nil, "updated_at": time.Now()}).Error; err != nil {
		return existing, err
	}
	existing.Name, existing.Alias, existing.EntityType, existing.IdentityType, existing.Status, existing.BusinessScope, existing.Notes = entity.Name, entity.Alias, entity.EntityType, entity.IdentityType, entity.Status, entity.BusinessScope, fixtureSource
	return existing, nil
}

func ensureNameHistory(db *gorm.DB, entityID uint, oldName, newName string) error {
	var count int64
	if err := db.Model(&models.EntityNameHistory{}).Where("entity_id = ? AND old_name = ? AND new_name = ?", entityID, oldName, newName).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	return db.Create(&models.EntityNameHistory{EntityID: entityID, OldName: oldName, NewName: newName, ChangeDate: time.Now().UTC(), ChangeReason: "虚构验收数据：覆盖曾用名检索"}).Error
}

func ensureEntityRelation(db *gorm.DB, sourceID, targetID uint, relation models.RelationType) error {
	var row models.EntityRelation
	err := db.Where("source_entity_id = ? AND target_entity_id = ? AND relation_type = ?", sourceID, targetID, relation).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return db.Create(&models.EntityRelation{SourceEntityID: sourceID, TargetEntityID: targetID, RelationType: relation, IsActive: true, DataSource: fixtureSource, Description: "虚构关联主体：仅用于验证关系穿透提醒。"}).Error
	}
	if err != nil {
		return err
	}
	return db.Model(&row).Updates(map[string]interface{}{"is_active": true, "data_source": fixtureSource, "description": "虚构关联主体：仅用于验证关系穿透提醒。", "deleted_at": nil, "updated_at": time.Now()}).Error
}

func ensureCaseParty(db *gorm.DB, caseID, entityID uint, role models.PartyRole, partyType models.PartyType, description string) error {
	var row models.CaseParty
	err := db.Where("case_id = ? AND entity_id = ? AND role = ?", caseID, entityID, role).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return db.Create(&models.CaseParty{CaseID: caseID, EntityID: entityID, Role: role, PartyType: partyType, Description: description}).Error
	}
	if err != nil {
		return err
	}
	return db.Model(&row).Updates(map[string]interface{}{"party_type": partyType, "description": description, "deleted_at": nil, "updated_at": time.Now()}).Error
}

func ensureWhitelist(db *gorm.DB, caseID, userID, grantedBy uint, reason string) error {
	var row models.CaseEthicalWallWhitelist
	err := db.Where("case_id = ? AND user_id = ?", caseID, userID).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return db.Create(&models.CaseEthicalWallWhitelist{CaseID: caseID, UserID: userID, GrantedBy: grantedBy, GrantedAt: time.Now().UTC(), Reason: reason}).Error
	}
	if err != nil {
		return err
	}
	return db.Model(&row).Updates(map[string]interface{}{"granted_by": grantedBy, "granted_at": time.Now().UTC(), "reason": reason}).Error
}

func upsertLegacyConflictEvidence(db *gorm.DB, result seedResult, clientA, clientB models.Client) error {
	now := time.Now().UTC()
	normalizedSubjects := []map[string]interface{}{
		{"role": "CLIENT", "originalName": clientA.Name, "normalizedName": clientA.Name, "entityType": "LEGAL_PERSON"},
		{"role": "OPPOSING", "originalName": "云杉数据服务有限公司", "normalizedName": "云杉数据服务有限公司", "entityType": "LEGAL_PERSON", "aliases": []string{"云杉数据", "云杉云存科技有限公司"}},
	}
	evidence := map[string]interface{}{
		"evidenceId": "QA-P0-A-EVIDENCE-20260719", "ruleCode": "SUBJECT_CANDIDATE_REVIEW", "matchType": "CANDIDATE",
		"sourceType": "CASE_ARCHIVE", "requestedParty": "云杉数据服务有限公司", "matchedEntity": "云杉数据服务有限公司", "partyRole": "OPPOSING",
		"historicalRole": "CLIENT", "sourceCaseId": strconv.FormatUint(uint64(result.caseB.ID), 10), "sourceCaseNumber": caseBNumber,
		"restricted": true, "summary": "存在受隔离记录，请联系独立冲突核查人。",
	}
	conflictCaseSummary := map[string]interface{}{
		"caseId": strconv.FormatUint(uint64(result.caseB.ID), 10), "caseNo": caseBNumber, "caseName": result.caseB.Title,
		"conflictType": "名称候选待核实", "riskLevel": "MEDIUM", "description": "当前对方名称与历史登记主体名称一致，但没有唯一主体标识，不能直接认定为同一主体。历史业务详情受隔离墙保护。",
		"caseStatus": "active", "matchType": "CANDIDATE", "ruleCode": "SUBJECT_CANDIDATE_REVIEW", "restricted": true,
		"evidence": []interface{}{evidence},
	}
	searchParameters := models.JSON{
		"query": "云杉数据服务有限公司", "subjectCaseId": strconv.FormatUint(uint64(result.caseA.ID), 10), "subjectCaseNumber": caseANumber,
		"searchYears": 0, "includeCorporateRelations": true, "searchDepth": "DEEP", "coverageStatus": "COMPLETE", "source": fixtureSource,
		"subjects": normalizedSubjects, "opposingParties": []string{"云杉数据服务有限公司"},
	}
	checkResult := models.JSON{
		"checkId": checkID, "isConflict": true,
		"decision":           map[string]interface{}{"status": "REVIEW_REQUIRED", "recommendation": "名称候选与历史登记主体一致，但缺少唯一主体标识；独立核查完成前不得作为无冲突放行。", "ruleCodes": []string{"SUBJECT_CANDIDATE_REVIEW"}, "requiresManualReview": true, "evidenceCount": 1, "restrictedCount": 1, "coverageStatus": "COMPLETE", "coverageNotice": "已覆盖本夹具登记的四类虚构来源；不代表真实律所档案覆盖。"},
		"riskAssessment":     map[string]interface{}{"overallRisk": "MEDIUM", "riskScore": 0, "riskReason": "存在待核实的名称候选和受隔离历史记录", "requiresApproval": true, "approvalLevel": "CONFLICT_OFFICER"},
		"normalizedSubjects": normalizedSubjects,
		"evidence":           []interface{}{evidence},
		"conflictCases":      []interface{}{conflictCaseSummary},
		"fixture":            fixtureSource,
	}
	row := models.ConflictCheckRecord{CheckID: checkID, ClientID: strconv.FormatUint(uint64(clientA.ID), 10), ClientName: clientA.Name, CaseName: result.caseA.Title, CaseType: result.caseA.CaseType, CheckStatus: "COMPLETED", HasConflict: true, RiskLevel: "MEDIUM", SearchParameters: searchParameters, CheckResult: checkResult, UserID: result.lawyerA.ID, Duration: 214, CheckTime: now, CreatedAt: now, UpdatedAt: now}
	var existing models.ConflictCheckRecord
	if err := db.Where("check_id = ?", checkID).First(&existing).Error; errors.Is(err, gorm.ErrRecordNotFound) {
		if err := db.Create(&row).Error; err != nil {
			return fmt.Errorf("写入冲突检测记录失败: %w", err)
		}
	} else if err != nil {
		return err
	} else if err := db.Model(&existing).Updates(map[string]interface{}{"client_id": row.ClientID, "client_name": row.ClientName, "case_name": row.CaseName, "case_type": row.CaseType, "check_status": row.CheckStatus, "has_conflict": row.HasConflict, "risk_level": row.RiskLevel, "search_parameters": row.SearchParameters, "check_result": row.CheckResult, "user_id": row.UserID, "duration": row.Duration, "check_time": row.CheckTime, "updated_at": row.UpdatedAt}).Error; err != nil {
		return fmt.Errorf("更新冲突检测记录失败: %w", err)
	}

	conflict := models.ConflictCase{ID: checkID + "-restricted-history", CheckID: checkID, CaseID: strconv.FormatUint(uint64(result.caseB.ID), 10), CaseName: result.caseB.Title, CaseNo: caseBNumber, CaseType: result.caseB.CaseType, ConflictType: "名称候选待核实", RiskLevel: "MEDIUM", Description: "当前对方名称与历史登记主体名称一致，但没有唯一主体标识，不能直接认定为同一主体。历史业务详情受隔离墙保护。", CaseStatus: "active", ClientID: strconv.FormatUint(uint64(clientB.ID), 10), OpposingParties: models.JSONStringArray{"云杉数据服务有限公司"}, ConflictDetails: "{\"restricted\":true,\"source\":\"" + fixtureSource + "\"}", CreatedAt: now}
	var existingConflict models.ConflictCase
	if err := db.Where("id = ?", conflict.ID).First(&existingConflict).Error; errors.Is(err, gorm.ErrRecordNotFound) {
		if err := db.Create(&conflict).Error; err != nil {
			return fmt.Errorf("写入冲突命中记录失败: %w", err)
		}
	} else if err != nil {
		return err
	} else if err := db.Model(&existingConflict).Updates(map[string]interface{}{"check_id": conflict.CheckID, "case_id": conflict.CaseID, "case_name": conflict.CaseName, "case_no": conflict.CaseNo, "case_type": conflict.CaseType, "conflict_type": conflict.ConflictType, "risk_level": conflict.RiskLevel, "description": conflict.Description, "case_status": conflict.CaseStatus, "client_id": conflict.ClientID, "opposing_parties": conflict.OpposingParties, "conflict_details": conflict.ConflictDetails}).Error; err != nil {
		return fmt.Errorf("更新冲突命中记录失败: %w", err)
	}

	check := models.ConflictCheck{CaseID: result.caseA.ID, Status: "REVIEW_REQUIRED", RequestedBy: result.lawyerA.ID, RequestedAt: now, CheckedBy: &result.officer.ID, CheckedAt: &now, Result: &models.CheckResult{HasConflict: true, TotalConflicts: 1, CompletedAt: now, CoverageStatus: "COMPLETE", CoverageNotice: "存在待独立核查的名称候选。"}, ResultSummary: "名称候选待独立核查；不能作为无冲突确认。", TotalConflicts: 1, MediumCount: 1, CheckParams: searchParameters}
	var existingCheck models.ConflictCheck
	err := db.Where("case_id = ? AND requested_by = ?", result.caseA.ID, result.lawyerA.ID).Order("id DESC").First(&existingCheck).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		if err := db.Create(&check).Error; err != nil {
			return fmt.Errorf("写入结构化冲突检测失败: %w", err)
		}
	} else if err != nil {
		return err
	} else if err := db.Model(&existingCheck).Updates(map[string]interface{}{"status": check.Status, "checked_by": check.CheckedBy, "checked_at": check.CheckedAt, "result": check.Result, "result_summary": check.ResultSummary, "total_conflicts": check.TotalConflicts, "medium_count": check.MediumCount, "check_params": check.CheckParams, "updated_at": now}).Error; err != nil {
		return fmt.Errorf("更新结构化冲突检测失败: %w", err)
	}
	return nil
}

func upsertReviewerAssignment(db *gorm.DB, result seedResult) error {
	row := models.ConflictReviewerAssignment{ID: assignmentID, CheckID: checkID, CaseID: &result.caseA.ID, ReviewerID: result.officer.ID, AssignedBy: result.officer.ID, Status: models.ConflictReviewerAssignmentActive, RecusalDeclared: true, IndependenceReason: "虚构验收数据：核查人不属于申请律师或承办律师的直接管理链。", EffectiveFrom: ptr(time.Now().UTC()), CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()}
	var existing models.ConflictReviewerAssignment
	err := db.Where("check_id = ? AND reviewer_id = ? AND status = ?", checkID, result.officer.ID, models.ConflictReviewerAssignmentActive).First(&existing).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return db.Create(&row).Error
	}
	if err != nil {
		return err
	}
	return db.Model(&existing).Updates(map[string]interface{}{"case_id": row.CaseID, "assigned_by": row.AssignedBy, "recusal_declared": true, "independence_reason": row.IndependenceReason, "effective_from": row.EffectiveFrom, "updated_at": row.UpdatedAt}).Error
}

func upsertApproval(db *gorm.DB, result seedResult) error {
	now := time.Now().UTC()
	approvalEvidence := map[string]interface{}{
		"evidenceId":       "QA-P0-A-EVIDENCE-20260719",
		"ruleCode":         "SUBJECT_CANDIDATE_REVIEW",
		"matchType":        "CANDIDATE",
		"sourceType":       "CASE_ARCHIVE",
		"requestedParty":   "云杉数据服务有限公司",
		"matchedEntity":    "云杉数据服务有限公司",
		"partyRole":        "OPPOSING",
		"historicalRole":   "CLIENT",
		"sourceCaseId":     strconv.FormatUint(uint64(result.caseB.ID), 10),
		"sourceCaseNumber": caseBNumber,
		"restricted":       true,
		"summary":          "存在受隔离记录，请联系独立冲突核查人。",
	}
	conflictCases := []map[string]interface{}{{
		"case_id":       result.caseB.ID,
		"case_no":       caseBNumber,
		"case_name":     result.caseB.Title,
		"conflict_type": "名称候选待核实",
		"risk_level":    "MEDIUM",
		"description":   "当前对方名称与历史登记主体名称一致，但没有唯一主体标识，不能直接认定为同一主体。历史业务详情受隔离墙保护。",
		"restricted":    true,
	}}
	decision := map[string]interface{}{
		"status":               "REVIEW_REQUIRED",
		"requiresManualReview": true,
		"coverageStatus":       "COMPLETE",
		"restrictedCount":      1,
		"recommendation":       "名称候选与历史登记主体一致，但缺少唯一主体标识；独立核查完成前不得作为无冲突放行。",
	}
	conflictResultPayload := map[string]interface{}{
		"checkId":  checkID,
		"decision": decision,
		"riskAssessment": map[string]interface{}{
			"overallRisk": "REVIEW_REQUIRED", "riskScore": 0, "requiresApproval": true,
			"requestedParty": "云杉数据服务有限公司", "matchEvidence": approvalEvidence,
		},
		"normalizedSubjects": []map[string]interface{}{
			{"role": "CLIENT", "originalName": "星河智联科技有限公司", "normalizedName": "星河智联科技有限公司"},
			{"role": "OPPOSING", "originalName": "云杉数据服务有限公司", "normalizedName": "云杉数据服务有限公司"},
		},
		"evidence":      []interface{}{approvalEvidence},
		"conflictCases": conflictCases,
		"record": map[string]interface{}{
			"check_id": checkID, "case_name": result.caseA.Title, "client_name": "星河智联科技有限公司",
			"risk_level": "REVIEW_REQUIRED", "has_conflict": true, "status": "COMPLETED",
		},
	}
	metadataPayload := map[string]interface{}{
		"source":              fixtureSource,
		"subject_case_id":     strconv.FormatUint(uint64(result.caseA.ID), 10),
		"subject_case_number": caseANumber,
		"conflict_task_id":    checkID,
		"client_name":         "星河智联科技有限公司",
		"opposing_parties":    []string{"云杉数据服务有限公司"},
		"subjects": []map[string]interface{}{
			{"role": "CLIENT", "name": "星河智联科技有限公司"},
			{"role": "OPPOSING", "name": "云杉数据服务有限公司"},
		},
		"normalizedSubjects": []map[string]interface{}{
			{"role": "CLIENT", "originalName": "星河智联科技有限公司", "normalizedName": "星河智联科技有限公司"},
			{"role": "OPPOSING", "originalName": "云杉数据服务有限公司", "normalizedName": "云杉数据服务有限公司"},
		},
		"decision":        decision,
		"evidence":        []map[string]interface{}{approvalEvidence},
		"conflict_result": conflictResultPayload,
		"conflict_cases":  conflictCases,
	}
	metadataBytes, err := json.Marshal(metadataPayload)
	if err != nil {
		return fmt.Errorf("生成审批快照元数据失败: %w", err)
	}
	conflictResultBytes, err := json.Marshal(conflictResultPayload)
	if err != nil {
		return fmt.Errorf("生成冲突审批结果失败: %w", err)
	}
	conflictResult := string(conflictResultBytes)
	metadata := string(metadataBytes)
	row := models.ApprovalRequest{ID: approvalID, RequestNumber: approvalNo, Title: "冲突审查审批 - 星河智联新案", Type: "conflict_approval", Category: "conflict_review", Content: "名称候选需要独立冲突核查人复核；信息不足时不得成案。", ApplicantID: strconv.FormatUint(uint64(result.lawyerA.ID), 10), ApplicantName: result.lawyerA.Name, ApplicantTitle: "律师", DepartmentID: "qa", DepartmentName: "争议解决部", Urgency: "urgent", Priority: "high", Status: "submitted", SubmissionDate: &now, CurrentStage: "独立冲突复核", CurrentApproverID: strconv.FormatUint(uint64(result.officer.ID), 10), CurrentApproverName: result.officer.Name, WorkflowType: "CONFLICT_APPROVAL", WorkflowConfig: "{}", Attachments: "[]", Metadata: metadata, ConflictCheckID: checkID, ConflictRiskLevel: "MEDIUM", ConflictCheckTime: &now, ConflictResult: conflictResult, CaseCreated: false, CaseCreationStatus: "BLOCKED_CONFLICT_REVIEW", CreatedBy: strconv.FormatUint(uint64(result.lawyerA.ID), 10), UpdatedBy: strconv.FormatUint(uint64(result.lawyerA.ID), 10), CreatedAt: now, UpdatedAt: now}
	var existing models.ApprovalRequest
	err = db.Where("id = ?", approvalID).First(&existing).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		if err := db.Create(&row).Error; err != nil {
			return err
		}
		return ensureApprovalSnapshot(db, result, row.Metadata, row.ConflictResult, now)
	}
	if err != nil {
		return err
	}
	if err := db.Model(&existing).Updates(map[string]interface{}{"request_number": row.RequestNumber, "title": row.Title, "content": row.Content, "applicant_id": row.ApplicantID, "applicant_name": row.ApplicantName, "status": row.Status, "submission_date": row.SubmissionDate, "current_stage": row.CurrentStage, "current_approver_id": row.CurrentApproverID, "current_approver_name": row.CurrentApproverName, "metadata": row.Metadata, "conflict_check_id": row.ConflictCheckID, "conflict_risk_level": row.ConflictRiskLevel, "conflict_check_time": row.ConflictCheckTime, "conflict_result": row.ConflictResult, "case_created": false, "created_case_id": "", "case_creation_status": row.CaseCreationStatus, "updated_by": row.UpdatedBy, "updated_at": row.UpdatedAt}).Error; err != nil {
		return err
	}
	return ensureApprovalSnapshot(db, result, row.Metadata, row.ConflictResult, now)
}

func ensureApprovalSnapshot(db *gorm.DB, result seedResult, metadata, conflictResult string, createdAt time.Time) error {
	if !db.Migrator().HasTable("approval_snapshots") {
		return nil
	}
	var metadataValue map[string]interface{}
	if err := json.Unmarshal([]byte(metadata), &metadataValue); err != nil {
		return fmt.Errorf("解析审批快照元数据失败: %w", err)
	}
	var conflictResultValue interface{}
	if err := json.Unmarshal([]byte(conflictResult), &conflictResultValue); err != nil {
		return fmt.Errorf("解析冲突审批结果失败: %w", err)
	}
	snapshot := map[string]interface{}{
		"snapshot_type":      "conflict_approval",
		"source":             fixtureSource,
		"conflict_task_id":   checkID,
		"client_name":        "星河智联科技有限公司",
		"opposing_parties":   []string{"云杉数据服务有限公司"},
		"metadata":           metadataValue,
		"conflict_result":    conflictResultValue,
		"conflict_cases":     metadataValue["conflict_cases"],
		"approval":           map[string]interface{}{"id": approvalID, "request_number": approvalNo, "title": "冲突审查审批 - 星河智联新案", "status": "submitted", "current_stage": "独立冲突复核", "current_approver_name": result.officer.Name, "applicant_name": result.lawyerA.Name},
		"subjects":           metadataValue["subjects"],
		"normalizedSubjects": metadataValue["normalizedSubjects"],
		"decision":           metadataValue["decision"],
		"evidence":           metadataValue["evidence"],
	}
	return db.Table("approval_snapshots").Create(map[string]interface{}{
		"approval_request_id": approvalID,
		"snapshot_type":       "conflict_approval",
		"snapshot_data":       snapshot,
		"source_version":      1,
		"created_at":          createdAt,
	}).Error
}

func ptr[T any](value T) *T { return &value }

func verifyFixture(ctx context.Context, db *gorm.DB) error {
	for _, email := range []string{lawyerAEmail, lawyerBEmail, officerEmail} {
		var user models.User
		if err := db.WithContext(ctx).Where("email = ? AND status = ? AND deleted_at IS NULL", email, "active").First(&user).Error; err != nil {
			return fmt.Errorf("账号 %s 不存在或未启用: %w", email, err)
		}
	}
	var caseA, caseB models.Case
	if err := db.WithContext(ctx).Where("case_number = ?", caseANumber).First(&caseA).Error; err != nil {
		return fmt.Errorf("A 案不存在: %w", err)
	}
	if err := db.WithContext(ctx).Where("case_number = ? AND ethical_wall_enabled = ?", caseBNumber, true).First(&caseB).Error; err != nil {
		return fmt.Errorf("B 隔离案件不存在或隔离墙未开启: %w", err)
	}
	var count int64
	if err := db.WithContext(ctx).Model(&models.ConflictCheckRecord{}).Where("check_id = ? AND user_id = ? AND check_status = ?", checkID, caseA.LawyerID, "COMPLETED").Count(&count).Error; err != nil || count != 1 {
		return fmt.Errorf("A 的冲突检测记录不符合预期: count=%d err=%v", count, err)
	}
	if err := db.WithContext(ctx).Model(&models.ConflictReviewerAssignment{}).Where("check_id = ? AND status = ? AND recusal_declared = ?", checkID, models.ConflictReviewerAssignmentActive, true).Count(&count).Error; err != nil || count != 1 {
		return fmt.Errorf("独立复核指定不符合预期: count=%d err=%v", count, err)
	}
	if err := db.WithContext(ctx).Model(&models.ApprovalRequest{}).Where("id = ? AND status = ? AND case_created = ? AND created_case_id = ''", approvalID, "submitted", false).Count(&count).Error; err != nil || count != 1 {
		return fmt.Errorf("审批门禁夹具不符合预期: count=%d err=%v", count, err)
	}
	var activeScopes int64
	if err := db.WithContext(ctx).Model(&models.ConflictSearchScope{}).Where("status = ? AND coverage_status = ?", services.ConflictScopeActive, services.ConflictCoverageComplete).Count(&activeScopes).Error; err != nil || activeScopes < 4 {
		return fmt.Errorf("四类冲突范围未全部登记: active_complete=%d err=%v", activeScopes, err)
	}
	return nil
}

func fatal(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "qa-fixture: "+format+"\n", args...)
	os.Exit(1)
}
