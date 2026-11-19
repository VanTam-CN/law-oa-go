package handlers

import (
	"crypto/sha256"
	"database/sql"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"law-oa-go/internal/common"
)

const npDescriptionSourceRevisionKey = "np_description_source_revision"

type npWriteRequest struct {
	Description *string            `json:"description"`
	ClearFields []string           `json:"clear_fields"`
	Changed     []string           `json:"changed_fields"`
	Source      *npSourceSelection `json:"source_selection"`
	Revision    string             `json:"expected_revision"`
}
type npWriteResult struct {
	ID             string `json:"id"`
	Revision       string `json:"revision"`
	State          string `json:"state"`
	RequestID      string `json:"request_id"`
	SourceRevision string `json:"source_revision"`
}
type npReceiptRow struct {
	ID        uint64
	ActorID   string
	RequestID string
	Action    string
	IntakeID  string
	State     string
	Digest    *string `gorm:"column:request_digest"`
	ResultRev *string `gorm:"column:result_revision"`
	SourceRev *string `gorm:"column:source_revision"`
	CreatedAt time.Time
}

func (npReceiptRow) TableName() string { return "case_intake_write_requests" }

type npSnapshotRow struct {
	ID        uint64
	IntakeID  string
	Revision  string `gorm:"column:saved_revision"`
	Pack      []byte `gorm:"column:serverpack"`
	CreatedAt time.Time
}

func (npSnapshotRow) TableName() string { return "case_intake_source_snapshots" }

func (n *NarrowPilotHandler) strictReady(c *gin.Context) *npError {
	db := n.db.WithContext(c.Request.Context())
	if n.db.Dialector.Name() != "postgres" || !db.Migrator().HasTable(&npReceiptRow{}) || !db.Migrator().HasTable(&npSnapshotRow{}) ||
		!db.Migrator().HasIndex(&npReceiptRow{}, "np_write_actor_request_unique") || !db.Migrator().HasIndex(&npSnapshotRow{}, "np_source_intake_revision_unique") {
		return &npError{"SAVE_SCHEMA_REQUIRED", 503}
	}
	return nil
}

// Every legacy caller is checked, including compatibility tokens and disabled assistants.
func (n *NarrowPilotHandler) HandleUpdate(c *gin.Context, raw map[string]interface{}) bool {
	if _, err := uuid.Parse(c.Param("id")); err != nil {
		npFail(c, &npError{"INVALID_REQUEST", 400})
		return true
	}
	if _, strict := raw["changed_fields"]; !strict {
		if !n.db.WithContext(c.Request.Context()).Migrator().HasTable(&npSnapshotRow{}) {
			if len(n.allowed) > 0 {
				npFail(c, &npError{"SAVE_SCHEMA_REQUIRED", 503})
				return true
			}
			return false
		}
		var count int64
		if err := n.db.WithContext(c.Request.Context()).Model(&npSnapshotRow{}).Where("intake_id=?", c.Param("id")).Count(&count).Error; err != nil {
			npFail(c, &npError{"SAVE_RESULT_UNKNOWN", 503})
			return true
		}
		if count > 0 {
			if _, e := n.authorize(c, n.db, c.Param("id"), false); e != nil {
				npFail(c, e)
			} else {
				npFail(c, &npError{"UPDATE_PROTOCOL_REQUIRED", 409})
			}
			return true
		}
		return false
	}
	actor, _, _, e := n.session(c, n.db)
	if e != nil {
		npFail(c, e)
		return true
	}
	req, e := npParseUpdate(raw)
	if e != nil {
		npFail(c, e)
		return true
	}
	key := c.GetHeader("Idempotency-Key")
	if !npValidKey(key) {
		npFail(c, &npError{"IDEMPOTENCY_KEY_REQUIRED", 400})
		return true
	}
	if e := n.strictReady(c); e != nil {
		npFail(c, e)
		return true
	}
	result, e := n.strictUpdate(c, actor, key, req)
	if e != nil {
		npFail(c, e)
	} else {
		common.APISuccess(c, result)
	}
	return true
}

func npParseUpdate(raw map[string]interface{}) (*npWriteRequest, *npError) {
	for _, key := range []string{"materials", "material_changes"} {
		if _, ok := raw[key]; ok {
			return nil, &npError{"MATERIAL_EDIT_NOT_SUPPORTED", 400}
		}
	}
	if fields, ok := raw["changed_fields"].([]interface{}); ok {
		for _, v := range fields {
			if v == "materials" {
				return nil, &npError{"MATERIAL_EDIT_NOT_SUPPORTED", 400}
			}
		}
	}
	encoded, err := json.Marshal(raw)
	if err != nil {
		return nil, &npError{"AMBIGUOUS_UPDATE", 400}
	}
	d := json.NewDecoder(strings.NewReader(string(encoded)))
	d.DisallowUnknownFields()
	var req npWriteRequest
	if d.Decode(&req) != nil || len(req.Changed) != 1 || req.Changed[0] != "description" || strings.TrimSpace(req.Revision) == "" {
		return nil, &npError{"AMBIGUOUS_UPDATE", 400}
	}
	if v, ok := raw["source_selection"]; ok && v == nil {
		return nil, &npError{"SOURCE_SELECTION_INVALID", 400}
	}
	_, hasDescription := raw["description"]
	if len(req.ClearFields) > 0 {
		if len(req.ClearFields) != 1 || req.ClearFields[0] != "description" || hasDescription || req.Source != nil {
			return nil, &npError{"AMBIGUOUS_UPDATE", 400}
		}
	} else if !hasDescription || req.Description == nil || strings.TrimSpace(*req.Description) == "" {
		return nil, &npError{"AMBIGUOUS_UPDATE", 400}
	}
	if v, ok := raw["clear_fields"]; ok && v == nil {
		return nil, &npError{"AMBIGUOUS_UPDATE", 400}
	}
	return &req, nil
}
func npValidKey(key string) bool {
	parsed, err := uuid.Parse(key)
	return err == nil && parsed.String() == key
}
func npAdvisoryKey(actor, key string) int64 {
	sum := sha256.Sum256([]byte(actor + "\x00" + key))
	return int64(binary.BigEndian.Uint64(sum[:8]))
}
func npRequestDigest(id string, req *npWriteRequest) string {
	raw, _ := json.Marshal([]interface{}{"update", id, req})
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}
func npDeref(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}
func npResult(row *npReceiptRow) *npWriteResult {
	return &npWriteResult{ID: row.IntakeID, Revision: npDeref(row.ResultRev), State: row.State, RequestID: row.RequestID, SourceRevision: npDeref(row.SourceRev)}
}

// Save and explicit abandon share this database lock and terminal unique key.
func (n *NarrowPilotHandler) keyTransaction(c *gin.Context, actor, key string, fn func(*gorm.DB) error) *npError {
	err := n.db.WithContext(c.Request.Context()).Transaction(func(tx *gorm.DB) error {
		var locked bool
		if err := tx.Raw("SELECT pg_try_advisory_xact_lock(?)", npAdvisoryKey(actor, key)).Scan(&locked).Error; err != nil {
			return err
		}
		if !locked {
			return &npError{"SAVE_GUARD_BUSY", 409}
		}
		if current, _, _, e := n.session(c, tx); e != nil {
			return e
		} else if current != actor {
			return &npError{"ACCESS_SESSION_REQUIRED", 401}
		}
		if err := fn(tx); err != nil {
			return err
		}
		_, _, _, e := n.session(c, tx)
		if e != nil {
			return e
		}
		return nil
	}, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		var e *npError
		if errors.As(err, &e) {
			return e
		}
		return &npError{"SAVE_RESULT_UNKNOWN", 503}
	}
	return nil
}
func (n *NarrowPilotHandler) strictUpdate(c *gin.Context, actor, key string, req *npWriteRequest) (*npWriteResult, *npError) {
	id := c.Param("id")
	digest := npRequestDigest(id, req)
	var result *npWriteResult
	var pack *npSourcePack
	e := n.keyTransaction(c, actor, key, func(tx *gorm.DB) error {
		var prior npReceiptRow
		if err := tx.Where("actor_id=? AND request_id=?", actor, key).Take(&prior).Error; err == nil {
			if _, e := n.authorize(c, tx, prior.IntakeID, false); e != nil {
				return e
			}
			if prior.State == "abandoned" {
				return &npError{"SAVE_REQUEST_ABANDONED", 409}
			}
			if prior.Digest == nil || *prior.Digest != digest {
				return &npError{"IDEMPOTENCY_PAYLOAD_MISMATCH", 409}
			}
			result = npResult(&prior)
			return nil
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		row, e := n.authorize(c, tx, id, true)
		if e != nil {
			return e
		}
		if intakeDraftRevisionFor(id, row["metadata"]) != req.Revision {
			return &npError{"INTAKE_DRAFT_REVISION_CONFLICT", 409}
		}
		desired := ""
		if req.Description != nil {
			desired = *req.Description
		}
		metadata := objectValue(row["metadata"])
		sourceRevision, _ := metadata[npDescriptionSourceRevisionKey].(string)
		revision := req.Revision
		state := "no_change"
		if req.Source != nil {
			pack, e = n.prepareSource(c, row, req.Source, desired)
			if e != nil {
				return e
			}
		}
		existingDescription := ""
		switch value := row["description"].(type) {
		case string:
			existingDescription = value
		case []byte:
			existingDescription = string(value)
		case nil:
		default:
			return &npError{"INTAKE_DESCRIPTION_UNAVAILABLE", 503}
		}
		if existingDescription != desired || pack != nil {
			revision = uuid.NewString()
			state = "committed"
			// Update reserved JSON paths in PostgreSQL; decoding unknown numbers through float64 would round them.
			metadataUpdate := gorm.Expr("jsonb_set(COALESCE(metadata, '{}'::jsonb) - ?, '{draft_revision}', to_jsonb(?::text), true)", npDescriptionSourceRevisionKey, revision)
			if pack != nil {
				encoded, err := json.Marshal(pack)
				if err != nil {
					return err
				}
				if err := tx.Exec("INSERT INTO case_intake_source_snapshots (intake_id,saved_revision,serverpack) VALUES (?,?,?::jsonb)", id, revision, string(encoded)).Error; err != nil {
					return err
				}
				sourceRevision = revision
				metadataUpdate = gorm.Expr("jsonb_set(jsonb_set(COALESCE(metadata, '{}'::jsonb), '{draft_revision}', to_jsonb(?::text), true), '{np_description_source_revision}', to_jsonb(?::text), true)", revision, revision)
			} else {
				sourceRevision = ""
			}
			if err := tx.Table("case_intakes").Where("id=?", id).Updates(map[string]interface{}{"description": desired, "metadata": metadataUpdate, "updated_at": time.Now()}).Error; err != nil {
				return err
			}
		}
		rowReceipt := npReceiptRow{ActorID: actor, RequestID: key, Action: "update", IntakeID: id, State: state, Digest: &digest, ResultRev: &revision, SourceRev: &sourceRevision, CreatedAt: time.Now()}
		if err := tx.Create(&rowReceipt).Error; err != nil {
			return err
		}
		result = npResult(&rowReceipt)
		return nil
	})
	n.finishSource(pack, e == nil)
	return result, e
}

func (n *NarrowPilotHandler) WriteResult(c *gin.Context) {
	actor, _, _, e := n.session(c, n.db)
	if e != nil {
		npFail(c, e)
		return
	}
	key := c.Param("request_id")
	if !npValidKey(key) {
		npFail(c, &npError{"INVALID_REQUEST", 400})
		return
	}
	if e := n.strictReady(c); e != nil {
		npFail(c, e)
		return
	}
	var row npReceiptRow
	err := n.db.WithContext(c.Request.Context()).Where("actor_id=? AND request_id=?", actor, key).Take(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		common.APISuccess(c, gin.H{"state": "unknown", "request_id": key})
		return
	}
	if err != nil {
		npFail(c, &npError{"SAVE_RESULT_UNKNOWN", 503})
		return
	}
	if _, e := n.authorize(c, n.db, row.IntakeID, false); e != nil {
		npFail(c, e)
		return
	}
	common.APISuccess(c, npResult(&row))
}
func (n *NarrowPilotHandler) Abandon(c *gin.Context) {
	var body struct {
		Action   string `json:"action"`
		IntakeID string `json:"intake_id"`
	}
	if !npDecode(c, &body) {
		return
	}
	key := c.Param("request_id")
	if !npValidKey(key) || body.Action != "update" || body.IntakeID == "" {
		npFail(c, &npError{"INVALID_REQUEST", 400})
		return
	}
	actor, _, _, e := n.session(c, n.db)
	if e != nil {
		npFail(c, e)
		return
	}
	if e := n.strictReady(c); e != nil {
		npFail(c, e)
		return
	}
	var result *npWriteResult
	e = n.keyTransaction(c, actor, key, func(tx *gorm.DB) error {
		var row npReceiptRow
		if err := tx.Where("actor_id=? AND request_id=?", actor, key).Take(&row).Error; err == nil {
			if _, e := n.authorize(c, tx, row.IntakeID, false); e != nil {
				return e
			}
			if row.IntakeID != body.IntakeID || row.Action != body.Action {
				return &npError{"IDEMPOTENCY_PAYLOAD_MISMATCH", 409}
			}
			result = npResult(&row)
			return nil
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		if _, e := n.authorize(c, tx, body.IntakeID, false); e != nil {
			return e
		}
		row = npReceiptRow{ActorID: actor, RequestID: key, Action: "update", IntakeID: body.IntakeID, State: "abandoned", CreatedAt: time.Now()}
		if err := tx.Create(&row).Error; err != nil {
			return err
		}
		result = npResult(&row)
		return nil
	})
	if e != nil {
		npFail(c, e)
		return
	}
	common.APISuccess(c, result)
}
func (n *NarrowPilotHandler) SourceSnapshot(c *gin.Context) {
	row, e := n.authorize(c, n.db, c.Param("id"), false)
	if e != nil {
		npFail(c, e)
		return
	}
	if e := n.strictReady(c); e != nil {
		npFail(c, e)
		return
	}
	revision := c.Param("revision")
	current := intakeDraftRevisionFor(c.Param("id"), row["metadata"])
	sourceRevision := revision
	if revision == current {
		sourceRevision, _ = objectValue(row["metadata"])[npDescriptionSourceRevisionKey].(string)
	}
	var source *npSourcePack
	if sourceRevision != "" {
		var saved npSnapshotRow
		err := n.db.WithContext(c.Request.Context()).Where("intake_id=? AND saved_revision=?", c.Param("id"), sourceRevision).Take(&saved).Error
		if err != nil || json.Unmarshal(saved.Pack, &source) != nil {
			npFail(c, &npError{"SOURCE_UNAVAILABLE", 409})
			return
		}
	}
	common.APISuccess(c, gin.H{"revision": revision, "source_revision": sourceRevision, "source": source, "current": revision == current})
}
