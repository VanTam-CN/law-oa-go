package services

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"law-oa-go/internal/repositories"
	"law-oa-go/internal/security"
)

// safeGo 安全启动 goroutine，捕获 panic 防止进程崩溃
// 接受 context 参数用于传播链路追踪信息
func safeGo(ctx context.Context, name string, f func(context.Context)) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				traceID := ctx.Value("trace_id")
				log.Printf("[ConflictHook] panic recovered in %s (traceID: %v): %v", name, traceID, r)
			}
		}()
		f(ctx)
	}()
}

// ConflictHookService 冲突检测自动触发服务
// 在案件创建/更新/实体变更时自动触发冲突检测
type ConflictHookService interface {
	OnCaseCreated(ctx context.Context, caseID uint, clientID uint)
	OnCaseUpdated(ctx context.Context, caseID uint)
	OnEntityAddedToCase(ctx context.Context, caseID uint, entityID uint)
	OnEntityUpdated(ctx context.Context, entityID uint)
}

type conflictHookService struct {
	conflictService ConflictCheckService
	caseRepo        repositories.CaseRepository
	entityRepo      repositories.EntityRepository
	clientRepo      repositories.ClientRepository
}

// NewConflictHookService 创建冲突检测 Hook 服务
func NewConflictHookService(
	conflictService ConflictCheckService,
	caseRepo repositories.CaseRepository,
	entityRepo repositories.EntityRepository,
	clientRepo repositories.ClientRepository,
) ConflictHookService {
	return &conflictHookService{
		conflictService: conflictService,
		caseRepo:        caseRepo,
		entityRepo:      entityRepo,
		clientRepo:      clientRepo,
	}
}

// OnCaseCreated 案件创建后自动触发冲突检测
func (s *conflictHookService) OnCaseCreated(ctx context.Context, caseID uint, clientID uint) {
	safeGo(ctx, "runConflictCheck", func(ctx context.Context) {
		bgCtx := context.WithValue(context.Background(), "trace_id", ctx.Value("trace_id"))
		s.runConflictCheck(bgCtx, caseID, clientID, "case_created")
	})
}

// OnCaseUpdated 案件更新后自动触发冲突检测
func (s *conflictHookService) OnCaseUpdated(ctx context.Context, caseID uint) {
	caseData, err := s.caseRepo.FindByID(ctx, caseID)
	if err != nil || caseData == nil {
		return
	}
	if caseData.ClientID > 0 {
		safeGo(ctx, "runConflictCheck", func(ctx context.Context) {
			bgCtx := context.WithValue(context.Background(), "trace_id", ctx.Value("trace_id"))
			s.runConflictCheck(bgCtx, caseID, caseData.ClientID, "case_updated")
		})
	}
}

// OnEntityAddedToCase 新当事人关联到案件时触发检测
func (s *conflictHookService) OnEntityAddedToCase(ctx context.Context, caseID uint, entityID uint) {
	safeGo(ctx, "checkEntityConflict", func(ctx context.Context) {
		bgCtx := context.WithValue(context.Background(), "trace_id", ctx.Value("trace_id"))
		s.checkEntityConflict(bgCtx, caseID, entityID, "party_added")
	})
}

// OnEntityUpdated 实体信息变更时触发关联案件检测
func (s *conflictHookService) OnEntityUpdated(ctx context.Context, entityID uint) {
	// 简化实现：仅记录日志，实际生产中应查询实体关联的活跃案件并逐一检测
	log.Printf("[ConflictHook] 实体信息已变更 (EntityID: %d)，建议手动触发关联案件冲突检测", entityID)
}

// runConflictCheck 基于客户信息执行冲突检测
func (s *conflictHookService) runConflictCheck(ctx context.Context, caseID uint, clientID uint, trigger string) {
	bgCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	traceID := ctx.Value("trace_id")

	caseData, err := s.caseRepo.FindByID(bgCtx, caseID)
	if err != nil || caseData == nil {
		log.Printf("[ConflictHook] 获取案件失败 (traceID: %v, CaseID: %d): %v", traceID, caseID, err)
		return
	}

	client, err := s.clientRepo.FindByID(bgCtx, clientID)
	if err != nil || client == nil {
		log.Printf("[ConflictHook] 获取客户失败 (traceID: %v, ClientID: %d): %v", traceID, clientID, err)
		return
	}

	identityNumber, _ := client.DecryptedIdentity()
	req := &ConflictCheckRequest{
		CaseID:    caseID,
		CaseTitle: caseData.Title,
		CheckEntities: []EntityCheckInfo{
			{
				EntityID:       clientID,
				EntityName:     client.Name,
				IdentityType:   string(client.EffectiveIdentityType()),
				IdentityNumber: identityNumber,
				PartyType:      "CLIENT",
			},
		},
		SearchDepth: 3,
		RequestedBy: 0,
	}

	resp, err := s.conflictService.CheckConflict(bgCtx, req)
	if err != nil {
		log.Printf("[ConflictHook] 冲突检测失败 (traceID: %v, CaseID: %d, Trigger: %s): %v", traceID, caseID, trigger, err)
		return
	}

	if resp.HasConflict {
		log.Printf("[ConflictHook] 检测到冲突 (traceID: %v, CaseID: %d, Trigger: %s, 冲突数: %d)", traceID, caseID, trigger, resp.TotalConflicts)
	}
}

// checkEntityConflict 基于指定实体执行冲突检测
func (s *conflictHookService) checkEntityConflict(ctx context.Context, caseID uint, entityID uint, trigger string) {
	bgCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	traceID := ctx.Value("trace_id")

	caseData, err := s.caseRepo.FindByID(bgCtx, caseID)
	if err != nil || caseData == nil {
		log.Printf("[ConflictHook] 获取案件失败 (traceID: %v, CaseID: %d): %v", traceID, caseID, err)
		return
	}

	entity, err := s.entityRepo.GetByID(bgCtx, entityID)
	if err != nil || entity == nil {
		log.Printf("[ConflictHook] 获取实体失败 (traceID: %v, EntityID: %d): %v", traceID, entityID, err)
		return
	}
	identityNumber := strings.TrimSpace(entity.IdentityNumber)
	if identityNumber == "" && strings.TrimSpace(entity.IdentityNumberCiphertext) != "" {
		identityNumber, _ = security.DecryptIdentityNumber(entity.IdentityNumberCiphertext)
	}

	req := &ConflictCheckRequest{
		CaseID:    caseID,
		CaseTitle: caseData.Title,
		CheckEntities: []EntityCheckInfo{
			{
				EntityID:       entityID,
				EntityName:     entity.Name,
				IdentityType:   string(entity.IdentityType),
				IdentityNumber: identityNumber,
				PartyType:      "OPPOSING",
			},
		},
		SearchDepth: 3,
		RequestedBy: 0,
	}

	resp, err := s.conflictService.CheckConflict(bgCtx, req)
	if err != nil {
		log.Printf("[ConflictHook] 实体冲突检测失败 (traceID: %v, CaseID: %d, EntityID: %d): %v", traceID, caseID, entityID, err)
		return
	}

	if resp.HasConflict {
		log.Printf("[ConflictHook] 实体冲突 (traceID: %v, CaseID: %d, EntityID: %d, 冲突数: %d)", traceID, caseID, entityID, resp.TotalConflicts)
	}
}

// recordConflictAlert 记录冲突告警（内部方法）
func recordConflictAlert(caseID uint, totalConflicts int, trigger string) {
	log.Printf("[ConflictHook] 冲突告警已记录 (CaseID: %d, Conflicts: %d, Trigger: %s)",
		caseID, totalConflicts, trigger)
	_ = fmt.Sprintf("冲突检测告警: %s (案件ID: %d, 冲突数: %d)", trigger, caseID, totalConflicts)
}
