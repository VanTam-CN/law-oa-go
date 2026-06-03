package repositories

import (
	"context"
	"time"

	"gorm.io/gorm"
	"law-oa-go/internal/models"
)

// GetPaymentAggregation 获取回款聚合统计（通过SQL聚合，避免加载全部记录）
func (r *PaymentRepositoryImpl) GetPaymentAggregation(ctx context.Context, monthStart, monthEnd string) (total int64, pendingCount int64, confirmedCount int64, rejectedCount int64, totalAmount float64, monthAmount float64, pendingAmount float64, err error) {
	type statusAgg struct {
		Status string
		Count  int64
		Sum    float64
	}
	var aggs []statusAgg

	err = r.db.WithContext(ctx).
		Model(&models.Payment{}).
		Select("status, COUNT(*) as count, COALESCE(SUM(amount), 0) as sum").
		Group("status").
		Scan(&aggs).Error
	if err != nil {
		return 0, 0, 0, 0, 0, 0, 0, err
	}

	for _, a := range aggs {
		total += a.Count
		switch a.Status {
		case "pending":
			pendingCount = a.Count
			pendingAmount = a.Sum
		case "confirmed":
			confirmedCount = a.Count
			totalAmount = a.Sum
		case "rejected":
			rejectedCount = a.Count
		}
	}

	if monthStart != "" && monthEnd != "" {
		var monthSum struct {
			Sum float64
		}
		err = r.db.WithContext(ctx).
			Model(&models.Payment{}).
			Select("COALESCE(SUM(amount), 0) as sum").
			Where("status = ? AND payment_date >= ? AND payment_date < ?", "confirmed", monthStart, monthEnd).
			Scan(&monthSum).Error
		if err != nil {
			return
		}
		monthAmount = monthSum.Sum
	}

	return
}

// ensure compile-time interface satisfaction
var _ = gorm.DB{}
var _ = time.Time{}
var _ = models.Payment{}
