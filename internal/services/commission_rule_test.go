package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// ============================================================================
// Sprint 2 - T10: 数据一致性校验测试
// ============================================================================

// TestCommissionRateFallback 校验分成规则fallback机制
func TestCommissionRateFallback(t *testing.T) {
	svc := &CommissionService{}

	tests := []struct {
		name     string
		role     string
		amount   float64
		expected float64
	}{
		// 案源提成
		{"案源-小额", "source", 5000, 10},
		{"案源-中额", "source", 30000, 15},
		{"案源-大额", "source", 80000, 20},
		{"案源-特大额", "source", 200000, 30},

		// 主办律师提成
		{"主办律师-小额", "lawyer", 5000, 20},
		{"主办律师-中额", "lawyer", 30000, 30},
		{"主办律师-大额", "lawyer", 80000, 40},
		{"主办律师-特大额", "lawyer", 200000, 50},

		// 协办律师提成
		{"协办律师-小额", "assistant", 5000, 5},
		{"协办律师-中额", "assistant", 30000, 8},
		{"协办律师-大额", "assistant", 80000, 12},
		{"协办律师-特大额", "assistant", 200000, 15},

		// 边界值
		{"案源-边界10000", "source", 10000, 15},
		{"案源-边界50000", "source", 50000, 20},
		{"案源-边界100000", "source", 100000, 30},

		// 未知角色
		{"未知角色", "unknown", 50000, 0},

		// 零金额
		{"零金额-案源", "source", 0, 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rate := svc.getCommissionRate(tt.role, tt.amount)
			assert.Equal(t, tt.expected, rate, "角色=%s 金额=%.0f 提成比例应为%.0f%%",
				tt.role, tt.amount, tt.expected)
		})
	}
}

// TestCommissionRuleAmountRange 校验金额区间匹配逻辑
func TestCommissionRuleAmountRange(t *testing.T) {
	svc := &CommissionService{}

	// 测试所有角色的金额区间连续性
	roles := []string{"source", "lawyer", "assistant"}
	amounts := []float64{0, 100, 5000, 9999, 10000, 10001, 50000, 50001, 100000, 100001, 500000}

	for _, role := range roles {
		t.Run("角色_"+role, func(t *testing.T) {
			prevRate := -1.0
			for _, amount := range amounts {
				rate := svc.getCommissionRate(role, amount)
				assert.GreaterOrEqual(t, rate, 0.0, "角色=%s 金额=%.0f 提成比例不应为负", role, amount)
				// 确认金额越大比例越高（非递减）
				if prevRate >= 0 {
					assert.GreaterOrEqual(t, rate, prevRate, "角色=%s 金额从%.0f增到%.0f时，比例不应降低",
						role, amount-100, amount)
				}
				prevRate = rate
			}
		})
	}
}

// TestCommissionRuleRoleConsistency 校验各角色提成比例的合理性
func TestCommissionRuleRoleConsistency(t *testing.T) {
	svc := &CommissionService{}
	amount := 50000.0

	sourceRate := svc.getCommissionRate("source", amount)
	lawyerRate := svc.getCommissionRate("lawyer", amount)
	assistantRate := svc.getCommissionRate("assistant", amount)

	// 业务规则：主办律师提成 > 案源提成 > 协办律师提成
	assert.Greater(t, lawyerRate, sourceRate,
		"主办律师提成(%.0f%%)应大于案源提成(%.0f%%)", lawyerRate, sourceRate)
	assert.Greater(t, sourceRate, assistantRate,
		"案源提成(%.0f%%)应大于协办律师提成(%.0f%%)", sourceRate, assistantRate)
}

// TestCommissionRuleUpperBound 校验提成比例上限
func TestCommissionRuleUpperBound(t *testing.T) {
	svc := &CommissionService{}

	roles := []string{"source", "lawyer", "assistant"}
	maxAmount := 1000000.0

	for _, role := range roles {
		t.Run("角色_"+role, func(t *testing.T) {
			rate := svc.getCommissionRate(role, maxAmount)
			assert.LessOrEqual(t, rate, 50.0,
				"角色=%s 最大金额的提成比例(%.0f%%)不应超过50%%", role, rate)
		})
	}
}
