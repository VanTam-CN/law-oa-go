package models

import "time"

type CaseTaskChange struct {
	ID            uint       `json:"id" gorm:"primarykey"`
	CreatedAt     time.Time  `json:"created_at"`
	TaskID        uint       `json:"task_id" gorm:"column:task_id;not null;index"`
	CaseID        uint       `json:"case_id" gorm:"column:case_id;not null;index"`
	ChangeType    string     `json:"change_type" gorm:"column:change_type;size:20;not null"`
	ActorID       uint       `json:"actor_id" gorm:"column:actor_id;not null"`
	OldAssigneeID *uint      `json:"old_assignee_id,omitempty" gorm:"column:old_assignee_id"`
	NewAssigneeID *uint      `json:"new_assignee_id,omitempty" gorm:"column:new_assignee_id"`
	OldDueDate    *time.Time `json:"old_due_date,omitempty" gorm:"column:old_due_date"`
	NewDueDate    *time.Time `json:"new_due_date,omitempty" gorm:"column:new_due_date"`
	// Version snapshots the case_tasks.version at change time. The gorm
	// column tag keeps dev-mode AutoMigrate and the SQL migration aligned;
	// it is audit data only and is never an UPDATE target.
	Version    *uint  `json:"version,omitempty" gorm:"column:version"`
	OldVersion uint   `json:"old_version" gorm:"column:old_version;not null;default:0"`
	NewVersion uint   `json:"new_version" gorm:"column:new_version;not null;default:0"`
	Reason     string `json:"reason,omitempty" gorm:"column:reason;type:text"`
}

func (CaseTaskChange) TableName() string { return "case_task_changes" }

const (
	CaseTaskChangeReassign      = "reassign"
	CaseTaskChangeDueDateChange = "due_date_change"
	CaseTaskChangeCancel        = "cancel"
)
