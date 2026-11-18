package handlers

import (
	"context"
	"testing"

	"law-oa-go/internal/models"
)

func TestZZProbeWallModel(t *testing.T) {
	db := newMatterCollabDB(t)
	var total, visible int64
	err1 := db.WithContext(context.Background()).Model(&models.Case{}).Where("client_id = ? AND deleted_at IS NULL", uint(10)).Count(&total).Error
	err2 := db.WithContext(context.Background()).Model(&models.Case{}).Where(`client_id = ? AND deleted_at IS NULL AND (
		ethical_wall_enabled = ?
		OR EXISTS (
			SELECT 1 FROM case_ethical_wall_whitelist wall_access
			WHERE wall_access.case_id = cases.id AND wall_access.user_id = ?
		)
	)`, uint(10), false, uint(6)).Count(&visible).Error
	t.Logf("total=%d err1=%v visible=%d err2=%v", total, err1, visible, err2)
}
