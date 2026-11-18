package services

// Server-verified intake continuation (spec: refine-trial-usability-mvp,
// "Server-verified continuation"). The continuation API reads the same
// server-owned evidence the structured resubmit gate reads: committed intake
// metadata, the formal check record bound to this intake, the independent
// review conclusion, and the active submission snapshot pointer. Browser or
// localStorage validity flags never enter this decision, and protected
// identity digests stay server-side: ordinary reads only receive masked
// display values and opaque workflow tokens.

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"

	"gorm.io/gorm"

	"law-oa-go/internal/models"
	"law-oa-go/internal/security"
)

// Continuation reason codes (fail closed). Every code names the exact
// server-side verification that failed; an empty list means the facts, the
// check binding, and the independent review chain are all proven current.
const (
	ContinuationReasonFactsUnconfirmed       = "FACTS_UNCONFIRMED"
	ContinuationReasonFactsInvalidated       = "FACTS_INVALIDATED"
	ContinuationReasonFactsVersionMismatch   = "FACTS_VERSION_MISMATCH"
	ContinuationReasonCheckMissing           = "CONFLICT_CHECK_MISSING"
	ContinuationReasonCheckIncomplete        = "CONFLICT_CHECK_INCOMPLETE"
	ContinuationReasonCoverageIncomplete     = "COVERAGE_INCOMPLETE"
	ContinuationReasonReviewMissing          = "REVIEW_MISSING"
	ContinuationReasonReviewInsufficient     = "REVIEW_INSUFFICIENT"
	ContinuationReasonReviewConflict         = "REVIEW_CONFLICT"
	ContinuationReasonReviewWaiverPending    = "REVIEW_WAIVER_PENDING"
	ContinuationReasonReviewUnresolved       = "REVIEW_UNRESOLVED"
	ContinuationReasonSubjectUnreadable      = "SUBJECT_UNKNOWN"
	ContinuationReasonSubjectIdentityChanged = "SUBJECT_IDENTITY_CHANGED"
	ContinuationReasonSnapshotUnresolved     = "SNAPSHOT_UNRESOLVED"
	ContinuationReasonApprovalInProgress     = "APPROVAL_IN_PROGRESS"
	ContinuationReasonApprovalClosed         = "APPROVAL_CLOSED"
	ContinuationReasonApprovalStateUnknown   = "APPROVAL_STATE_UNKNOWN"
)

// ContinuationSubject is the masked projection of one protected subject of
// the intake. Plaintext identity numbers and lookup digests never enter the
// response: the digest is only used server-side to prove the subject still
// matches the recorded evidence; the browser receives the masked hint only.
type ContinuationSubject struct {
	Kind           string `json:"kind"`
	PartyID        uint   `json:"party_id,omitempty"`
	Name           string `json:"name"`
	Role           string `json:"role,omitempty"`
	EntityType     string `json:"entity_type,omitempty"`
	IdentityType   string `json:"identity_type,omitempty"`
	IdentityMasked string `json:"identity_masked"`
	// Aliases are the same-naming facts a re-run conflict check would send
	// (intake metadata client_aliases / party aliases). They are identity
	// FACTS for the frozen-tuple comparison but display-only names, never
	// protected identifier material; they stay out of the JSON projection.
	Aliases []string `json:"-"`
}

// ContinuationBinding carries opaque workflow tokens of the verified check
// chain. It never carries identity material.
type ContinuationBinding struct {
	CheckID        string `json:"check_id,omitempty"`
	CheckStatus    string `json:"check_status,omitempty"`
	CoverageStatus string `json:"coverage_status,omitempty"`
	ReviewID       uint64 `json:"review_id,omitempty"`
	ReviewDecision string `json:"review_decision,omitempty"`
	ReviewerID     uint64 `json:"reviewer_id,omitempty"`
}

// ContinuationApprovalContext anchors the workflow side: which approval
// still owns this intake and which frozen submission round it reads. Only
// populated when approval metadata names this intake.
type ContinuationApprovalContext struct {
	ApprovalID string `json:"approval_id"`
	Status     string `json:"status,omitempty"`
	SnapshotID string `json:"snapshot_id,omitempty"`
	Round      int    `json:"round,omitempty"`
}

// ContinuationFacts carries opaque server-side state tokens only.
type ContinuationFacts struct {
	FactsVersion  string `json:"facts_version,omitempty"`
	ConfirmedAt   string `json:"confirmed_at,omitempty"`
	InvalidatedAt string `json:"invalidated_at,omitempty"`
}

// ContinuationDecision separates the independent reuse questions so a
// missing independent review never forces identity re-entry or a new check:
//
//   - CanReuseIdentity: protected subjects and facts are verified current
//     server-side; the lawyer never re-enters identities.
//   - CanReuseCheck: the recorded formal check still owns these facts
//     (status COMPLETED, coverage COMPLETE, record ownership, version match).
//   - CanSubmit: the full submission gate additionally holds: an acceptable
//     independent review on that same check and a resolvable active snapshot.
//
// RecheckRequired is true only when the check chain no longer owns the
// current facts (facts changed/unconfirmed, check missing/incomplete, or
// coverage incomplete) and a new confirmation plus a new formal check is
// required. Waiting for the original independent review keeps RecheckRequired
// false so the lawyer resumes instead of re-running detection.
type ContinuationDecision struct {
	IntakeID         string                       `json:"intake_id"`
	Revision         string                       `json:"revision"`
	Status           string                       `json:"status"`
	CanReuseIdentity bool                         `json:"can_reuse_identity"`
	CanReuseCheck    bool                         `json:"can_reuse_check"`
	CanSubmit        bool                         `json:"can_submit"`
	RecheckRequired  bool                         `json:"recheck_required"`
	Reasons          []string                     `json:"reasons"`
	Facts            *ContinuationFacts           `json:"facts,omitempty"`
	Binding          *ContinuationBinding         `json:"conflict_binding,omitempty"`
	Subjects         []ContinuationSubject        `json:"subjects,omitempty"`
	Approval         *ContinuationApprovalContext `json:"approval,omitempty"`
}

func (d *ContinuationDecision) appendReason(reason string) {
	for _, existing := range d.Reasons {
		if existing == reason {
			return
		}
	}
	d.Reasons = append(d.Reasons, reason)
}

func (d *ContinuationDecision) hasReason(reason string) bool {
	for _, existing := range d.Reasons {
		if existing == reason {
			return true
		}
	}
	return false
}

// continuationSubjectEvidence keeps the server-side digest alongside the
// masked projection so the frozen-anchor comparison never re-reads identity
// material. Evidence never enters the JSON response.
type continuationSubjectEvidence struct {
	Subject        ContinuationSubject
	IdentityDigest string
}

// continuationClientMirror carries the client-master facts a guarded
// case_intake_parties mirror row (party_role=client, written by
// prepareCaseIntakeClientPartyRow) must prove before continuation may
// normalize that row into the client subject instead of treating it as an
// added party. A real intake check freezes the client ONLY from the master,
// so an unverified mirror row can never silently disappear from the
// current-evidence set.
type continuationClientMirror struct {
	name          string
	identityType  string
	digest        string
	columnAliases []string
}

// continuationIntakeEntityTypeMatchesMaster applies the same guarded-save
// compatibility rule as clientEntityTypeMatchesMaster: the intake mirror
// row stores the batch entity-type bucket (INDIVIDUAL/ORGANIZATION/
// LEGAL_PERSON), while the master row stores its identity type. A mirror
// row is the same client only when that bucket agrees with the master
// identity class; an empty bucket is the save path default (LEGAL_PERSON).
func continuationIntakeEntityTypeMatchesMaster(entityType, identityType string) bool {
	// Same alias set as the guarded save path (normalizeIntakeEntityType).
	// Unlike that save helper this verifier must also reject unknown
	// non-empty spellings: prepareCaseIntakePartyRow would never persist
	// them for a guarded client row, so anything else is evidence tampering.
	normalized := ""
	switch strings.ToUpper(strings.TrimSpace(entityType)) {
	case "PERSON", "INDIVIDUAL", "个人", "自然人":
		normalized = "INDIVIDUAL"
	case "COMPANY", "LEGAL_PERSON", "企业", "公司":
		normalized = "LEGAL_PERSON"
	case "ORGANIZATION", "组织", "其他组织":
		normalized = "ORGANIZATION"
	case "":
		// Save-contract compatibility only: rows stored by save versions
		// that defaulted an omitted value to the legal-person bucket. The
		// master identity type below still gates the final answer.
		normalized = "LEGAL_PERSON"
	default:
		return false
	}
	switch strings.ToUpper(strings.TrimSpace(identityType)) {
	case "SOCIAL_CREDIT_CODE", "BUSINESS_LICENSE", "ORGANIZATION_CODE":
		return normalized == "LEGAL_PERSON"
	case "ID_CARD", "PASSPORT":
		return normalized == "INDIVIDUAL"
	default:
		// Empty mirrors the Client.EffectiveIdentityType legacy fallback
		// used by the guarded save path; unknown values stay rejected.
		if strings.TrimSpace(identityType) == "" {
			return normalized == "LEGAL_PERSON"
		}
		return false
	}
}

// IntakeContinuationService answers "may this user safely resume the same
// intake" from server-owned evidence only.
type IntakeContinuationService struct {
	db *gorm.DB
}

func NewIntakeContinuationService(db *gorm.DB) *IntakeContinuationService {
	return &IntakeContinuationService{db: db}
}

// EvaluateIntakeContinuation derives the continuation decision inside one
// read transaction. The binding gate reuses the exact verification the
// structured resubmit uses (verifyIntakeBindingGateTx), so the read-only
// view can never accept a binding the resubmit path would reject. Unknown
// infrastructure failures propagate as errors (fail closed) instead of
// silently clearing the reuse flags.
func (s *IntakeContinuationService) EvaluateIntakeContinuation(intakeID string) (*ContinuationDecision, error) {
	var decision *ContinuationDecision
	err := s.db.Transaction(func(tx *gorm.DB) error {
		var evalErr error
		decision, evalErr = evaluateIntakeContinuationTx(tx, intakeID)
		return evalErr
	})
	if err != nil {
		return nil, err
	}
	return decision, nil
}

func evaluateIntakeContinuationTx(tx *gorm.DB, intakeID string) (*ContinuationDecision, error) {
	binding, metadata, intakeStatus, err := readIntakeConflictBindingTx(tx, intakeID)
	if err != nil {
		return nil, err
	}
	revision := metadataString(metadata, "draft_revision")
	if revision == "" {
		revision = "legacy-" + intakeID
	}
	decision := &ContinuationDecision{
		IntakeID: intakeID,
		Revision: revision,
		Status:   intakeStatus,
		Reasons:  []string{},
	}
	factsVersion := metadataString(metadata, "lawyer_facts_version")
	invalidatedAt := metadataString(metadata, "conflict_invalidated_at")
	committedFactsVersion := metadataString(metadata, "conflict_check_facts_version")
	factsUnconfirmed := intakeStatus != "conflict_ready"
	factsInvalidated := invalidatedAt != ""
	factsVersionMismatch := committedFactsVersion == "" || factsVersion == "" || committedFactsVersion != factsVersion
	factsBroken := factsUnconfirmed || factsInvalidated || factsVersionMismatch
	if factsVersion != "" || factsUnconfirmed || factsInvalidated {
		decision.Facts = &ContinuationFacts{
			FactsVersion:  factsVersion,
			ConfirmedAt:   metadataString(metadata, "lawyer_facts_confirmed_at"),
			InvalidatedAt: invalidatedAt,
		}
	}
	if factsUnconfirmed {
		decision.appendReason(ContinuationReasonFactsUnconfirmed)
	}
	if factsInvalidated {
		decision.appendReason(ContinuationReasonFactsInvalidated)
	}
	if factsVersionMismatch {
		decision.appendReason(ContinuationReasonFactsVersionMismatch)
	}
	// Stage A: run the resubmit gate only when the committed facts state is
	// intact; the gate re-verifies status, invalidation, version match, and
	// ownership before any reason code may be cleared.
	var gateBinding *intakeConflictBinding
	if !factsBroken {
		gateBinding, err = verifyIntakeBindingGateTx(tx, intakeID, binding)
		if err != nil {
			reasons := mapContinuationGateError(err)
			if len(reasons) == 0 {
				// Unknown database/table failure: never clear the flags on
				// an unverifiable state. Propagate so the caller sees 5xx.
				return nil, err
			}
			for _, reason := range reasons {
				decision.appendReason(reason)
			}
		}
	}
	if gateBinding != nil {
		decision.Binding = &ContinuationBinding{
			CheckID:        gateBinding.CheckID,
			CheckStatus:    gateBinding.CheckStatus,
			CoverageStatus: gateBinding.CoverageStatus,
			ReviewID:       gateBinding.ReviewID,
			ReviewDecision: gateBinding.ReviewDecision,
			ReviewerID:     gateBinding.ReviewerID,
		}
	} else if binding.CheckID != "" {
		decision.Binding = &ContinuationBinding{CheckID: binding.CheckID}
	}
	// Subjects are projected masked with their server-side digests retained
	// for the frozen-anchor comparison. An unreadable, missing, or tampered
	// protected subject fails closed and hides every subject entry.
	evidence, subjectErr := continuationSubjects(tx, intakeID)
	if subjectErr != nil {
		decision.appendReason(ContinuationReasonSubjectUnreadable)
	} else {
		subjects := make([]ContinuationSubject, 0, len(evidence))
		for _, item := range evidence {
			subjects = append(subjects, item.Subject)
		}
		decision.Subjects = subjects
	}
	// The old check may only back the current facts when the current
	// protected subjects still match the identity anchors frozen inside the
	// check's audit search parameters. A client whose identity changed in
	// client management cannot silently reuse the old result.
	identityAnchored := false
	if subjectErr != nil {
		// SUBJECT_UNKNOWN already appended; identity cannot be proven.
	} else if gateBinding == nil || strings.TrimSpace(gateBinding.FrozenParameters) == "" {
		// No owned check record exists to compare against. The registered
		// subjects are readable and digest-consistent, so identity re-entry
		// is not forced; the missing/incomplete check reason already forces
		// a fresh check on the same identities.
		identityAnchored = true
	} else {
		if anchorErr := verifyContinuationFrozenIdentity(gateBinding.FrozenParameters, evidence); anchorErr != nil {
			if errors.Is(anchorErr, errContinuationIdentityChanged) {
				decision.appendReason(ContinuationReasonSubjectIdentityChanged)
			} else {
				decision.appendReason(ContinuationReasonSubjectUnreadable)
			}
		} else {
			identityAnchored = true
		}
	}
	decision.CanReuseIdentity = !factsBroken && subjectErr == nil && identityAnchored
	decision.CanReuseCheck = decision.CanReuseIdentity &&
		!decision.hasReason(ContinuationReasonCheckMissing) &&
		!decision.hasReason(ContinuationReasonCheckIncomplete) &&
		!decision.hasReason(ContinuationReasonCoverageIncomplete)
	// Recheck means the detection chain itself no longer binds these facts;
	// a pending or failing review never sets it.
	decision.RecheckRequired = !decision.CanReuseCheck
	decision.CanSubmit = decision.CanReuseCheck &&
		!decision.hasReason(ContinuationReasonReviewMissing) &&
		!decision.hasReason(ContinuationReasonReviewInsufficient) &&
		!decision.hasReason(ContinuationReasonReviewConflict) &&
		!decision.hasReason(ContinuationReasonReviewWaiverPending) &&
		!decision.hasReason(ContinuationReasonReviewUnresolved)
	if err := attachContinuationApprovalContext(tx, intakeID, decision); err != nil {
		return nil, err
	}
	// An intake already owned by an approval can only resubmit that same
	// request in needs_revision. Phase blockers do not invalidate identity
	// or conflict evidence and therefore never force a new check.
	if decision.Approval != nil {
		switch decision.Approval.Status {
		case models.ApprovalStatusNeedsRevision:
			// Retain the existing evidence and active-snapshot gates.
		case models.ApprovalStatusSubmitted, models.ApprovalStatusUnderReview, models.ApprovalStatusResubmitted:
			decision.appendReason(ContinuationReasonApprovalInProgress)
			decision.CanSubmit = false
		case models.ApprovalStatusApproved, models.ApprovalStatusRejected, models.ApprovalStatusCancelled, models.ApprovalStatusExpired:
			decision.appendReason(ContinuationReasonApprovalClosed)
			decision.CanSubmit = false
		default:
			decision.appendReason(ContinuationReasonApprovalStateUnknown)
			decision.CanSubmit = false
		}
	}
	// Submitting requires a resolvable active intake_submission round only
	// when an approval already owns this intake. A first submission has no
	// approval yet: that is the normal state, not a snapshot failure.
	if decision.CanSubmit && decision.Approval != nil && strings.TrimSpace(decision.Approval.SnapshotID) == "" {
		decision.appendReason(ContinuationReasonSnapshotUnresolved)
		decision.CanSubmit = false
	}
	return decision, nil
}

// verifyIntakeBindingGateTx verifies the check record + independent review +
// coverage chain server-side. It mirrors the structured resubmit gate
// (verifyIntakeConflictBindingTx) so the read-only continuation view applies
// the same evidence rules as the write path. The gate NEVER relaxes the
// resubmit contract, and continuation clears review reasons separately.
func verifyIntakeBindingGateTx(tx *gorm.DB, intakeID string, binding *intakeConflictBinding) (*intakeConflictBinding, error) {
	if binding == nil || strings.TrimSpace(binding.CheckID) == "" {
		return nil, &ResubmitContractError{
			Code:    models.ErrCodeApprovalSnapshotUnresolved,
			Message: "接案缺少可验证的冲突检测绑定",
		}
	}
	var record struct {
		CheckStatus      string
		CheckResult      string
		SearchParameters string
	}
	if err := tx.Table("conflict_check_records").Select("check_status, check_result, search_parameters").
		Where("check_id = ?", binding.CheckID).Take(&record).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, &ResubmitContractError{
				Code:    models.ErrCodeApprovalSnapshotUnresolved,
				Message: "绑定的冲突检测记录不存在",
			}
		}
		return nil, fmt.Errorf("读取冲突检测记录失败：%w", err)
	}
	// The immutable audit record must name this intake: a record built for
	// another intake cannot back these facts even if a pointer names it.
	recordedIntakeID := conflictRecordIntakeID(parseSearchParametersMap(record.SearchParameters))
	if recordedIntakeID == "" || recordedIntakeID != intakeID {
		return nil, &ResubmitContractError{
			Code:    models.ErrCodeApprovalSnapshotUnresolved,
			Message: "冲突检测记录未绑定当前接案",
		}
	}
	binding.CheckStatus = strings.TrimSpace(record.CheckStatus)
	binding.FrozenParameters = record.SearchParameters
	if !strings.EqualFold(binding.CheckStatus, "COMPLETED") {
		return binding, &ResubmitContractError{
			Code:    models.ErrCodeApprovalSnapshotUnresolved,
			Message: "绑定的冲突检测尚未完成",
		}
	}
	binding.CoverageStatus = conflictCoverageStatusFromRecord(record.CheckResult, record.SearchParameters)
	if !strings.EqualFold(binding.CoverageStatus, "COMPLETE") {
		return binding, &ResubmitContractError{
			Code:    models.ErrCodeApprovalSnapshotUnresolved,
			Message: "冲突检测覆盖范围未确认完整",
		}
	}
	var review models.ConflictReview
	if err := tx.Where("check_id = ?", binding.CheckID).Order("created_at DESC").Take(&review).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return binding, &ResubmitContractError{
				Code:    models.ErrCodeApprovalSnapshotUnresolved,
				Message: "冲突检测尚未完成独立人工复核",
			}
		}
		return nil, fmt.Errorf("读取冲突复核结论失败：%w", err)
	}
	binding.ReviewDecision = strings.ToLower(strings.TrimSpace(review.Decision))
	binding.ReviewID = uint64(review.ID)
	binding.ReviewerID = uint64(review.ReviewerID)
	switch binding.ReviewDecision {
	case "no_conflict", "false_positive":
		return binding, nil
	case "insufficient_information":
		return binding, &ResubmitContractError{Code: models.ErrCodeApprovalSnapshotUnresolved, Message: "冲突复核结论为信息不足"}
	case "confirmed_conflict":
		return binding, &ResubmitContractError{Code: models.ErrCodeApprovalSnapshotUnresolved, Message: "独立复核已确认存在利益冲突"}
	case "waiver_requested":
		return binding, &ResubmitContractError{Code: models.ErrCodeApprovalSnapshotUnresolved, Message: "冲突复核已申请豁免"}
	default:
		return binding, &ResubmitContractError{Code: models.ErrCodeApprovalSnapshotUnresolved, Message: "冲突复核结论尚未达到可提交状态"}
	}
}

// errContinuationIdentityChanged marks a proven mismatch between the current
// protected subject identity and the identity anchors frozen with the old
// check. It is distinct from an infrastructure failure: the chain is intact
// but the facts no longer describe the same subject.
var errContinuationIdentityChanged = errors.New("continuation identity anchor mismatch")

// continuationFrozenSubject is one complete subject fact tuple frozen inside
// the old check's audit search parameters (written by auditSafeConflictRequest
// from the real intake check request). Anchors carry the exact
// "hmac-sha256:<hex>" strings produced by digestIdentifiers; keys are
// canonical identity types so a stored type spelling cannot bypass equality.
type continuationFrozenSubject struct {
	Kind       string
	Name       string
	Role       string
	EntityType string
	Aliases    []string
	Anchors    map[string]string
}

// parseFrozenAnchorMap canonicalizes one frozen identifier map. Any missing
// or malformed digest (including the empty placeholder written when the
// subject key was unavailable) makes the whole set unprovable: the caller
// must fail closed instead of guessing which anchors still count.
func parseFrozenAnchorMap(values interface{}) (map[string]string, bool) {
	if values == nil {
		return nil, false
	}
	raw, err := json.Marshal(values)
	if err != nil {
		return nil, false
	}
	var pairs map[string]interface{}
	if json.Unmarshal(raw, &pairs) != nil {
		return nil, false
	}
	anchors := map[string]string{}
	for key, value := range pairs {
		text, _ := value.(string)
		text = strings.TrimSpace(text)
		canonicalKey := canonicalIdentityType(key)
		if canonicalKey == "" || !strings.HasPrefix(text, "hmac-sha256:") || len(text) <= len("hmac-sha256:") {
			return nil, false
		}
		anchors[canonicalKey] = text
	}
	if len(anchors) == 0 {
		return nil, false
	}
	return anchors, true
}

// frozenStringList canonicalizes one frozen alias list. Missing aliases are
// NOT an integrity failure (very short audits may carry none), but any
// malformed entry fails closed: aliases belong to the frozen identity facts.
func frozenStringList(values interface{}) ([]string, bool) {
	if values == nil {
		return []string{}, true
	}
	raw, err := json.Marshal(values)
	if err != nil {
		return nil, false
	}
	var list []interface{}
	if json.Unmarshal(raw, &list) != nil {
		return nil, false
	}
	out := make([]string, 0, len(list))
	for _, item := range list {
		text, ok := item.(string)
		if !ok {
			return nil, false
		}
		text = strings.TrimSpace(text)
		if text != "" && text != "<nil>" {
			out = append(out, text)
		}
	}
	sort.Strings(out)
	// The tuple compare treats aliases as one sorted set; the raw frozen
	// payload may repeat the same alias across storage copies, so collapse
	// duplicates here instead of letting the set compare count them.
	deduped := make([]string, 0, len(out))
	for index, value := range out {
		if index > 0 && out[index-1] == value {
			continue
		}
		deduped = append(deduped, value)
	}
	return deduped, true
}

// frozenContinuationSubjects extracts the complete frozen subject fact
// tuples: name, role, entity type, aliases, and canonical identity anchors.
// A real intake check always freezes one client plus at least one protected
// party; anything less proves nothing and must fail closed.
func frozenContinuationSubjects(frozen models.JSON) ([]continuationFrozenSubject, bool) {
	subjects := []continuationFrozenSubject{}
	var rawClient interface{}
	var haveClient bool
	for _, key := range []string{"clientIdentifiers", "client_identifiers"} {
		if value, ok := frozen[key]; ok {
			rawClient, haveClient = value, true
			break
		}
	}
	if !haveClient {
		return nil, false
	}
	anchors, ok := parseFrozenAnchorMap(rawClient)
	if !ok {
		return nil, false
	}
	clientName := strings.TrimSpace(firstContinuationString(frozen, "clientName", "client_name"))
	if clientName == "" {
		return nil, false
	}
	clientAliases, aliasesOK := frozenStringList(frozen["clientAliases"])
	if !aliasesOK {
		if _, present := frozen["clientAliases"]; present {
			return nil, false
		}
		clientAliases = []string{}
	}
	clientType := normalizeConflictSubjectEntityType(firstContinuationString(frozen, "clientType", "client_type"))
	subjects = append(subjects, continuationFrozenSubject{
		Kind: "client", Name: clientName, Role: "CLIENT", EntityType: clientType,
		Aliases: clientAliases, Anchors: anchors,
	})
	rawParties, ok := frozen["parties"]
	if !ok {
		return nil, false
	}
	raw, err := json.Marshal(rawParties)
	if err != nil {
		return nil, false
	}
	var partyList []map[string]interface{}
	if json.Unmarshal(raw, &partyList) != nil || len(partyList) == 0 {
		return nil, false
	}
	for _, party := range partyList {
		name := strings.TrimSpace(firstContinuationString(party, "name"))
		partyAnchors, anchorsOK := parseFrozenAnchorMap(party["identifiers"])
		roleValue, hasRole := party["role"]
		if name == "" || !anchorsOK || !hasRole {
			return nil, false
		}
		partyAliases, partyAliasesOK := frozenStringList(party["aliases"])
		if !partyAliasesOK {
			if _, present := party["aliases"]; present {
				return nil, false
			}
			partyAliases = []string{}
		}
		entityType := normalizeConflictSubjectEntityType(firstContinuationString(party, "entityType", "entity_type"))
		subjects = append(subjects, continuationFrozenSubject{
			Kind: "party", Name: name,
			Role:       normalizeConflictSubjectRole(fmt.Sprint(roleValue)),
			EntityType: entityType,
			Aliases:    partyAliases,
			Anchors:    partyAnchors,
		})
	}
	return subjects, true
}

func firstContinuationString(values map[string]interface{}, keys ...string) string {
	for _, key := range keys {
		if value, ok := values[key].(string); ok {
			if trimmed := strings.TrimSpace(value); trimmed != "" && trimmed != "<nil>" {
				return trimmed
			}
		}
	}
	return ""
}

// continuationStringList mirrors the real intake check's alias extraction
// pair (firstNonEmptyValue + stringListValue): a non-empty string candidate
// is selected as-is and split with the shared separator contract, a
// non-empty array candidate wins with its entries trimmed, and an EMPTY
// array falls through to the next candidate exactly like firstNonEmptyValue.
// A selected non-array scalar contributes no alias fact, matching the
// handler's stringListValue, which only understands lists and strings.
func continuationStringList(values ...interface{}) []string {
	for _, value := range values {
		switch typed := value.(type) {
		case string:
			if strings.TrimSpace(typed) == "" {
				continue
			}
			// Selected exactly like firstNonEmptyValue selects a non-blank
			// string; the split result is the alias fact even when every
			// item is a separator, so no later candidate is retried.
			return splitContinuationStoredAliases(typed)
		case []interface{}:
			if len(typed) == 0 {
				continue
			}
			out := make([]string, 0, len(typed))
			for _, item := range typed {
				if trimmed := strings.TrimSpace(fmt.Sprint(item)); trimmed != "" && trimmed != "<nil>" {
					out = append(out, trimmed)
				}
			}
			return out
		case []string:
			if len(typed) == 0 {
				continue
			}
			out := make([]string, 0, len(typed))
			for _, item := range typed {
				if trimmed := strings.TrimSpace(item); trimmed != "" {
					out = append(out, trimmed)
				}
			}
			return out
		default:
			if typed == nil {
				continue
			}
			rendered := strings.TrimSpace(fmt.Sprint(typed))
			if rendered == "" || rendered == "<nil>" {
				continue
			}
			return []string{}
		}
	}
	return []string{}
}

func continuationStringOrEmpty(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

// continuationMasterIdentityType resolves the protected identity type the
// guarded save copied into the mirror row: an explicit master value wins,
// and only the legacy Client.EffectiveIdentityType fallback (company master
// without identity_type) may fill an empty one. Unknown values stay as-is so
// the mirror comparison rejects them instead of guessing.
func continuationMasterIdentityType(identityType, clientType string) string {
	if strings.TrimSpace(identityType) != "" {
		return identityType
	}
	switch strings.EqualFold(strings.TrimSpace(clientType), "企业") ||
		strings.EqualFold(strings.TrimSpace(clientType), "公司") ||
		strings.EqualFold(strings.TrimSpace(clientType), "COMPANY") {
	case true:
		return "SOCIAL_CREDIT_CODE"
	default:
		return "ID_CARD"
	}
}

// parseContinuationStoredAliasList decodes the delimited alias column the
// guarded intake save path writes. The separator set matches the real
// intake check's splitConflictAliases (comma, half/full-width semicolon,
// newline), so one stored string freezes and compares as the same alias
// list on both sides of the boundary. Missing columns or values are empty
// facts; the split keeps raw stored text so a changed alias cannot silently
// disappear from the mirror comparison.
func parseContinuationStoredAliasList(value *string) []string {
	if value == nil || strings.TrimSpace(*value) == "" {
		return []string{}
	}
	parts := strings.FieldsFunc(*value, func(r rune) bool {
		return r == ',' || r == ';' || r == '；' || r == '\n'
	})
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}

// continuationClientColumnAliases reads the plain clients.aliases storage
// column. Client management stores one delimited string here, so the
// splitter must accept the exact separator set the real intake check uses;
// a master stored with semicolons or newlines would otherwise compare as a
// single glued alias and reject the client's own unchanged identity.
func continuationClientColumnAliases(value *string) []string {
	if value == nil {
		return []string{}
	}
	return splitContinuationStoredAliases(*value)
}

// splitContinuationStoredAliases is the continuation-side twin of the real
// intake check's splitConflictAliases: identical separator set, trimming,
// and blank-item removal. Every column or list alias crosses the frozen
// boundary through one of these two aligned splitters.
func splitContinuationStoredAliases(value string) []string {
	items := strings.FieldsFunc(value, func(r rune) bool {
		return r == ',' || r == ';' || r == '；' || r == '\n'
	})
	out := make([]string, 0, len(items))
	for _, item := range items {
		if trimmed := strings.TrimSpace(item); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}

// normalizeContinuationAliasSet collapses the duplicate storage copies of
// one alias fact (master column, intake metadata, mirror column, mirror
// metadata) into the trimmed, de-duplicated, sorted set the frozen audit
// compares. Only blank padding, repetition, and ordering collapse here; a
// genuinely new or removed alias value still changes the set and fails
// closed.
func normalizeContinuationAliasSet(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	out := make([]string, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		if _, exists := seen[trimmed]; exists {
			continue
		}
		seen[trimmed] = struct{}{}
		out = append(out, trimmed)
	}
	sort.Strings(out)
	return out
}

// continuationEvidenceTuple projects one current subject into the same
// canonical fact tuple shape as the frozen audit payload. Identity digests
// are re-prefixed exactly as digestIdentifiers wrote them, so a stored
// digest can only match the anchor of the same identity type.
func continuationEvidenceTuple(item continuationSubjectEvidence) (continuationFrozenSubject, bool) {
	subject := item.Subject
	name := strings.TrimSpace(subject.Name)
	if name == "" {
		return continuationFrozenSubject{}, false
	}
	anchors := map[string]string{}
	if digest := strings.TrimSpace(item.IdentityDigest); digest != "" {
		identityType := canonicalIdentityType(subject.IdentityType)
		if identityType == "" {
			return continuationFrozenSubject{}, false
		}
		anchors[identityType] = "hmac-sha256:" + digest
	}
	aliases := make([]string, 0, len(subject.Aliases))
	for _, alias := range subject.Aliases {
		if trimmed := strings.TrimSpace(alias); trimmed != "" {
			aliases = append(aliases, trimmed)
		}
	}
	sort.Strings(aliases)
	tuple := continuationFrozenSubject{
		Kind:       subject.Kind,
		Name:       name,
		Role:       "CLIENT",
		EntityType: normalizeConflictSubjectEntityType(subject.EntityType),
		Aliases:    aliases,
		Anchors:    anchors,
	}
	if subject.Kind != "client" {
		tuple.Role = normalizeConflictSubjectRole(subject.Role)
	}
	return tuple, true
}

// frozenTupleEquals requires the exact fact tuple: name, canonical role and
// entity type, the sorted alias set, and the canonical identity anchors must
// all match. Names and aliases are identity facts re-frozen by the real
// intake check request, so client-management edits to any of them invalidate
// the old result; only an identical subject tuple may reuse it.
func frozenTupleEquals(frozen, current continuationFrozenSubject) bool {
	return frozen.Kind == current.Kind && frozen.Name == current.Name &&
		frozen.Role == current.Role && frozen.EntityType == current.EntityType &&
		len(frozen.Aliases) == len(current.Aliases) && sameContinuationStrings(frozen.Aliases, current.Aliases) &&
		len(frozen.Anchors) == len(current.Anchors) && sameContinuationAnchors(frozen.Anchors, current.Anchors)
}

func sameContinuationStrings(frozen, current []string) bool {
	if len(frozen) != len(current) {
		return false
	}
	for index, value := range frozen {
		if current[index] != value {
			return false
		}
	}
	return true
}

func sameContinuationAnchors(frozen, current map[string]string) bool {
	for key, value := range frozen {
		if current[key] != value {
			return false
		}
	}
	return true
}

// verifyContinuationFrozenIdentity re-proves that the current protected
// subjects are the SAME fact tuples the old check froze: client and every
// party must match by exact name plus canonical role, entity type, alias
// set, and the frozen identity anchor of the same identity type; the two
// sets must correspond one-to-one, so a frozen party deleted from the
// intake, an added party, or any changed fact fails closed. A frozen audit
// payload that cannot prove the complete chain (no anchors, missing role,
// malformed aliases) is a mismatch, never silently accepted. Digests never
// leave this function.
func verifyContinuationFrozenIdentity(frozenParameters string, evidence []continuationSubjectEvidence) error {
	frozen := parseSearchParametersMap(frozenParameters)
	if len(frozen) == 0 {
		return errContinuationIdentityChanged
	}
	frozenSubjects, ok := frozenContinuationSubjects(frozen)
	if !ok {
		return errContinuationIdentityChanged
	}
	hasFrozenClient, hasFrozenParty := false, false
	for _, subject := range frozenSubjects {
		if subject.Kind == "client" {
			hasFrozenClient = true
		} else {
			hasFrozenParty = true
		}
	}
	if !hasFrozenClient || !hasFrozenParty {
		return errContinuationIdentityChanged
	}
	consumed := make([]bool, len(frozenSubjects))
	for _, item := range evidence {
		current, tupleOK := continuationEvidenceTuple(item)
		if !tupleOK {
			return errContinuationIdentityChanged
		}
		matched := false
		for index, candidate := range frozenSubjects {
			if consumed[index] {
				continue
			}
			if frozenTupleEquals(candidate, current) {
				consumed[index] = true
				matched = true
				break
			}
		}
		if !matched {
			return errContinuationIdentityChanged
		}
	}
	for _, used := range consumed {
		if !used {
			return errContinuationIdentityChanged
		}
	}
	return nil
}

// mapContinuationGateError translates the shared gate rejection into the
// stable continuation reason codes. The message phrases are owned by this
// file, so the mapping cannot drift silently.
func mapContinuationGateError(err error) []string {
	var contractErr *ResubmitContractError
	if !errors.As(err, &contractErr) {
		return nil
	}
	switch contractErr.Message {
	case "接案缺少可验证的冲突检测绑定", "绑定的冲突检测记录不存在", "冲突检测记录未绑定当前接案":
		return []string{ContinuationReasonCheckMissing}
	case "绑定的冲突检测尚未完成":
		return []string{ContinuationReasonCheckIncomplete}
	case "冲突检测覆盖范围未确认完整":
		return []string{ContinuationReasonCoverageIncomplete}
	case "冲突检测尚未完成独立人工复核":
		return []string{ContinuationReasonReviewMissing}
	case "冲突复核结论为信息不足":
		return []string{ContinuationReasonReviewInsufficient}
	case "独立复核已确认存在利益冲突":
		return []string{ContinuationReasonReviewConflict}
	case "冲突复核已申请豁免":
		return []string{ContinuationReasonReviewWaiverPending}
	default:
		return []string{ContinuationReasonReviewUnresolved}
	}
}

func continuationSubjects(tx *gorm.DB, intakeID string) ([]continuationSubjectEvidence, error) {
	metadata := map[string]interface{}{}
	var metadataRaw struct {
		Metadata []byte
	}
	if err := tx.Table("case_intakes").Select("metadata").Where("id = ?", intakeID).Take(&metadataRaw).Error; err != nil {
		return nil, err
	}
	if len(metadataRaw.Metadata) > 0 {
		if err := json.Unmarshal(metadataRaw.Metadata, &metadata); err != nil {
			metadata = map[string]interface{}{}
		}
	}
	evidence := []continuationSubjectEvidence{}
	var clientMirror *continuationClientMirror
	// The frozen client alias fact comes from the intake metadata list; the
	// mirror row and the master must both be proven against that same
	// alias source so an alias cannot survive in one copy and vanish from
	// the other.
	clientAliasSource := continuationStringList(metadata["client_aliases"], metadata["clientAliases"])
	if tx.Migrator().HasTable("clients") {
		var clientRow struct {
			ClientID     *uint
			Name         *string
			Type         *string
			IdentityType *string
			Ciphertext   *string
			Digest       *string
			Aliases      *string
		}
		if err := tx.Table("case_intakes AS i").
			Select("i.client_id AS client_id, c.name AS name, c.type AS type, c.identity_type AS identity_type, "+
				"c.identity_number_ciphertext AS ciphertext, c.identity_number_digest AS digest, c.aliases AS aliases").
			Joins("LEFT JOIN clients c ON c.id = i.client_id AND c.deleted_at IS NULL").
			Where("i.id = ?", intakeID).Take(&clientRow).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, gorm.ErrRecordNotFound
			}
			return nil, err
		}
		if clientRow.ClientID != nil && *clientRow.ClientID > 0 {
			if clientRow.Name == nil || strings.TrimSpace(*clientRow.Name) == "" {
				return nil, fmt.Errorf("接案客户档案不可读")
			}
			identityType, ciphertext, digest := "", "", ""
			if clientRow.IdentityType != nil {
				identityType = strings.TrimSpace(*clientRow.IdentityType)
			}
			if clientRow.Ciphertext != nil {
				ciphertext = strings.TrimSpace(*clientRow.Ciphertext)
			}
			if clientRow.Digest != nil {
				digest = strings.TrimSpace(*clientRow.Digest)
			}
			identityType = continuationMasterIdentityType(identityType, continuationStringOrEmpty(clientRow.Type))
			// The full master alias set is the union of the intake
			// metadata list and the clients.aliases column - exactly the
			// copies the real StartIntakeConflictCheck freezes into the
			// audit record. The guarded mirror row stores ONLY the master
			// column value (prepareCaseIntakeClientPartyRow copies
			// clients.aliases), so the mirror holds the column aliases
			// alone; intake-scoped metadata extensions must not be pushed
			// into it or an unchanged intake can never prove its own
			// stored copy.
			clientColumnAliases := continuationClientColumnAliases(clientRow.Aliases)
			clientAliases := normalizeContinuationAliasSet(append(append([]string{}, clientAliasSource...), clientColumnAliases...))
			// Client entity type follows the exact StartIntakeConflictCheck
			// mapping from clients.type, so a re-run check would freeze the
			// same PERSON/COMPANY/ANY fact the tuple compare requires.
			clientType := strings.ToUpper(strings.TrimSpace(continuationStringOrEmpty(clientRow.Type)))
			switch clientType {
			case "企业", "公司", "COMPANY":
				clientType = "COMPANY"
			case "个人", "PERSON":
				clientType = "PERSON"
			default:
				clientType = "ANY"
			}
			masked, err := continuationMaskedIdentity(identityType, ciphertext, digest)
			if err != nil {
				return nil, err
			}
			evidence = append(evidence, continuationSubjectEvidence{
				Subject: ContinuationSubject{
					Kind:           "client",
					Name:           strings.TrimSpace(*clientRow.Name),
					Role:           "CLIENT",
					EntityType:     clientType,
					IdentityType:   identityType,
					IdentityMasked: masked,
					Aliases:        clientAliases,
				},
				IdentityDigest: digest,
			})
			clientMirror = &continuationClientMirror{
				name:         strings.TrimSpace(*clientRow.Name),
				identityType: identityType,
				digest:       digest,
				// The mirror compare is a set compare: both sides go through
				// the same normalize (trim, dedup, sort), so padding and
				// duplicate storage copies of one alias never fail the row,
				// while a genuinely new or removed value still does.
				columnAliases: normalizeContinuationAliasSet(clientColumnAliases),
			}
		}
	}
	if len(evidence) == 0 {
		// A normal intake always names one protected client subject; without
		// it the identity claim cannot be proven, so fail closed.
		return nil, fmt.Errorf("接案缺少可验证的客户主体，不能续办")
	}
	if tx.Migrator().HasTable("case_intake_parties") {
		var partyRows []struct {
			ID           uint
			EntityName   string
			EntityType   string
			PartyRole    string
			IdentityType string
			Ciphertext   string
			Digest       string
			Aliases      *string
			Metadata     []byte
		}
		if err := tx.Table("case_intake_parties").
			Select("id, entity_name, entity_type, party_role, identity_type, "+
				"identity_number_ciphertext AS ciphertext, identity_number_digest AS digest, aliases, metadata").
			Where("intake_id = ?", intakeID).Order("id ASC").Find(&partyRows).Error; err != nil {
			return nil, err
		}
		for _, row := range partyRows {
			name := strings.TrimSpace(row.EntityName)
			// StartIntakeConflictCheck skips any party row whose name equals
			// the client master name. The guarded save path stores that client
			// as a mirror row (party_role=client) copied from the protected
			// master, so continuation must prove the row against the master
			// before it may normalize it into the client subject; anything
			// else must stay visible as a party tuple and fail the frozen
			// one-to-one comparison.
			if strings.EqualFold(strings.TrimSpace(row.PartyRole), "client") {
				if clientMirror == nil {
					return nil, fmt.Errorf("接案缺少客户主档案事实，不能核验客户当事人镜像行")
				}
				// The guarded save only normalizes a row into the client
				// subject while the stored name still identifies the client.
				// Once the name changed, the row is no longer the mirror the
				// frozen check skipped by client name: fall through to the
				// normal party path so the renamed row appears as an extra
				// tuple and fails the one-to-one frozen comparison instead of
				// being silently swallowed by the mirror branch.
				if strings.EqualFold(name, clientMirror.name) {
					// The mirror row is a guarded copy of the client master, so
					// the alias fact must survive in the same shape the real save
					// wrote it: the comma-separated column, with the row metadata
					// list kept as an additional alias source so an alias stored
					// only in metadata cannot silently disappear.
					rowMetadata := map[string]interface{}{}
					if len(row.Metadata) > 0 {
						if err := json.Unmarshal(row.Metadata, &rowMetadata); err != nil {
							return nil, fmt.Errorf("解析客户当事人镜像行元数据失败：%w", err)
						}
					}
					rowAliases := normalizeContinuationAliasSet(append(
						parseContinuationStoredAliasList(row.Aliases),
						continuationStringList(rowMetadata["aliases"], rowMetadata["client_aliases"], rowMetadata["clientAliases"])...,
					))
					// Mirror-row readability: the row's alias column must equal
					// the master column split (tamper fails closed), while
					// intake-scoped metadata aliases stay out of the compare.
					if _, err := continuationMaskedIdentity(row.IdentityType, strings.TrimSpace(row.Ciphertext), strings.TrimSpace(row.Digest)); err != nil {
						return nil, err
					}
					if !continuationIntakeEntityTypeMatchesMaster(row.EntityType, clientMirror.identityType) ||
						strings.TrimSpace(row.IdentityType) != clientMirror.identityType ||
						strings.TrimSpace(row.Digest) != clientMirror.digest ||
						!sameContinuationStrings(rowAliases, clientMirror.columnAliases) {
						return nil, fmt.Errorf("客户当事人镜像行与客户主档案不一致，不能续办")
					}
					continue
				}
			}
			partyMetadata := map[string]interface{}{}
			if len(row.Metadata) > 0 {
				if err := json.Unmarshal(row.Metadata, &partyMetadata); err != nil {
					return nil, fmt.Errorf("解析接案当事人元数据失败：%w", err)
				}
			}
			// Business role follows the exact StartIntakeConflictCheck
			// derivation: the storage bucket party_role is only a fallback;
			// the reviewed business role lives in metadata.party_type.
			role := strings.ToUpper(firstNonEmpty(
				strings.TrimSpace(row.PartyRole),
				"RELATED_PARTY",
			))
			if strings.EqualFold(strings.TrimSpace(role), "RELATED_PARTY") {
				if businessRole := strings.ToUpper(strings.TrimSpace(firstNonEmpty(
					firstContinuationString(partyMetadata, "party_type"),
				))); businessRole != "" {
					role = businessRole
				}
			}
			identityType := strings.TrimSpace(row.IdentityType)
			ciphertext := strings.TrimSpace(row.Ciphertext)
			digest := strings.TrimSpace(row.Digest)
			// A real intake check rejects every party without a protected
			// identifier (INTAKE_PARTY_IDENTITY_REQUIRED), so a name-only
			// party row can never prove the frozen chain and must fail
			// closed instead of passing by display name.
			if name == "" || identityType == "" || ciphertext == "" || digest == "" {
				return nil, fmt.Errorf("接案当事人缺少受保护的可核验身份，不能续办")
			}
			partyAliases := normalizeContinuationAliasSet(append(
				parseContinuationStoredAliasList(row.Aliases),
				continuationStringList(partyMetadata["aliases"], partyMetadata["client_aliases"], partyMetadata["clientAliases"])...,
			))
			masked, err := continuationMaskedIdentity(row.IdentityType, row.Ciphertext, row.Digest)
			if err != nil {
				return nil, err
			}
			evidence = append(evidence, continuationSubjectEvidence{
				Subject: ContinuationSubject{
					Kind:           "party",
					PartyID:        row.ID,
					Name:           name,
					Role:           role,
					EntityType:     strings.TrimSpace(row.EntityType),
					IdentityType:   identityType,
					IdentityMasked: masked,
					Aliases:        partyAliases,
				},
				IdentityDigest: digest,
			})
		}
	}
	return evidence, nil
}

// continuationMaskedIdentity decrypts the stored ciphertext only long enough
// to verify it against the stored digest and produce the masked display
// value. Plaintext never leaves this function; a missing, unreadable, or
// tampered protected identity is an error, never a masked guess.
func continuationMaskedIdentity(identityType, ciphertext, digest string) (string, error) {
	if ciphertext == "" || digest == "" {
		return "", fmt.Errorf("受保护身份标识缺失，不能续办")
	}
	plaintext, err := security.DecryptIdentityNumber(ciphertext)
	if err != nil || strings.TrimSpace(plaintext) == "" {
		return "", fmt.Errorf("受保护身份标识无法安全读取，不能续办")
	}
	normalized := security.NormalizeIdentityNumber(identityType, plaintext)
	if normalized == "" {
		return "", fmt.Errorf("受保护身份标识为空，不能续办")
	}
	recomputed, err := security.IdentityDigest(normalized)
	if err != nil || recomputed != digest {
		return "", fmt.Errorf("受保护身份标识与登记摘要不一致，不能续办")
	}
	return security.MaskIdentityNumber(normalized), nil
}

// attachContinuationApprovalContext resolves the approval that still names
// this intake (approval metadata intake_id anchor) and its active
// intake_submission round. Missing rows leave the context unset.
func attachContinuationApprovalContext(tx *gorm.DB, intakeID string, decision *ContinuationDecision) error {
	if !tx.Migrator().HasTable("approval_requests") {
		return nil
	}
	var approvalRow struct {
		ID     string
		Status string
	}
	if err := tx.Table("approval_requests").
		Select("id, status").
		// Native JSON extraction works for PostgreSQL json/jsonb and the
		// current SQLite trial schema; formatting or nested keys cannot
		// change the exact top-level intake binding.
		Where("deleted_at IS NULL AND metadata->>'intake_id' = ?", intakeID).
		Order("created_at DESC").Order("id DESC").Limit(1).
		Take(&approvalRow).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return fmt.Errorf("读取接案审批上下文失败：%w", err)
	}
	ctx := &ContinuationApprovalContext{ApprovalID: approvalRow.ID, Status: approvalRow.Status}
	// The active submission pointer is the ONLY authority for which frozen
	// round this approval still owns (same rule as the structured resubmit
	// path). Reading it inside the transaction propagates failures, and a
	// pointer that does not resolve to an intake_submission snapshot of this
	// approval stays unresolved: falling back to the newest round would let
	// an unactivated draft round silently replace the authoritative one.
	if tx.Migrator().HasTable("approval_snapshots") {
		pointer, pointerErr := activeSnapshotIDTx(tx, approvalRow.ID)
		if pointerErr != nil {
			return fmt.Errorf("读取接案审批当前快照指针失败：%w", pointerErr)
		}
		if pointer != "" {
			var snapshotRow struct {
				ID            string
				SourceVersion int
			}
			if err := tx.Table("approval_snapshots").
				Select("id, source_version").
				Where("id = ? AND approval_request_id = ? AND snapshot_type = ?", pointer, approvalRow.ID, models.ApprovalSnapshotTypeIntakeSubmission).
				Take(&snapshotRow).Error; err == nil {
				ctx.SnapshotID = snapshotRow.ID
				ctx.Round = snapshotRow.SourceVersion
			} else if !errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("读取接案审批当前快照失败：%w", err)
			}
		}
	}
	decision.Approval = ctx
	return nil
}
