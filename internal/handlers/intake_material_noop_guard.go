package handlers

import (
	"bytes"
	"encoding/json"
	"reflect"

	"gorm.io/gorm"
	"law-oa-go/internal/services"
)

// Legacy forms can echo materials on a native PUT. Only accept an unchanged
// projection of every stored row; even a successful echo performs no material writes.
func validateIntakeMaterialEcho(tx *gorm.DB, intakeID string, echo []map[string]interface{}) error {
	var stored []map[string]interface{}
	if err := tx.Table("case_materials").Where("intake_id = ?", intakeID).Order("id").Find(&stored).Error; err != nil {
		return err
	}
	mismatch := services.NewSubjectWorkflowError("INTAKE_MATERIAL_ECHO_MISMATCH", "材料与已存记录不一致，请重新核对草稿")
	if len(echo) != len(stored) {
		return mismatch
	}
	byName := make(map[string][]map[string]interface{}, len(stored))
	for _, row := range stored {
		name, _ := row["name"].(string)
		byName[name] = append(byName[name], row)
	}
	for _, item := range echo {
		name, ok := item["name"].(string)
		rows := byName[name]
		if !ok || name == "" || len(rows) == 0 {
			return mismatch
		}
		// A partial echo cannot identify one of several identically named rows.
		if len(rows) > 1 && item["id"] == nil {
			return mismatch
		}
		matched := false
		for i, row := range rows {
			if matchesMaterialFields(item, row) {
				byName[name] = append(rows[:i], rows[i+1:]...)
				matched = true
				break
			}
		}
		if !matched {
			return mismatch
		}
	}
	return nil
}

func matchesMaterialFields(item, stored map[string]interface{}) bool {
	for key, value := range item {
		if _, allowed := allowedCaseMaterialFields[key]; !allowed && key != "id" {
			return false
		}
		want, exists := stored[key]
		if !exists {
			return false
		}
		if key == "metadata" {
			obj, ok := value.(map[string]interface{})
			old, err := decodeIntakeMetadata(want)
			if !ok || err != nil || !reflect.DeepEqual(obj, old) {
				return false
			}
			continue
		}
		got, err := json.Marshal(value)
		if err != nil {
			return false
		}
		expected, err := json.Marshal(want)
		if err != nil || !bytes.Equal(got, expected) {
			return false
		}
	}
	return true
}
