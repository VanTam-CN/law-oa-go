package services

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"law-oa-go/internal/models"
	"law-oa-go/internal/repositories"
)

// mockRecursiveRepo 测试递归穿透的 mock，只覆盖必要方法
type mockRecursiveRepo struct {
	repositories.EntityRepository
	nodes        []*repositories.RelatedEntityNode
	relatedCases map[uint][]*models.Case
}

func (m *mockRecursiveRepo) GetRelatedEntitiesRecursive(_ context.Context, _ uint, _ int) ([]*repositories.RelatedEntityNode, error) {
	return m.nodes, nil
}

func (m *mockRecursiveRepo) GetActiveCasesByEntity(_ context.Context, entityID uint) ([]*models.Case, error) {
	if cases, ok := m.relatedCases[entityID]; ok {
		return cases, nil
	}
	return []*models.Case{}, nil
}

func TestBuildPathDescription(t *testing.T) {
	tests := []struct {
		name     string
		node     *repositories.RelatedEntityNode
		expected string
	}{
		{
			name: "空路径",
			node: &repositories.RelatedEntityNode{
				RelationType: models.RelationTypeMajorShareholder,
				Path:         []repositories.RelatedEdge{},
			},
			expected: "MAJOR_SHAREHOLDER",
		},
		{
			name: "单层正向路径",
			node: &repositories.RelatedEntityNode{
				RelationType: models.RelationTypeParentCompany,
				Direction:    "outgoing",
				Path: []repositories.RelatedEdge{
					{FromID: 1, ToID: 2, RelationType: models.RelationTypeParentCompany},
				},
			},
			expected: "[PARENT_COMPANY(→)2]",
		},
		{
			name: "单层反向路径",
			node: &repositories.RelatedEntityNode{
				RelationType: models.RelationTypeSubsidiary,
				Direction:    "incoming",
				Path: []repositories.RelatedEdge{
					{FromID: 3, ToID: 1, RelationType: models.RelationTypeSubsidiary},
				},
			},
			expected: "[SUBSIDIARY(←)1]",
		},
		{
			name: "多层穿透路径",
			node: &repositories.RelatedEntityNode{
				RelationType: models.RelationTypeBranch,
				Direction:    "outgoing",
				Path: []repositories.RelatedEdge{
					{FromID: 1, ToID: 10, RelationType: models.RelationTypeActualController},
					{FromID: 10, ToID: 20, RelationType: models.RelationTypeSubsidiary},
					{FromID: 20, ToID: 30, RelationType: models.RelationTypeBranch},
				},
			},
			expected: "[ACTUAL_CONTROLLER(→)10 → SUBSIDIARY(→)20 → BRANCH(→)30]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildPathDescription(tt.node)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDetectByRelationship_Recursive3Layers(t *testing.T) {
	mockRepo := &mockRecursiveRepo{
		nodes: []*repositories.RelatedEntityNode{
			{
				EntityID:     10,
				Depth:        1,
				RelationType: models.RelationTypeActualController,
				Direction:    "outgoing",
				Entity:       &models.Entity{ID: 10, Name: "张三控股集团", EntityType: models.EntityTypeLegalPerson},
				Path:         []repositories.RelatedEdge{{FromID: 1, ToID: 10, RelationType: models.RelationTypeActualController}},
			},
			{
				EntityID:     20,
				Depth:        2,
				RelationType: models.RelationTypeSubsidiary,
				Direction:    "outgoing",
				Entity:       &models.Entity{ID: 20, Name: "子公司A", EntityType: models.EntityTypeLegalPerson},
				Path: []repositories.RelatedEdge{
					{FromID: 1, ToID: 10, RelationType: models.RelationTypeActualController},
					{FromID: 10, ToID: 20, RelationType: models.RelationTypeSubsidiary},
				},
			},
			{
				EntityID:     30,
				Depth:        3,
				RelationType: models.RelationTypeBranch,
				Direction:    "outgoing",
				Entity:       &models.Entity{ID: 30, Name: "孙公司B", EntityType: models.EntityTypeLegalPerson},
				Path: []repositories.RelatedEdge{
					{FromID: 1, ToID: 10, RelationType: models.RelationTypeActualController},
					{FromID: 10, ToID: 20, RelationType: models.RelationTypeSubsidiary},
					{FromID: 20, ToID: 30, RelationType: models.RelationTypeBranch},
				},
			},
		},
		relatedCases: map[uint][]*models.Case{
			10: {{ID: 100, Title: "相关案件X", Status: "IN_PROGRESS", CaseType: "民事案件"}},
			20: {{ID: 200, Title: "相关案件Y", Status: "ACTIVE", CaseType: "商事案件"}},
		},
	}

	svc := &conflictCheckService{entityRepo: mockRepo}
	conflicts := svc.detectByRelationship(context.Background(), EntityCheckInfo{
		EntityID: 1, EntityName: "测试客户公司", PartyType: "CLIENT",
	}, 3, 999)

	assert.Equal(t, 2, len(conflicts))

	// 实际控制人 -> HIGH
	assert.Equal(t, "HIGH", conflicts[0].RiskLevel)
	assert.Contains(t, conflicts[0].Description, "测试客户公司")
	assert.Contains(t, conflicts[0].Description, "张三控股集团")
	assert.Contains(t, conflicts[0].MatchReason, "ACTUAL_CONTROLLER")

	// 子公司 -> MEDIUM
	assert.Equal(t, "MEDIUM", conflicts[1].RiskLevel)
	assert.Contains(t, conflicts[1].Description, "子公司A")
}

func TestDetectByRelationship_SpouseAndFamily_HighRisk(t *testing.T) {
	tests := []struct {
		name         string
		relationType models.RelationType
		entityName   string
	}{
		{"配偶关系", models.RelationTypeSpouse, "李四"},
		{"家庭成员", models.RelationTypeFamilyMember, "赵六"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockRecursiveRepo{
				nodes: []*repositories.RelatedEntityNode{
					{
						EntityID:     50,
						Depth:        1,
						RelationType: tt.relationType,
						Direction:    "outgoing",
						Entity:       &models.Entity{ID: 50, Name: tt.entityName, EntityType: models.EntityTypeIndividual},
						Path:         []repositories.RelatedEdge{{FromID: 1, ToID: 50, RelationType: tt.relationType}},
					},
				},
				relatedCases: map[uint][]*models.Case{
					50: {{ID: 300, Title: "关联案件", Status: "IN_PROGRESS", CaseType: "民事案件"}},
				},
			}

			svc := &conflictCheckService{entityRepo: mockRepo}
			conflicts := svc.detectByRelationship(context.Background(), EntityCheckInfo{
				EntityID: 1, EntityName: "王五", PartyType: "CLIENT",
			}, 1, 999)

			assert.Equal(t, 1, len(conflicts))
			assert.Equal(t, "HIGH", conflicts[0].RiskLevel)
			assert.Contains(t, conflicts[0].Description, tt.entityName)
			assert.Contains(t, conflicts[0].MatchReason, string(tt.relationType))
		})
	}
}

func TestDetectByRelationship_ZeroEntityID(t *testing.T) {
	mockRepo := &mockRecursiveRepo{}
	svc := &conflictCheckService{entityRepo: mockRepo}
	conflicts := svc.detectByRelationship(context.Background(), EntityCheckInfo{EntityID: 0}, 3, 999)
	assert.Empty(t, conflicts)
}

func TestDetectByRelationship_DedupSameEntity(t *testing.T) {
	// 同一实体通过不同路径到达应去重
	mockRepo := &mockRecursiveRepo{
		nodes: []*repositories.RelatedEntityNode{
			{
				EntityID:     10,
				Depth:        1,
				RelationType: models.RelationTypeMajorShareholder,
				Entity:       &models.Entity{ID: 10, Name: "同一实体"},
				Path:         []repositories.RelatedEdge{{FromID: 1, ToID: 10, RelationType: models.RelationTypeMajorShareholder}},
			},
			{
				EntityID:     10,
				Depth:        2,
				RelationType: models.RelationTypeRelatedParty,
				Entity:       &models.Entity{ID: 10, Name: "同一实体"},
				Path: []repositories.RelatedEdge{
					{FromID: 1, ToID: 5, RelationType: models.RelationTypeSubsidiary},
					{FromID: 5, ToID: 10, RelationType: models.RelationTypeRelatedParty},
				},
			},
		},
		relatedCases: map[uint][]*models.Case{
			10: {{ID: 100, Title: "案件", Status: "IN_PROGRESS"}},
		},
	}

	svc := &conflictCheckService{entityRepo: mockRepo}
	conflicts := svc.detectByRelationship(context.Background(), EntityCheckInfo{EntityID: 1}, 3, 999)
	assert.Equal(t, 1, len(conflicts))
}

func TestDetectByRelationship_NoConflictingCases(t *testing.T) {
	// 关联实体存在但无活跃案件
	mockRepo := &mockRecursiveRepo{
		nodes: []*repositories.RelatedEntityNode{
			{
				EntityID:     10,
				Depth:        1,
				RelationType: models.RelationTypeParentCompany,
				Entity:       &models.Entity{ID: 10, Name: "母公司"},
				Path:         []repositories.RelatedEdge{{FromID: 1, ToID: 10, RelationType: models.RelationTypeParentCompany}},
			},
		},
		relatedCases: map[uint][]*models.Case{},
	}

	svc := &conflictCheckService{entityRepo: mockRepo}
	conflicts := svc.detectByRelationship(context.Background(), EntityCheckInfo{EntityID: 1}, 3, 999)
	assert.Empty(t, conflicts)
}
