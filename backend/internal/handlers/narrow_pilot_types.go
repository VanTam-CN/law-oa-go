package handlers

import (
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// The first slice only admits the description of an existing owned draft.
type NarrowPilotHandler struct {
	h        *DemoAggregateHandler
	db       *gorm.DB
	allowed  map[string]bool
	mu       sync.Mutex
	contexts map[string]*npContext
}

type npError struct {
	Code   string
	Status int
}

func (e *npError) Error() string { return e.Code }
func npFail(c *gin.Context, e *npError) {
	message := map[string]string{
		"INTAKE_DRAFT_REVISION_CONFLICT": "草稿版本已变化，本次未写入。未保存输入已保留，请重新加载后逐字段核对。",
		"UPDATE_PROTOCOL_REQUIRED":       "此草稿已有来源历史，请刷新页面；首切片仅支持摘要的局部保存。",
		"SOURCE_SELECTION_STALE":         "候选或来源已失效，请重新整理并核对；当前文字保留。",
		"SAVE_GUARD_BUSY":                "原保存仍在处理，结果待核实；请沿用原键回查或明确结束。",
		"SAVE_REQUEST_ABANDONED":         "原保存已明确结束，迟到请求未写入。",
		"MATERIAL_EDIT_NOT_SUPPORTED":    "摘要整理不接受材料变更，请使用原生材料入口。",
	}[e.Code]
	if message == "" {
		message = "操作未完成，请保留输入并核对当前草稿或保存结果"
	}
	c.JSON(e.Status, gin.H{"success": false, "error": gin.H{"code": e.Code, "message": message}})
}

type npFieldSelection struct {
	CandidateID   string `json:"candidate_id"`
	RequestedMode string `json:"requested_mode"`
	Reviewed      bool   `json:"reviewed"`
}
type npSourceSelection struct {
	ContextID              string                      `json:"context_id"`
	ExpectedContextVersion int                         `json:"expected_context_version"`
	Fields                 map[string]npFieldSelection `json:"fields"`
}
type npSourcePack struct {
	ContextID        string     `json:"context_id"`
	ContextVersion   int        `json:"context_version"`
	CandidateID      string     `json:"candidate_id"`
	OriginalRevision string     `json:"original_revision"`
	Description      string     `json:"description"`
	Sources          []npSource `json:"sources"`
	ReviewedBy       string     `json:"reviewed_by"`
	ReviewedAt       time.Time  `json:"reviewed_at"`
}
type npSource struct {
	ID         string    `json:"id"`
	Kind       string    `json:"kind"`
	Text       string    `json:"text"`
	Turn       int       `json:"turn"`
	EnteredBy  string    `json:"entered_by"`
	AcceptedAt time.Time `json:"accepted_at"`
}
