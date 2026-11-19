package handlers

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
	"law-oa-go/internal/common"
	"law-oa-go/internal/middleware"
	"law-oa-go/internal/models"
	"law-oa-go/internal/services"
)

func (n *NarrowPilotHandler) RequestBudget(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 45*time.Second)
	defer cancel()
	c.Request = c.Request.WithContext(ctx)
	c.Header("Cache-Control", "private, no-store")
	c.Next()
}

type npCandidate struct {
	ID          string     `json:"id"`
	Description string     `json:"description"`
	Sources     []npSource `json:"sources"`
}
type npTurnReplay struct {
	Body    string
	Version int
}
type npContext struct {
	ID                                      string       `json:"id"`
	Version                                 int          `json:"context_version"`
	Expires                                 time.Time    `json:"expires_at"`
	Candidate                               *npCandidate `json:"candidate"`
	Actor, Access, Intake, Client, Revision string
	Created, Touched, AccessExpiry          time.Time
	Sources                                 []npSource
	Replays                                 map[string]npTurnReplay
	Saving                                  bool
	timer                                   *time.Timer
}

func NewNarrowPilotHandler(h *DemoAggregateHandler) *NarrowPilotHandler {
	n := &NarrowPilotHandler{h: h, db: h.db.Session(&gorm.Session{Logger: logger.Default.LogMode(logger.Silent)}), allowed: map[string]bool{}, contexts: map[string]*npContext{}}
	for _, id := range strings.Split(os.Getenv("NP_ASSISTANT_ACTORS"), ",") {
		if parsed, err := strconv.ParseUint(strings.TrimSpace(id), 10, 64); err == nil && parsed > 0 {
			n.allowed[strconv.FormatUint(parsed, 10)] = true
		}
	}
	h.narrowPilot = n
	return n
}

// Sensitive endpoints require a registered access token, even while legacy OA accepts compatibility tokens.
func (n *NarrowPilotHandler) session(c *gin.Context, db *gorm.DB) (string, string, time.Time, *npError) {
	c.Header("Cache-Control", "private, no-store")
	claims, err := middleware.ValidateToken(strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer "))
	if err != nil || claims.Type != "access" || claims.UUID == "" || claims.ExpiresAt == nil {
		return "", "", time.Time{}, &npError{"ACCESS_SESSION_REQUIRED", http.StatusUnauthorized}
	}
	actor := strconv.FormatUint(uint64(claims.UserID), 10)
	if current, ok := middleware.GetCurrentUserID(c); !ok || current != claims.UserID {
		return "", "", time.Time{}, &npError{"ACCESS_SESSION_REQUIRED", 401}
	}
	var s models.AuthTokenSession
	if err := db.WithContext(c.Request.Context()).Where("access_token_uuid = ? AND user_id = ?", claims.UUID, claims.UserID).Take(&s).Error; err != nil || s.AccessRevokedAt != nil || s.RevokedAt != nil || s.DeviceRevokedAt != nil {
		return "", "", time.Time{}, &npError{"ACCESS_SESSION_REQUIRED", 401}
	}
	var user models.User
	if err := db.WithContext(c.Request.Context()).Where("id = ?", claims.UserID).Take(&user).Error; err != nil || user.Status != "active" || user.Role != "lawyer" {
		return "", "", time.Time{}, &npError{"OBJECT_ACCESS_DENIED", 403}
	}
	expires := claims.ExpiresAt.Time
	if s.AccessTokenExpires.Before(expires) {
		expires = s.AccessTokenExpires
	}
	if !time.Now().Before(expires) {
		return "", "", time.Time{}, &npError{"ACCESS_SESSION_REQUIRED", 401}
	}
	return actor, claims.UUID, expires, nil
}

func (n *NarrowPilotHandler) authorize(c *gin.Context, db *gorm.DB, id string, requireDraft bool) (map[string]interface{}, *npError) {
	actor, _, _, e := n.session(c, db)
	if e != nil {
		return nil, e
	}
	var row map[string]interface{}
	query := db.WithContext(c.Request.Context()).Table("case_intakes")
	if requireDraft {
		query = query.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	if query.Where("id = ?", id).Take(&row).Error != nil || strings.TrimSpace(valueString(row, "created_by")) != actor {
		return nil, &npError{"OBJECT_ACCESS_DENIED", 403}
	}
	clientID := intakeClientID(row["client_id"])
	allowed, err := services.NewAuthorizationServiceWithDB(db).CanReadClient(c.Request.Context(), services.AuthActor{UserID: uintValue(actor), Role: "lawyer"}, clientID)
	if err != nil || !allowed || clientID == 0 {
		return nil, &npError{"OBJECT_ACCESS_DENIED", 403}
	}
	if requireDraft && valueString(row, "status") != "draft" {
		return nil, &npError{"DRAFT_REQUIRED", 409}
	}
	return row, nil
}

func uintValue(s string) uint { v, _ := strconv.ParseUint(s, 10, 64); return uint(v) }
func npDecode(c *gin.Context, target interface{}) bool {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 128*1024)
	d := json.NewDecoder(c.Request.Body)
	d.DisallowUnknownFields()
	if d.Decode(target) != nil {
		npFail(c, &npError{"INVALID_REQUEST", 400})
		return false
	}
	var extra interface{}
	if d.Decode(&extra) != io.EOF {
		npFail(c, &npError{"INVALID_REQUEST", 400})
		return false
	}
	return true
}

func (n *NarrowPilotHandler) Capabilities(c *gin.Context) {
	actor, _, _, e := n.session(c, n.db)
	common.APISuccess(c, gin.H{"enabled": e == nil && n.allowed[actor]})
}

func npExpiry(x *npContext) time.Time {
	expires := x.AccessExpiry
	for _, t := range []time.Time{x.Created.Add(2 * time.Hour), x.Touched.Add(30 * time.Minute)} {
		if t.Before(expires) {
			expires = t
		}
	}
	return expires
}
func npView(x *npContext) gin.H {
	return gin.H{"id": x.ID, "context_version": x.Version, "expires_at": npExpiry(x), "candidate": x.Candidate, "saving": x.Saving}
}
func (n *NarrowPilotHandler) CreateContext(c *gin.Context) {
	var body struct {
		Intake   string `json:"intake_id"`
		Revision string `json:"expected_revision"`
	}
	if !npDecode(c, &body) {
		return
	}
	actor, access, expires, e := n.session(c, n.db)
	if e != nil {
		npFail(c, e)
		return
	}
	if !n.allowed[actor] {
		npFail(c, &npError{"ASSISTANT_DISABLED", 403})
		return
	}
	row, e := n.authorize(c, n.db, body.Intake, true)
	if e != nil {
		npFail(c, e)
		return
	}
	if intakeDraftRevisionFor(body.Intake, row["metadata"]) != body.Revision {
		npFail(c, &npError{"INTAKE_DRAFT_REVISION_CONFLICT", 409})
		return
	}
	now := time.Now()
	if expires.Sub(now) < 45*time.Second {
		npFail(c, &npError{"CONTEXT_EXPIRING", 409})
		return
	}
	response, release := n.lockResponse(c)
	defer release()
	for id, x := range n.contexts {
		if !now.Before(npExpiry(x)) && !x.Saving {
			n.dropContext(id)
		}
	}
	for id, x := range n.contexts {
		if x.Actor == actor {
			if x.Saving {
				response.err = &npError{"SAVE_IN_PROGRESS", 409}
				return
			}
			n.dropContext(id)
		}
	}
	if len(n.contexts) >= 32 {
		response.err = &npError{"ASSISTANT_BUSY", 429}
		return
	}
	x := &npContext{ID: uuid.NewString(), Version: 1, Actor: actor, Access: access, Intake: body.Intake, Client: strconv.FormatUint(uint64(intakeClientID(row["client_id"])), 10), Revision: body.Revision, Created: now, Touched: now, AccessExpiry: expires, Replays: map[string]npTurnReplay{}}
	n.contexts[x.ID] = x
	contextID := x.ID
	x.timer = time.AfterFunc(time.Until(npExpiry(x)), func() { n.expireContext(contextID) })
	response.data = npView(x)
}

// Read authorization is checked outside the mutex; a second mutex check guards replacement/expiry.
func (n *NarrowPilotHandler) contextAccess(c *gin.Context) (string, string, *npError) {
	actor, access, _, e := n.session(c, n.db)
	if e != nil {
		return "", "", e
	}
	n.mu.Lock()
	x := n.contexts[c.Param("id")]
	intake := ""
	revision := ""
	client := ""
	if x != nil && x.Actor == actor && x.Access == access && time.Now().Before(npExpiry(x)) {
		intake = x.Intake
		revision = x.Revision
		client = x.Client
	}
	n.mu.Unlock()
	if intake == "" || !n.allowed[actor] {
		return "", "", &npError{"CONTEXT_UNAVAILABLE", 409}
	}
	row, e := n.authorize(c, n.db, intake, true)
	if e != nil {
		return "", "", e
	}
	if intakeDraftRevisionFor(intake, row["metadata"]) != revision || strconv.FormatUint(uint64(intakeClientID(row["client_id"])), 10) != client {
		return "", "", &npError{"INTAKE_DRAFT_REVISION_CONFLICT", 409}
	}
	return actor, access, nil
}
func (n *NarrowPilotHandler) GetContext(c *gin.Context) {
	actor, access, e := n.contextAccess(c)
	if e != nil {
		npFail(c, e)
		return
	}
	response, release := n.lockResponse(c)
	defer release()
	x := n.contexts[c.Param("id")]
	if x == nil || x.Actor != actor || x.Access != access || !time.Now().Before(npExpiry(x)) {
		response.err = &npError{"CONTEXT_UNAVAILABLE", 409}
		return
	}
	view := npView(x)
	if x.Saving {
		view["candidate"] = nil
	}
	response.data = view
}
func (n *NarrowPilotHandler) DeleteContext(c *gin.Context) {
	var body struct {
		Version int `json:"expected_context_version"`
	}
	if !npDecode(c, &body) {
		return
	}
	actor, access, e := n.contextAccess(c)
	if e != nil {
		npFail(c, e)
		return
	}
	response, release := n.lockResponse(c)
	defer release()
	x := n.contexts[c.Param("id")]
	if x == nil || x.Actor != actor || x.Access != access || x.Version != body.Version || x.Saving {
		response.err = &npError{"CONTEXT_STALE", 409}
		return
	}
	n.dropContext(x.ID)
	response.data = gin.H{"ended": true}
}
func (n *NarrowPilotHandler) Turn(c *gin.Context) {
	var body struct {
		Request string `json:"request_id"`
		Version int    `json:"expected_context_version"`
		Text    string `json:"text"`
	}
	if !npDecode(c, &body) {
		return
	}
	if _, err := uuid.Parse(body.Request); err != nil || strings.TrimSpace(body.Text) == "" || utf8.RuneCountInString(body.Text) > 6000 {
		npFail(c, &npError{"INVALID_TURN", 400})
		return
	}
	actor, access, e := n.contextAccess(c)
	if e != nil {
		npFail(c, e)
		return
	}
	response, release := n.lockResponse(c)
	defer release()
	x := n.contexts[c.Param("id")]
	if x == nil || x.Actor != actor || x.Access != access || !time.Now().Before(npExpiry(x)) || x.Saving {
		response.err = &npError{"CONTEXT_UNAVAILABLE", 409}
		return
	}
	if prior, ok := x.Replays[body.Request]; ok {
		if prior.Body != body.Text || prior.Version != body.Version {
			response.err = &npError{"TURN_KEY_REUSED", 409}
			return
		}
		response.data = npView(x)
		return
	}
	if x.Version != body.Version {
		response.err = &npError{"CONTEXT_STALE", 409}
		return
	}
	if len(x.Sources) >= 8 || npExpiry(x).Sub(time.Now()) < 45*time.Second {
		response.err = &npError{"CONTEXT_BUDGET_EXHAUSTED", 409}
		return
	}
	total := utf8.RuneCountInString(body.Text)
	for _, s := range x.Sources {
		total += utf8.RuneCountInString(s.Text)
	}
	if total > 20000 {
		response.err = &npError{"CONTEXT_BUDGET_EXHAUSTED", 409}
		return
	}
	x.Sources = append(x.Sources, npSource{ID: uuid.NewString(), Kind: "user_input", Text: body.Text, Turn: len(x.Sources) + 1, EnteredBy: actor, AcceptedAt: time.Now().UTC()})
	x.Replays[body.Request] = npTurnReplay{Body: body.Text, Version: body.Version}
	x.Version++
	x.Touched = time.Now()
	texts := make([]string, 0, len(x.Sources))
	for _, s := range x.Sources {
		texts = append(texts, s.Text)
	}
	// Stub copies accepted input verbatim; it makes no legal or factual correctness claim.
	x.Candidate = &npCandidate{ID: uuid.NewString(), Description: "据本次录入（待核实）：\n" + strings.Join(texts, "\n"), Sources: append([]npSource(nil), x.Sources...)}
	response.data = npView(x)
}

func (n *NarrowPilotHandler) prepareSource(c *gin.Context, row map[string]interface{}, selection *npSourceSelection, description string) (*npSourcePack, *npError) {
	// The caller has already checked the durable session in its write transaction.
	claims, err := middleware.ValidateToken(strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer "))
	if err != nil || claims.Type != "access" {
		return nil, &npError{"ACCESS_SESSION_REQUIRED", 401}
	}
	actor, access := strconv.FormatUint(uint64(claims.UserID), 10), claims.UUID
	if !n.allowed[actor] || selection == nil || len(selection.Fields) != 1 {
		return nil, &npError{"SOURCE_SELECTION_INVALID", 409}
	}
	field, ok := selection.Fields["description"]
	if !ok || !field.Reviewed || field.RequestedMode != "candidate_sources" {
		return nil, &npError{"SOURCE_SELECTION_INVALID", 409}
	}
	n.mu.Lock()
	defer n.mu.Unlock()
	x := n.contexts[selection.ContextID]
	if x == nil || x.Actor != actor || x.Access != access || x.Intake != valueString(row, "id") || x.Client != strconv.FormatUint(uint64(intakeClientID(row["client_id"])), 10) || x.Revision != intakeDraftRevisionFor(x.Intake, row["metadata"]) || x.Version != selection.ExpectedContextVersion || !time.Now().Before(npExpiry(x)) || x.Saving || x.Candidate == nil || x.Candidate.ID != field.CandidateID || x.Candidate.Description != description {
		return nil, &npError{"SOURCE_SELECTION_STALE", 409}
	}
	x.Saving = true
	return &npSourcePack{ContextID: x.ID, ContextVersion: x.Version, CandidateID: field.CandidateID, OriginalRevision: x.Revision, Description: description, Sources: append([]npSource(nil), x.Candidate.Sources...), ReviewedBy: actor, ReviewedAt: time.Now().UTC()}, nil
}
func (n *NarrowPilotHandler) finishSource(pack *npSourcePack, committed bool) {
	if pack == nil {
		return
	}
	n.mu.Lock()
	defer n.mu.Unlock()
	if x := n.contexts[pack.ContextID]; x != nil {
		if committed || !time.Now().Before(npExpiry(x)) {
			n.dropContext(x.ID)
		} else {
			x.Saving = false
		}
	}
}

func (n *NarrowPilotHandler) expireContext(id string) {
	n.mu.Lock()
	defer n.mu.Unlock()
	x := n.contexts[id]
	if x == nil {
		return
	}
	remaining := time.Until(npExpiry(x))
	if remaining <= 0 && !x.Saving {
		n.dropContext(id)
		return
	}
	if remaining <= 0 {
		remaining = time.Minute
	}
	x.timer = time.AfterFunc(remaining, func() { n.expireContext(id) })
}

// Build a response under the short memory lock, then release it before socket I/O.
type npResponse struct {
	data gin.H
	err  *npError
}

func (n *NarrowPilotHandler) lockResponse(c *gin.Context) (*npResponse, func()) {
	n.mu.Lock()
	r := &npResponse{}
	return r, func() {
		n.mu.Unlock()
		if r.err != nil {
			npFail(c, r.err)
		} else {
			common.APISuccess(c, r.data)
		}
	}
}

// Called with mu held, so ended conversations do not retain bodies or timer closures.
func (n *NarrowPilotHandler) dropContext(id string) {
	if x := n.contexts[id]; x != nil && x.timer != nil {
		x.timer.Stop()
	}
	delete(n.contexts, id)
}
