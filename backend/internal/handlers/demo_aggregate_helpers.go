package handlers

import (
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"law-oa-go/internal/security"
)

func intakeClientID(value interface{}) uint {
	switch typed := value.(type) {
	case uint:
		return typed
	case uint64:
		return uint(typed)
	case int:
		if typed > 0 {
			return uint(typed)
		}
	case int64:
		if typed > 0 {
			return uint(typed)
		}
	case float64:
		if typed > 0 && typed == float64(uint(typed)) {
			return uint(typed)
		}
	case json.Number:
		if parsed, err := strconv.ParseUint(string(typed), 10, 32); err == nil {
			return uint(parsed)
		}
	case string:
		if parsed, err := strconv.ParseUint(strings.TrimSpace(typed), 10, 32); err == nil {
			return uint(parsed)
		}
	}
	return 0
}

func objectValue(value interface{}) map[string]interface{} {
	switch typed := value.(type) {
	case map[string]interface{}:
		return typed
	case []byte:
		var result map[string]interface{}
		if json.Unmarshal(typed, &result) == nil {
			return result
		}
	case string:
		var result map[string]interface{}
		if json.Unmarshal([]byte(typed), &result) == nil {
			return result
		}
	}
	return map[string]interface{}{}
}

func valueString(values map[string]interface{}, key string) string {
	if values == nil {
		return ""
	}
	value := strings.TrimSpace(fmt.Sprint(values[key]))
	if value == "<nil>" {
		return ""
	}
	return value
}

// uintFromAny converts a loosely typed aggregate row value into a uint;
// unsupported shapes degrade to 0 instead of guessing an ID.
func uintFromAny(value interface{}) uint {
	switch typed := value.(type) {
	case uint:
		return typed
	case uint64:
		return uint(typed)
	case int:
		if typed > 0 {
			return uint(typed)
		}
	case int64:
		if typed > 0 {
			return uint(typed)
		}
	case float64:
		if typed > 0 && typed == float64(uint(typed)) {
			return uint(typed)
		}
	case json.Number:
		if parsed, err := strconv.ParseUint(string(typed), 10, 32); err == nil {
			return uint(parsed)
		}
	case string:
		if parsed, err := strconv.ParseUint(strings.TrimSpace(typed), 10, 32); err == nil {
			return uint(parsed)
		}
	}
	return 0
}

func subjectCaseIDFromAggregateSearchParameters(value interface{}) string {
	parameters := objectValue(value)
	for _, key := range []string{"subjectCaseId", "subject_case_id", "caseId", "case_id"} {
		if candidate := valueString(parameters, key); candidate != "" && candidate != "<nil>" {
			return candidate
		}
	}
	return ""
}

func subjectIntakeIDFromAggregateSearchParameters(value interface{}) string {
	parameters := objectValue(value)
	for _, key := range []string{"intakeId", "intake_id"} {
		if candidate := valueString(parameters, key); candidate != "" && candidate != "<nil>" {
			return candidate
		}
	}
	return ""
}

func stringListValue(value interface{}) []string {
	var result []string
	switch typed := value.(type) {
	case []interface{}:
		for _, item := range typed {
			if value := strings.TrimSpace(fmt.Sprint(item)); value != "" && value != "<nil>" {
				result = append(result, value)
			}
		}
	case []string:
		for _, item := range typed {
			if value := strings.TrimSpace(item); value != "" {
				result = append(result, value)
			}
		}
	case string:
		result = append(result, splitConflictAliases(typed)...)
	}
	return result
}

// deduplicateConflictAliasList trims, drops blanks, and removes duplicates
// while preserving first-seen order, so the frozen alias fact is a clean
// set regardless of how many storage copies carried the same value. The
// continuation gate sorts before comparing, so order is never part of the
// frozen contract.
func deduplicateConflictAliasList(values []string) []string {
	seen := map[string]struct{}{}
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
	return out
}

func stringMapValue(value interface{}) map[string]string {
	result := map[string]string{}
	if values, ok := value.(map[string]interface{}); ok {
		for key, item := range values {
			value := strings.TrimSpace(fmt.Sprint(item))
			if value != "" && value != "<nil>" {
				result[strings.ToLower(strings.TrimSpace(key))] = value
			}
		}
	}
	if len(result) == 0 {
		return nil
	}
	return result
}

// containsPlaintextIdentityField rejects identity numbers embedded in free-form
// intake metadata. Identity values must enter through the protected client or
// entity registry so they can be encrypted, hashed for lookup, and audited.
func containsPlaintextIdentityField(value interface{}) bool {
	sensitiveKeys := map[string]struct{}{
		"idcard": {}, "clientidcard": {}, "identitynumber": {}, "passport": {},
		"unifiedsocialcreditcode": {}, "socialcreditcode": {}, "taxid": {},
		"entitytaxid": {}, "clienttaxid": {}, "organizationcode": {}, "orgcode": {},
	}
	var visit func(interface{}) bool
	visit = func(current interface{}) bool {
		switch typed := current.(type) {
		case map[string]interface{}:
			for key, item := range typed {
				normalized := strings.ToLower(strings.NewReplacer("_", "", "-", "", " ", "").Replace(strings.TrimSpace(key)))
				if _, sensitive := sensitiveKeys[normalized]; sensitive {
					return true
				}
				if visit(item) {
					return true
				}
			}
		case []interface{}:
			for _, item := range typed {
				if visit(item) {
					return true
				}
			}
		case []map[string]interface{}:
			for _, item := range typed {
				if visit(item) {
					return true
				}
			}
		}
		return false
	}
	return visit(value)
}

func firstNonEmptyValue(values ...interface{}) interface{} {
	for _, value := range values {
		switch typed := value.(type) {
		case nil:
			continue
		case string:
			if strings.TrimSpace(typed) != "" {
				return value
			}
		case map[string]interface{}:
			if len(typed) > 0 {
				return value
			}
		case []interface{}:
			if len(typed) > 0 {
				return value
			}
		default:
			if rendered := strings.TrimSpace(fmt.Sprint(value)); rendered != "" && rendered != "<nil>" {
				return value
			}
		}
	}
	return nil
}

var allowedCaseIntakeCreateFields = map[string]struct{}{
	"client_id":   {},
	"title":       {},
	"case_type":   {},
	"priority":    {},
	"description": {},
	"metadata":    {},
}

var allowedCaseIntakePartyFields = map[string]struct{}{
	"case_id":        {},
	"intake_id":      {},
	"entity_name":    {},
	"entity_type":    {},
	"party_role":     {},
	"relation_depth": {},
	"metadata":       {},
	"created_at":     {},
}

var allowedCaseMaterialFields = map[string]struct{}{
	"case_id":       {},
	"intake_id":     {},
	"name":          {},
	"material_type": {},
	"status":        {},
	"required":      {},
	"storage_url":   {},
	"metadata":      {},
	"created_at":    {},
	"updated_at":    {},
}

// toUintValue parses one JSON-decoded id into a uint. Accepts the numeric
// shapes encoding/json produces plus decimal strings; anything else fails
// closed so a weird payload cannot smuggle an unusable id through the
// material merge.
func toUintValue(value interface{}) (uint, error) {
	switch v := value.(type) {
	case float64:
		if v < 1 || v != math.Trunc(v) {
			return 0, fmt.Errorf("not a material id")
		}
		return uint(v), nil
	case int:
		if v < 1 {
			return 0, fmt.Errorf("not a material id")
		}
		return uint(v), nil
	case uint:
		return v, nil
	case string:
		parsed, err := strconv.ParseUint(strings.TrimSpace(v), 10, 64)
		if err != nil || parsed == 0 {
			return 0, fmt.Errorf("not a material id")
		}
		return uint(parsed), nil
	default:
		return 0, fmt.Errorf("not a material id")
	}
}

var allowedCaseIntakeUpdateFields = map[string]struct{}{
	"client_id":   {},
	"title":       {},
	"case_type":   {},
	"priority":    {},
	"description": {},
	"metadata":    {},
}

// conflictRelevantMetadataKeys lists the metadata keys that feed the facts a
// conflict check runs against. Absence and empty are normalized to the same
// side so a client that stops sending a key does not look like a change.
var conflictRelevantMetadataKeys = []string{
	"business_area",
	"sub_area",
	"dispute_amount",
	"source_channel",
	"source_contact",
	"investment_agreement_date",
	"dispute_date",
	"breach_date",
	"proposed_filing_date",
	"jurisdiction",
	"billing_method",
	"fee_base",
	"contingency_rate",
	"minimum_fee",
	"lawyer_id",
	"case_tags",
}

// intakeConflictFactsSnapshot is the canonical comparable projection of the
// stored facts (from a case_intakes row) or the submitted facts (from an
// update payload). Values are compared as normalized strings; missing and
// empty normalize to "" so they match each other.
type intakeConflictFactsSnapshot struct {
	ClientID string
	Title    string
	CaseType string
	Metadata map[string]string
	Parties  []string
}

func intakeFactsScalar(value interface{}) string {
	switch typed := value.(type) {
	case nil:
		return ""
	case string:
		return strings.TrimSpace(typed)
	case float64:
		return strconv.FormatFloat(typed, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(typed)
	case json.Number:
		return strings.TrimSpace(typed.String())
	default:
		return strings.TrimSpace(fmt.Sprint(typed))
	}
}

// intakeFactsNumberString compares numeric-looking ids across payload
// (float64) and storage (uint/int64) encodings: "5" must equal 5.
func intakeFactsNumberString(value interface{}) string {
	switch typed := value.(type) {
	case nil:
		return ""
	case float64:
		return strconv.FormatFloat(typed, 'f', -1, 64)
	case json.Number:
		return strings.TrimSpace(typed.String())
	default:
		return strings.TrimSpace(fmt.Sprint(typed))
	}
}

// intakeFactsMetadataString decodes one metadata value into its comparable
// scalar string. Lists (case_tags) compare by their trimmed elements joined
// with "|"; scalars compare as trimmed strings. nil and "" normalize equal.
func intakeFactsMetadataString(value interface{}) string {
	switch typed := value.(type) {
	case []interface{}:
		items := make([]string, 0, len(typed))
		for _, item := range typed {
			rendered := strings.TrimSpace(fmt.Sprint(item))
			if rendered != "" && rendered != "<nil>" {
				items = append(items, rendered)
			}
		}
		sort.Strings(items)
		return strings.Join(items, "|")
	case nil:
		return ""
	default:
		rendered := strings.TrimSpace(fmt.Sprint(typed))
		if rendered == "<nil>" {
			return ""
		}
		return rendered
	}
}

// intakePartyConflictFingerprint builds the comparable projection of one
// party row: role, name, normalized type, identity type, the DIGEST of the
// normalized identity number (plaintext never enters the fingerprint), the
// normalized alias list, related-party depth, and the party_type metadata.
func intakePartyConflictFingerprint(party map[string]interface{}) string {
	role := strings.ToLower(strings.TrimSpace(stringValue(
		firstNonEmptyValue(party["party_role"], party["role"]),
		"related_party",
	)))
	entityType := normalizeIntakeEntityType(stringValue(
		firstNonEmptyValue(party["entity_type"], party["entityType"]), "",
	))
	identityType := strings.ToUpper(strings.TrimSpace(stringValue(
		firstNonEmptyValue(party["identity_type"], party["identityType"]), "",
	)))
	identityFingerprint := intakeIdentityFingerprint(
		identityType,
		stringValue(firstNonEmptyValue(party["identity_number"], party["identityNumber"]), ""),
	)
	aliases := make([]string, 0)
	for _, item := range stringListValue(firstNonEmptyValue(party["aliases"], party["alias"])) {
		aliases = append(aliases, strings.TrimSpace(item))
	}
	sort.Strings(aliases)
	partyMetadata := map[string]interface{}(nil)
	if typed, ok := party["metadata"].(map[string]interface{}); ok {
		partyMetadata = typed
	}
	return strings.Join([]string{
		role,
		strings.TrimSpace(stringValue(firstNonEmptyValue(party["entity_name"], party["entityName"]), "")),
		entityType,
		identityType,
		identityFingerprint,
		strings.Join(aliases, "|"),
		intakePartyDepthKey(firstNonEmptyValue(party["relation_depth"], party["relationDepth"])),
		intakeFactsScalar(partyMetadata["party_type"]),
	}, "")
}

// storedIntakePartyFingerprint reads one stored party row into the same
// projection as intakePartyConflictFingerprint. Identity is compared via the
// stored digest, so a round-trip through the protected columns compares
// equal when the underlying normalized number is unchanged. Stored rows keep
// metadata as a JSON string; the compact form always stores party_type.
func storedIntakePartyFingerprint(row map[string]interface{}) string {
	partyMetadata := map[string]interface{}{}
	switch typed := row["metadata"].(type) {
	case []byte:
		_ = json.Unmarshal(typed, &partyMetadata)
	case string:
		_ = json.Unmarshal([]byte(typed), &partyMetadata)
	case map[string]interface{}:
		partyMetadata = typed
	}
	aliases := make([]string, 0)
	for _, item := range stringListValue(row["aliases"]) {
		aliases = append(aliases, strings.TrimSpace(item))
	}
	sort.Strings(aliases)
	return strings.Join([]string{
		strings.ToLower(strings.TrimSpace(fmt.Sprint(row["party_role"]))),
		strings.TrimSpace(fmt.Sprint(row["entity_name"])),
		normalizeIntakeEntityType(fmt.Sprint(row["entity_type"])),
		strings.ToUpper(strings.TrimSpace(fmt.Sprint(row["identity_type"]))),
		strings.TrimSpace(fmt.Sprint(row["identity_number_digest"])),
		strings.Join(aliases, "|"),
		intakePartyDepthKey(row["relation_depth"]),
		intakeFactsScalar(partyMetadata["party_type"]),
	}, "")
}

// intakeIdentityFingerprint projects one identity value into its comparable
// digest form. Absent identity input normalizes to "", mirroring the write
// path where an empty identity_number means "inherit the stored identity".
// A provided but unnormalizable identity marks the fingerprint unavailable
// so the comparison fails closed instead of looking like "no identity".
func intakeIdentityFingerprint(identityType, identityRaw string) string {
	if strings.TrimSpace(identityType) == "" && strings.TrimSpace(identityRaw) == "" {
		return ""
	}
	normalized := security.NormalizeIdentityNumber(identityType, identityRaw)
	if strings.TrimSpace(normalized) == "" {
		return "unavailable"
	}
	digest, err := security.IdentityDigest(normalized)
	if err != nil {
		return "unavailable"
	}
	return digest
}

// intakeStableFingerprint removes the identity-digest slot from a party
// fingerprint so stored rows and payload parties compare on stable fields.
func intakeStableFingerprint(fingerprint string) string {
	parts := strings.Split(fingerprint, "\x1f")
	if len(parts) < 5 {
		return fingerprint
	}
	parts[4] = ""
	return strings.Join(parts, "\x1f")
}

// intakePartyDepthKey canonicalizes the related-party depth for comparison.
// The write path stores 0 when the input omits the depth, so absent, null,
// and explicit zero must project to the same key instead of reading as a
// changed fact.
func intakePartyDepthKey(value interface{}) string {
	rendered := intakeFactsNumberString(value)
	if rendered == "" || rendered == "<nil>" || rendered == "0" {
		return "0"
	}
	return rendered
}

// loadStoredIntakeConflictFacts reads the committed facts projection for the
// intake row plus its current party rows. Called on the handler's shared DB
// before the update transaction; the authoritative comparison still runs on
// the values this read produces, and the row lock inside the transaction
// serializes concurrent writers.
func (h *DemoAggregateHandler) loadStoredIntakeConflictFacts(existing map[string]interface{}) intakeConflictFactsSnapshot {
	snapshot := intakeConflictFactsSnapshot{
		ClientID: intakeFactsNumberString(existing["client_id"]),
		Title:    intakeFactsScalar(existing["title"]),
		CaseType: intakeFactsScalar(existing["case_type"]),
		Metadata: map[string]string{},
	}
	storedMetadata := map[string]interface{}{}
	switch typed := existing["metadata"].(type) {
	case []byte:
		_ = json.Unmarshal(typed, &storedMetadata)
	case string:
		_ = json.Unmarshal([]byte(typed), &storedMetadata)
	case map[string]interface{}:
		storedMetadata = typed
	}
	for _, key := range conflictRelevantMetadataKeys {
		snapshot.Metadata[key] = intakeFactsScalar(storedMetadata[key])
	}
	if h.tableExists("case_intake_parties") {
		rows, err := h.allStoredIntakePartyRows(existing)
		if err == nil {
			for _, row := range rows {
				snapshot.Parties = append(snapshot.Parties, storedIntakePartyFingerprint(row))
			}
		}
	}
	sort.Strings(snapshot.Parties)
	return snapshot
}

// submittedIntakeConflictFacts builds the same projection from the update
// payload. Fields the payload omits resolve to the stored value, because the
// update keeps stored values for absent keys.
func (h *DemoAggregateHandler) submittedIntakeConflictFacts(existing map[string]interface{}, rawPayload map[string]interface{}, parties []map[string]interface{}) intakeConflictFactsSnapshot {
	stored := h.loadStoredIntakeConflictFacts(existing)
	snapshot := intakeConflictFactsSnapshot{
		Metadata: map[string]string{},
	}
	if clientIDValue, present := rawPayload["client_id"]; present {
		snapshot.ClientID = intakeFactsNumberString(clientIDValue)
	} else {
		snapshot.ClientID = stored.ClientID
	}
	if value, present := rawPayload["title"]; present {
		snapshot.Title = intakeFactsScalar(value)
	} else {
		snapshot.Title = stored.Title
	}
	if value, present := rawPayload["case_type"]; present {
		snapshot.CaseType = intakeFactsScalar(value)
	} else {
		snapshot.CaseType = stored.CaseType
	}
	storedMetadata := map[string]interface{}{}
	switch typed := existing["metadata"].(type) {
	case []byte:
		_ = json.Unmarshal(typed, &storedMetadata)
	case string:
		_ = json.Unmarshal([]byte(typed), &storedMetadata)
	case map[string]interface{}:
		storedMetadata = typed
	}
	providedMetadata := map[string]interface{}{}
	if typed, ok := rawPayload["metadata"].(map[string]interface{}); ok {
		providedMetadata = typed
	}
	for _, key := range conflictRelevantMetadataKeys {
		if value, present := providedMetadata[key]; present {
			snapshot.Metadata[key] = intakeFactsScalar(value)
		} else {
			snapshot.Metadata[key] = stored.Metadata[key]
		}
	}
	// Mirror the write-path replacement semantics exactly: the rows are
	// replaced whenever the parties key is present and non-nil, so an
	// explicit empty array clears the party set and must compare as that
	// empty set (failing closed against a committed conflict result). Only
	// an absent (or null) key keeps the stored parties unchanged.
	partiesValue, partiesKeyPresent := rawPayload["parties"]
	partiesProvided := partiesKeyPresent && partiesValue != nil
	if partiesProvided {
		for _, party := range parties {
			partyFingerprint := intakePartyConflictFingerprint(party)
			if !caseIntakePartyIdentityInputPresent(party) {
				// A payload without identity input keeps the stored identity (the
				// same rule as the write path). Re-project the fingerprint from
				// the matching stored row so the inherited digest compares equal
				// instead of colliding with the non-empty stored digest.
				if storedParty, ok := h.fingerprintMatchedStoredPartyRow(existing, intakeStableFingerprint(partyFingerprint)); ok {
					partyFingerprint = storedIntakePartyFingerprint(storedParty)
				}
			}
			snapshot.Parties = append(snapshot.Parties, partyFingerprint)
		}
		sort.Strings(snapshot.Parties)
	} else {
		snapshot.Parties = stored.Parties
	}
	return snapshot
}

// fingerprintMatchedStoredPartyRow finds the committed party row that
// matches the payload party on every stable (non-identity) field: role,
// name, entity type, identity type, aliases, depth, and party metadata.
// Identity is excluded because a payload without identity input inherits
// the stored digest. Only a UNIQUE match inherits that identity; ambiguous
// or missing candidates fail closed so the save invalidates the committed
// conflict result instead of silently borrowing an identity.
func (h *DemoAggregateHandler) fingerprintMatchedStoredPartyRow(existing map[string]interface{}, stableFingerprint string) (map[string]interface{}, bool) {
	if !h.tableExists("case_intake_parties") {
		return nil, false
	}
	rows, err := h.allStoredIntakePartyRows(existing)
	if err != nil {
		return nil, false
	}
	matched := map[string]interface{}(nil)
	found := 0
	for _, row := range rows {
		if intakeStableFingerprint(storedIntakePartyFingerprint(row)) == stableFingerprint {
			matched = row
			found++
		}
	}
	if found != 1 {
		return nil, false
	}
	return matched, true
}

// allStoredIntakePartyRows reads every committed party row for one intake.
// The empty name and role intentionally bypass lookupCaseIntakePartyRows:
// that helper treats them as match filters, which would return nothing.
func (h *DemoAggregateHandler) allStoredIntakePartyRows(existing map[string]interface{}) ([]map[string]interface{}, error) {
	rows := []map[string]interface{}{}
	err := h.db.Table("case_intake_parties").
		Where("intake_id = ?", fmt.Sprint(existing["id"])).
		Find(&rows).Error
	return rows, err
}

// caseIntakeConflictFactsChanged compares the submitted projection against
// the stored projection and reports a real facts difference. It is the gate
// that decides whether a save may invalidate a committed conflict result.
func (h *DemoAggregateHandler) caseIntakeConflictFactsChanged(existing map[string]interface{}, rawPayload map[string]interface{}, parties []map[string]interface{}) bool {
	stored := h.loadStoredIntakeConflictFacts(existing)
	submitted := h.submittedIntakeConflictFacts(existing, rawPayload, parties)
	if stored.ClientID != submitted.ClientID || stored.Title != submitted.Title || stored.CaseType != submitted.CaseType {
		return true
	}
	for _, key := range conflictRelevantMetadataKeys {
		if stored.Metadata[key] != submitted.Metadata[key] {
			return true
		}
	}
	if len(stored.Parties) != len(submitted.Parties) {
		return true
	}
	for index := range stored.Parties {
		if stored.Parties[index] != submitted.Parties[index] {
			return true
		}
	}
	return false
}

var allowedDemoSumColumns = map[string]map[string]struct{}{
	"payments": {
		"amount": {},
	},
}

func (h *DemoAggregateHandler) first(table, id string, dest *map[string]interface{}) error {
	if !h.tableExists(table) {
		return gorm.ErrRecordNotFound
	}
	q := h.db.Table(table).Where("id = ?", id)
	if h.hasColumn(table, "deleted_at") {
		q = q.Where("deleted_at IS NULL")
	}
	return q.Take(dest).Error
}

func (h *DemoAggregateHandler) inboxFilter(where string) string {
	if h.hasColumn("inbox_items", "deleted_at") {
		return "deleted_at IS NULL AND " + where
	}
	return where
}

func (h *DemoAggregateHandler) recentRows(table, where string, limit int, args ...interface{}) []map[string]interface{} {
	if !h.tableExists(table) {
		return []map[string]interface{}{}
	}
	var rows []map[string]interface{}
	q := h.db.Table(table)
	if where != "" {
		if len(args) == 1 && args[0] == nil {
			q = q.Where(where)
		} else {
			q = q.Where(where, args...)
		}
	}
	if h.hasColumn(table, "created_at") {
		q = q.Order("created_at DESC")
	}
	_ = q.Limit(limit).Find(&rows).Error
	for index, row := range rows {
		rows[index] = sanitizeAggregateRow(table, row)
	}
	return rows
}

func (h *DemoAggregateHandler) recentRowsAny(tables []string, where string, limit int, args ...interface{}) []map[string]interface{} {
	for _, table := range tables {
		if h.tableExists(table) {
			return h.recentRows(table, where, limit, args...)
		}
	}
	return []map[string]interface{}{}
}

func (h *DemoAggregateHandler) count(table, where string, args ...interface{}) int64 {
	if !h.tableExists(table) {
		return 0
	}
	var total int64
	q := h.db.Table(table)
	if where != "" {
		if len(args) == 1 && args[0] == nil {
			q = q.Where(where)
		} else {
			q = q.Where(where, args...)
		}
	}
	_ = q.Count(&total).Error
	return total
}

func (h *DemoAggregateHandler) countAny(tables []string, where string, args ...interface{}) int64 {
	for _, table := range tables {
		if h.tableExists(table) {
			return h.count(table, where, args...)
		}
	}
	return 0
}

func (h *DemoAggregateHandler) distinctCount(table, column, where string) int64 {
	if !h.tableExists(table) || !h.hasColumn(table, column) {
		return 0
	}
	var total int64
	q := h.db.Table(table).Distinct(column)
	if where != "" {
		q = q.Where(where)
	}
	_ = q.Count(&total).Error
	return total
}

func (h *DemoAggregateHandler) sumAny(tables []string, column, where string, args ...interface{}) float64 {
	for _, table := range tables {
		if !h.isAllowedSumColumn(table, column) || !h.tableExists(table) || !h.hasColumn(table, column) {
			continue
		}
		var total float64
		q := h.db.Table(table).Select(clause.Expr{
			SQL:                "COALESCE(SUM(?), 0)",
			Vars:               []interface{}{clause.Column{Name: column}},
			WithoutParentheses: true,
		})
		if where != "" {
			if len(args) == 1 && args[0] == nil {
				q = q.Where(where)
			} else {
				q = q.Where(where, args...)
			}
		}
		_ = q.Scan(&total).Error
		return total
	}
	return 0
}

func (h *DemoAggregateHandler) groupedCounts(table, column, where string, limit int, args ...interface{}) []gin.H {
	if !h.tableExists(table) || !h.hasColumn(table, column) {
		return []gin.H{}
	}
	var rows []struct {
		Key   string `gorm:"column:metric_key"`
		Count int64  `gorm:"column:count"`
	}
	q := h.db.Table(table).Select(fmt.Sprintf("%s AS metric_key, COUNT(*) AS count", column))
	if where != "" {
		if len(args) == 1 && args[0] == nil {
			q = q.Where(where)
		} else {
			q = q.Where(where, args...)
		}
	}
	_ = q.Group(column).Order("count DESC").Limit(limit).Scan(&rows).Error
	result := make([]gin.H, 0, len(rows))
	for _, row := range rows {
		result = append(result, gin.H{"key": row.Key, "count": row.Count})
	}
	return result
}

func (h *DemoAggregateHandler) rbacRoleRows(limit int) []gin.H {
	if !h.tableExists("roles") || !h.hasColumn("roles", "code") {
		return h.groupedCounts("users", "role", "deleted_at IS NULL", limit)
	}

	var rows []struct {
		Key   string `gorm:"column:metric_key"`
		Count int64  `gorm:"column:count"`
	}

	query := h.db.Table("roles").Where("roles.deleted_at IS NULL")
	if h.hasColumn("roles", "status") {
		query = query.Where("roles.status = ?", "active")
	}

	if h.tableExists("user_roles") {
		query = query.
			Select("roles.code AS metric_key, COUNT(user_roles.user_id) AS count").
			Joins("LEFT JOIN user_roles ON user_roles.role_id = roles.id").
			Group("roles.id, roles.code, roles.sort_order").
			Order("roles.sort_order ASC, roles.id ASC")
	} else {
		query = query.
			Select("roles.code AS metric_key, 0 AS count").
			Order("roles.sort_order ASC, roles.id ASC")
	}

	_ = query.Limit(limit).Scan(&rows).Error
	result := make([]gin.H, 0, len(rows))
	for _, row := range rows {
		result = append(result, gin.H{"key": row.Key, "count": row.Count})
	}
	return result
}

func (h *DemoAggregateHandler) tableExists(table string) bool {
	return h.db != nil && h.db.Migrator().HasTable(table)
}

func (h *DemoAggregateHandler) hasColumn(table, column string) bool {
	return h.db != nil && h.db.Migrator().HasColumn(table, column)
}

// sanitizeAggregateRow is the last response boundary for legacy workbench
// aggregates. These handlers intentionally use map-backed queries for
// compatibility with several schema generations, so returning the map as-is
// would also return password hashes, identity ciphertext/digests, or raw
// workflow metadata that the screen never needs.
func sanitizeAggregateRow(table string, row map[string]interface{}) map[string]interface{} {
	if row == nil {
		return map[string]interface{}{}
	}
	result := make(map[string]interface{}, len(row))
	for key, value := range row {
		result[key] = value
	}

	switch strings.ToLower(strings.TrimSpace(table)) {
	case "users":
		for key := range result {
			if isAggregateSecretKey(key) {
				delete(result, key)
			}
		}
	case "clients":
		result = sanitizeClientAggregateRow(result)
	case "case_intake_parties", "case_materials", "case_intakes":
		if value, ok := result["metadata"]; ok {
			result["metadata"] = redactAggregateSensitiveValue(value)
		}
		identityPresent := false
		for _, key := range []string{
			"identifiers", "client_identifiers", "clientIdentifiers", "id_card", "id_card_ciphertext",
			"id_card_digest", "unified_social_credit_code", "identity_number", "identityNumber",
			"identity_number_ciphertext", "identity_number_digest",
		} {
			if value := strings.TrimSpace(fmt.Sprint(result[key])); value != "" && value != "<nil>" {
				identityPresent = true
			}
			delete(result, key)
		}
		if strings.EqualFold(strings.TrimSpace(table), "case_intake_parties") && identityPresent {
			result["identity_status"] = "已登记（受保护）"
		}
	case "approval_requests":
		// The approval workbench only needs summary columns. Full snapshots,
		// attachments and arbitrary metadata are served by their dedicated,
		// authorization-aware endpoints.
		for _, key := range []string{"metadata", "workflow_config", "attachments", "conflict_result", "approval_snapshot"} {
			delete(result, key)
		}
	case "system_settings":
		for key := range result {
			if isAggregateSecretKey(key) || strings.EqualFold(strings.TrimSpace(key), "value") || strings.EqualFold(strings.TrimSpace(key), "setting_value") {
				result[key] = "已配置（敏感值不在聚合接口返回）"
			}
		}
	case "risk_audit_events", "audit_logs":
		for _, key := range []string{"payload", "details", "metadata", "request_body", "response_body", "context"} {
			delete(result, key)
		}
	}
	return result
}

func sanitizeClientAggregateRow(row map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{}, len(row))
	for key, value := range row {
		result[key] = value
	}
	identityPresent := false
	for _, key := range []string{"id_card", "id_card_ciphertext", "id_card_digest", "identity_number", "identity_number_ciphertext", "identity_number_digest", "unified_social_credit_code", "unifiedSocialCreditCode", "identifiers"} {
		if value := strings.TrimSpace(fmt.Sprint(result[key])); value != "" && value != "<nil>" {
			identityPresent = true
		}
		delete(result, key)
	}
	if identityPresent {
		result["identity_status"] = "已登记（受保护）"
	}
	if aliases, ok := result["aliases"]; ok {
		result["aliases"] = stringListValue(aliases)
	}
	for _, key := range []string{"phone", "contact_phone"} {
		if value := strings.TrimSpace(fmt.Sprint(result[key])); value != "" && value != "<nil>" {
			result[key] = maskAggregatePhone(value)
		}
	}
	return result
}

func isAggregateSecretKey(key string) bool {
	normalized := strings.ToLower(strings.NewReplacer("_", "", "-", "", " ", "").Replace(strings.TrimSpace(key)))
	for _, marker := range []string{"password", "passwd", "secret", "token", "apikey", "privatekey", "ciphertext", "digest", "identifier", "identitynumber", "idcard", "socialcreditcode"} {
		if strings.Contains(normalized, marker) {
			return true
		}
	}
	return false
}

func redactAggregateSensitiveValue(value interface{}) interface{} {
	switch typed := value.(type) {
	case map[string]interface{}:
		result := make(map[string]interface{}, len(typed))
		for key, item := range typed {
			if isAggregateSecretKey(key) {
				continue
			}
			result[key] = redactAggregateSensitiveValue(item)
		}
		return result
	case []interface{}:
		result := make([]interface{}, len(typed))
		for index, item := range typed {
			result[index] = redactAggregateSensitiveValue(item)
		}
		return result
	default:
		return value
	}
}

func maskAggregatePhone(value string) string {
	value = strings.TrimSpace(value)
	if len(value) <= 7 {
		return "***"
	}
	return value[:3] + "****" + value[len(value)-4:]
}

func (h *DemoAggregateHandler) isAllowedSumColumn(table, column string) bool {
	columns, ok := allowedDemoSumColumns[table]
	if !ok {
		return false
	}
	_, ok = columns[column]
	return ok
}

func filterMap(input map[string]interface{}, allowed map[string]struct{}) map[string]interface{} {
	filtered := make(map[string]interface{}, len(input))
	for key, value := range input {
		if _, ok := allowed[key]; ok {
			filtered[key] = value
		}
	}
	return filtered
}

func cleanInsertMap(input gin.H) map[string]interface{} {
	cleaned := make(map[string]interface{}, len(input))
	for key, value := range input {
		if value != nil {
			cleaned[key] = value
		}
	}
	return cleaned
}

func stringValue(value interface{}, fallback string) string {
	if s, ok := value.(string); ok && s != "" {
		return s
	}
	return fallback
}

func jsonStringValue(value interface{}) string {
	if value == nil {
		return "{}"
	}
	if s, ok := value.(string); ok && s != "" {
		return s
	}
	data, err := json.Marshal(value)
	if err != nil {
		return "{}"
	}
	return string(data)
}

func mapSliceValue(value interface{}) []map[string]interface{} {
	items, ok := value.([]interface{})
	if !ok {
		return nil
	}
	rows := make([]map[string]interface{}, 0, len(items))
	for _, item := range items {
		row, ok := item.(map[string]interface{})
		if ok {
			rows = append(rows, row)
		}
	}
	return rows
}
