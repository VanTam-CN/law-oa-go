package models

import (
	"time"
)

// FeeTemplate 费率模板
type FeeTemplate struct {
	ID                   uint      `json:"id" gorm:"primarykey"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
	Name                 string    `json:"name" gorm:"size:100;not null;comment:模板名称"`
	CaseType             string    `json:"case_type" gorm:"size:50;not null;index:idx_case_type;comment:适用案件类型: litigation/non_litigation/consulting"`
	BillingType          string    `json:"billing_type" gorm:"size:50;not null;comment:计费模式: hourly/fixed/hybrid/retainer"`
	BaseRates            JSON      `json:"base_rates" gorm:"type:jsonb;not null;comment:按角色定义基础费率: {source: 0.15, lawyer: 0.30, assistant: 0.10}"`
	PerformanceBonusRate float64   `json:"performance_bonus_rate" gorm:"type:decimal(5,2);default:0;comment:绩效奖金比例"`
	MinAmount            float64   `json:"min_amount" gorm:"type:decimal(15,2);default:0;comment:最小适用金额"`
	MaxAmount            float64   `json:"max_amount" gorm:"type:decimal(15,2);default:0;comment:最大适用金额"`
	CostRate             float64   `json:"cost_rate" gorm:"type:decimal(5,2);default:0;comment:成本扣除比例"`
	Active               bool      `json:"active" gorm:"default:true;index:idx_active;comment:是否启用"`
}

func (FeeTemplate) TableName() string {
	return "fee_templates"
}
