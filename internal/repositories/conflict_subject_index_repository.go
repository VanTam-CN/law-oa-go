package repositories

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"law-oa-go/internal/models"
	"law-oa-go/internal/security"
)

type conflictSubjectIndexSource struct {
	SourceType           string
	SourceID             string
	CaseID               string
	ClientID             string
	SubjectRole          string
	SubjectType          string
	OriginalName         string
	Aliases              []string
	RelationType         string
	IdentifierType       string
	IdentifierDigest     string
	IdentifierCiphertext string
	CaseNumber           string
	CaseTitle            string
	CaseType             string
	CaseDescription      string
	ClientName           string
	LawyerName           string
	CreatedAt            time.Time
}

// SyncConflictSubjectIndex materializes current source records into immutable
// versions. Changed source content receives a new version ID; old versions are
// retained so former names and former clients remain searchable.
func (r *conflictRepository) SyncConflictSubjectIndex(ctx context.Context) error {
	sources, err := r.collectConflictSubjectIndexSources(ctx)
	if err != nil {
		return err
	}
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, source := range sources {
			if err := persistConflictSubjectIndexSource(tx, source); err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *conflictRepository) collectConflictSubjectIndexSources(ctx context.Context) ([]conflictSubjectIndexSource, error) {
	if r == nil || r.db == nil {
		return nil, fmt.Errorf("冲突主体索引数据库未初始化")
	}
	for _, table := range []string{
		"conflict_subject_versions", "conflict_subject_identifiers", "cases", "clients", "entities",
		"case_parties", "entity_name_history", "entity_relations", "users",
	} {
		if !r.db.Migrator().HasTable(table) {
			return nil, fmt.Errorf("冲突主体索引依赖表未部署: %s", table)
		}
	}
	db := r.db.WithContext(ctx)
	var clients []models.Client
	if err := db.Where("deleted_at IS NULL").Find(&clients).Error; err != nil {
		return nil, fmt.Errorf("读取客户主体源失败: %w", err)
	}
	var cases []models.Case
	if err := db.Where("deleted_at IS NULL").Find(&cases).Error; err != nil {
		return nil, fmt.Errorf("读取案件主体源失败: %w", err)
	}
	var entities []models.Entity
	if err := db.Where("deleted_at IS NULL").Find(&entities).Error; err != nil {
		return nil, fmt.Errorf("读取结构化主体源失败: %w", err)
	}
	var parties []models.CaseParty
	if err := db.Where("deleted_at IS NULL").Find(&parties).Error; err != nil {
		return nil, fmt.Errorf("读取案件当事人源失败: %w", err)
	}
	var histories []models.EntityNameHistory
	if err := db.Where("deleted_at IS NULL").Find(&histories).Error; err != nil {
		return nil, fmt.Errorf("读取主体曾用名源失败: %w", err)
	}
	var relations []models.EntityRelation
	if err := db.Where("deleted_at IS NULL AND is_active = ?", true).Find(&relations).Error; err != nil {
		return nil, fmt.Errorf("读取主体关系源失败: %w", err)
	}
	var users []models.User
	if err := db.Where("deleted_at IS NULL").Find(&users).Error; err != nil {
		return nil, fmt.Errorf("读取律师主体源失败: %w", err)
	}

	clientByID := make(map[uint]models.Client, len(clients))
	for _, client := range clients {
		if strings.TrimSpace(client.Name) == "" {
			return nil, fmt.Errorf("客户 %d 缺少主体名称，无法完成冲突索引对账", client.ID)
		}
		if strings.TrimSpace(client.IDCard) != "" && strings.TrimSpace(client.IDCardDigest) == "" {
			return nil, fmt.Errorf("客户 %d 的身份标识尚未完成保护回填", client.ID)
		}
		clientByID[client.ID] = client
	}
	entityByID := make(map[uint]models.Entity, len(entities))
	historyByEntity := make(map[uint][]string)
	for _, entity := range entities {
		if strings.TrimSpace(entity.Name) == "" {
			return nil, fmt.Errorf("主体 %d 缺少主体名称，无法完成冲突索引对账", entity.ID)
		}
		if strings.TrimSpace(entity.IdentityNumber) != "" && strings.TrimSpace(entity.IdentityNumberDigest) == "" {
			return nil, fmt.Errorf("主体 %d 的身份标识尚未完成保护回填", entity.ID)
		}
		entityByID[entity.ID] = entity
	}
	for _, history := range histories {
		if _, ok := entityByID[history.EntityID]; !ok {
			return nil, fmt.Errorf("主体曾用名 %d 引用了不存在的主体 %d", history.ID, history.EntityID)
		}
		if value := strings.TrimSpace(history.OldName); value != "" {
			historyByEntity[history.EntityID] = append(historyByEntity[history.EntityID], value)
		}
		if value := strings.TrimSpace(history.NewName); value != "" {
			historyByEntity[history.EntityID] = append(historyByEntity[history.EntityID], value)
		}
	}
	caseByID := make(map[uint]models.Case, len(cases))
	for _, matter := range cases {
		caseByID[matter.ID] = matter
		if _, ok := clientByID[matter.ClientID]; !ok {
			return nil, fmt.Errorf("案件 %d 引用了不存在的客户主体 %d", matter.ID, matter.ClientID)
		}
	}
	userByID := make(map[uint]models.User, len(users))
	for _, user := range users {
		userByID[user.ID] = user
	}

	var sources []conflictSubjectIndexSource
	for _, matter := range cases {
		client := clientByID[matter.ClientID]
		lawyer := userByID[matter.LawyerID]
		sources = append(sources, conflictSubjectIndexSource{
			SourceType: "CLIENT_ARCHIVE_CASE", SourceID: fmt.Sprintf("case:%d:client:%d", matter.ID, matter.ClientID),
			CaseID: fmt.Sprint(matter.ID), ClientID: fmt.Sprint(matter.ClientID), SubjectRole: "CLIENT",
			SubjectType: indexSubjectType(client.Type), OriginalName: client.Name, Aliases: nonEmptyAliases(client.Company),
			IdentifierType: "ID_CARD", IdentifierDigest: client.IDCardDigest, IdentifierCiphertext: client.IDCardCiphertext,
			CaseNumber: matter.CaseNumber, CaseTitle: matter.Title, CaseType: matter.CaseType, CaseDescription: matter.Description,
			ClientName: client.Name, LawyerName: lawyer.Name, CreatedAt: matter.CreatedAt,
		})
	}
	clientsWithCases := make(map[uint]bool)
	for _, matter := range cases {
		clientsWithCases[matter.ClientID] = true
	}
	for _, client := range clients {
		if clientsWithCases[client.ID] {
			continue
		}
		sources = append(sources, conflictSubjectIndexSource{
			SourceType: "CLIENT_ARCHIVE", SourceID: fmt.Sprintf("client:%d", client.ID), ClientID: fmt.Sprint(client.ID),
			SubjectRole: "CLIENT", SubjectType: indexSubjectType(client.Type), OriginalName: client.Name,
			Aliases: nonEmptyAliases(client.Company), IdentifierType: "ID_CARD", IdentifierDigest: client.IDCardDigest,
			IdentifierCiphertext: client.IDCardCiphertext, ClientName: client.Name, CreatedAt: client.CreatedAt,
		})
	}

	// The subject registry is independent from case history. An entity that has
	// not yet been attached to a case must still be searchable firm-wide.
	for _, entity := range entities {
		aliases := append([]string{}, splitAliases(entity.Alias)...)
		aliases = append(aliases, historyByEntity[entity.ID]...)
		sources = append(sources, conflictSubjectIndexSource{
			SourceType: "ENTITY_REGISTRY", SourceID: fmt.Sprintf("entity:%d", entity.ID), SubjectRole: "SUBJECT",
			SubjectType: indexSubjectType(string(entity.EntityType)), OriginalName: entity.Name, Aliases: aliases,
			IdentifierType: string(entity.IdentityType), IdentifierDigest: entity.IdentityNumberDigest, IdentifierCiphertext: entity.IdentityNumberCiphertext,
			CreatedAt: entity.CreatedAt,
		})
	}

	for _, party := range parties {
		matter, caseOK := caseByID[party.CaseID]
		if !caseOK {
			return nil, fmt.Errorf("案件当事人 %d 引用了不存在的案件 %d", party.ID, party.CaseID)
		}
		entity, entityOK := entityByID[party.EntityID]
		if !entityOK {
			return nil, fmt.Errorf("案件当事人 %d 引用了不存在的主体 %d", party.ID, party.EntityID)
		}
		client, clientOK := clientByID[matter.ClientID]
		if !clientOK {
			return nil, fmt.Errorf("案件 %d 的客户主体 %d 不存在", matter.ID, matter.ClientID)
		}
		lawyer := userByID[matter.LawyerID]
		aliases := append([]string{}, splitAliases(entity.Alias)...)
		aliases = append(aliases, historyByEntity[entity.ID]...)
		sources = append(sources, conflictSubjectIndexSource{
			SourceType: "ENTITY_CASE_PARTY", SourceID: fmt.Sprintf("case:%d:entity:%d:party:%d", matter.ID, entity.ID, party.ID),
			CaseID: fmt.Sprint(matter.ID), ClientID: fmt.Sprint(matter.ClientID), SubjectRole: indexPartyRole(party),
			SubjectType: indexSubjectType(string(entity.EntityType)), OriginalName: entity.Name, Aliases: aliases,
			IdentifierType: string(entity.IdentityType), IdentifierDigest: entity.IdentityNumberDigest, IdentifierCiphertext: entity.IdentityNumberCiphertext,
			CaseNumber: matter.CaseNumber, CaseTitle: matter.Title, CaseType: matter.CaseType, CaseDescription: matter.Description,
			ClientName: client.Name, LawyerName: lawyer.Name, CreatedAt: matter.CreatedAt,
		})
	}

	// Keep relation evidence searchable even when the relation has not yet been
	// attached to a case party. Case-bound relation snapshots are added below.
	for _, relation := range relations {
		source, target := entityByID[relation.SourceEntityID], entityByID[relation.TargetEntityID]
		if source.ID == 0 || target.ID == 0 {
			return nil, fmt.Errorf("主体关系 %d 引用了不存在的主体", relation.ID)
		}
		sources = append(sources, conflictSubjectIndexSource{
			SourceType: "RELATED_ENTITY", SourceID: fmt.Sprintf("relation:%d:entity:%d", relation.ID, target.ID),
			SubjectRole: "RELATED_PARTY", SubjectType: indexSubjectType(string(target.EntityType)), OriginalName: target.Name,
			Aliases:      append(append([]string{}, splitAliases(target.Alias)...), historyByEntity[target.ID]...),
			RelationType: string(relation.RelationType), IdentifierType: string(target.IdentityType),
			IdentifierDigest: target.IdentityNumberDigest, IdentifierCiphertext: target.IdentityNumberCiphertext, CreatedAt: relation.CreatedAt,
		})
	}
	return sources, nil
}

type conflictSubjectIndexMaterialized struct {
	SourceType  string
	SourceID    string
	VersionID   string
	VersionHash string
	SubjectKey  string
	Normalized  string
	Aliases     []string
	Raw         string
}

// ReconcileConflictSubjectIndex is the explicit historical build path. Dry
// runs collect and count source records without writing. Apply runs create
// durable RUNNING/COMPLETED/FAILED evidence and verify every expected version
// is present in the immutable index before returning success.
func (r *conflictRepository) ReconcileConflictSubjectIndex(ctx context.Context, actorID uint, evidenceReference string, apply bool) ([]models.ConflictIndexBuildRun, error) {
	if r == nil || r.db == nil {
		return nil, fmt.Errorf("冲突主体索引数据库未初始化")
	}
	sources, err := r.collectConflictSubjectIndexSources(ctx)
	if err != nil {
		return nil, err
	}
	evidenceReference = strings.TrimSpace(evidenceReference)
	if apply && evidenceReference == "" {
		return nil, fmt.Errorf("写入索引对账运行记录必须提供凭证引用")
	}
	materializedByScope := make(map[string][]conflictSubjectIndexMaterialized)
	allFingerprints := make([]string, 0, len(sources))
	for _, source := range sources {
		materialized, err := materializeConflictSubjectIndexSource(source)
		if err != nil {
			return nil, err
		}
		if materialized == nil {
			continue
		}
		allFingerprints = append(allFingerprints, materialized.VersionID+":"+materialized.VersionHash)
		for _, scopeType := range conflictIndexScopeTypes(source) {
			materializedByScope[scopeType] = append(materializedByScope[scopeType], *materialized)
		}
	}
	sort.Strings(allFingerprints)
	globalSum := sha256.Sum256([]byte(strings.Join(allFingerprints, "\n")))
	globalSourceVersion := "index-" + fmt.Sprintf("%x", globalSum[:])
	startedAt := time.Now()
	runs := make([]models.ConflictIndexBuildRun, 0, 4)
	for _, scopeType := range []string{"CASE_ARCHIVE", "CLIENT_ARCHIVE", "SUBJECT_REGISTRY", "RELATION_ARCHIVE"} {
		items := materializedByScope[scopeType]
		fingerprints := make([]string, 0, len(items))
		for _, item := range items {
			fingerprints = append(fingerprints, item.VersionID+":"+item.VersionHash)
		}
		sort.Strings(fingerprints)
		scopeSum := sha256.Sum256([]byte(strings.Join(fingerprints, "\n")))
		run := models.ConflictIndexBuildRun{
			ID: stableConflictP0ID("IDX_RUN", scopeType+":"+globalSourceVersion), ScopeType: scopeType,
			SourceVersion: globalSourceVersion, Status: models.ConflictIndexBuildRunning,
			SourceRecordCount: int64(len(items)), ReconciliationHash: fmt.Sprintf("%x", scopeSum[:]),
			EvidenceReference: evidenceReference, StartedAt: startedAt, CreatedAt: startedAt, UpdatedAt: startedAt,
		}
		if actorID != 0 {
			run.CreatedBy = &actorID
		}
		runs = append(runs, run)
		if !apply {
			runs[len(runs)-1].Status = "DRY_RUN"
			runs[len(runs)-1].IndexedRecordCount = int64(len(items))
			continue
		}
	}
	if !apply {
		return runs, nil
	}

	for _, run := range runs {
		if err := r.db.WithContext(ctx).Clauses(clause.OnConflict{UpdateAll: true}).Create(&run).Error; err != nil {
			return nil, fmt.Errorf("创建索引对账运行记录失败: %w", err)
		}
	}
	if err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, source := range sources {
			if err := persistConflictSubjectIndexSource(tx, source); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		message := err.Error()
		for _, run := range runs {
			_ = r.db.WithContext(ctx).Model(&models.ConflictIndexBuildRun{}).Where("id = ?", run.ID).Updates(map[string]interface{}{
				"status": models.ConflictIndexBuildFailed, "error_message": message, "updated_at": time.Now(),
			}).Error
		}
		return nil, err
	}

	for index := range runs {
		items := materializedByScope[runs[index].ScopeType]
		ids := make([]string, 0, len(items))
		for _, item := range items {
			ids = append(ids, item.VersionID)
		}
		var indexed int64
		if len(ids) > 0 {
			if err := r.db.WithContext(ctx).Model(&models.ConflictSubjectVersion{}).Where("id IN ?", ids).Count(&indexed).Error; err != nil {
				return nil, fmt.Errorf("核对索引对账记录失败: %w", err)
			}
		}
		runs[index].IndexedRecordCount = indexed
		runs[index].MissingRecordCount = runs[index].SourceRecordCount - indexed
		if runs[index].MissingRecordCount < 0 {
			runs[index].MissingRecordCount = 0
		}
		if indexed != runs[index].SourceRecordCount {
			message := fmt.Sprintf("索引记录数不闭合: source=%d indexed=%d", runs[index].SourceRecordCount, indexed)
			_ = r.db.WithContext(ctx).Model(&models.ConflictIndexBuildRun{}).Where("id = ?", runs[index].ID).Updates(map[string]interface{}{
				"status": models.ConflictIndexBuildFailed, "indexed_record_count": indexed, "missing_record_count": runs[index].MissingRecordCount,
				"error_message": message, "updated_at": time.Now(),
			}).Error
			return nil, fmt.Errorf("%s: %s", runs[index].ScopeType, message)
		}
		completedAt := time.Now()
		runs[index].Status = models.ConflictIndexBuildComplete
		runs[index].CompletedAt = &completedAt
		runs[index].UpdatedAt = completedAt
		if err := r.db.WithContext(ctx).Model(&models.ConflictIndexBuildRun{}).Where("id = ?", runs[index].ID).Updates(map[string]interface{}{
			"status": models.ConflictIndexBuildComplete, "indexed_record_count": indexed, "missing_record_count": 0,
			"completed_at": completedAt, "updated_at": completedAt, "error_message": "",
		}).Error; err != nil {
			return nil, fmt.Errorf("完成索引对账运行记录失败: %w", err)
		}
	}
	return runs, nil
}

func conflictIndexScopeTypes(source conflictSubjectIndexSource) []string {
	scopes := make([]string, 0, 2)
	if strings.TrimSpace(source.CaseID) != "" {
		scopes = append(scopes, "CASE_ARCHIVE")
	}
	switch source.SourceType {
	case "CLIENT_ARCHIVE", "CLIENT_ARCHIVE_CASE":
		scopes = append(scopes, "CLIENT_ARCHIVE")
	case "ENTITY_REGISTRY", "ENTITY_CASE_PARTY":
		scopes = append(scopes, "SUBJECT_REGISTRY")
	case "RELATED_ENTITY":
		scopes = append(scopes, "RELATION_ARCHIVE")
	}
	return scopes
}

func materializeConflictSubjectIndexSource(source conflictSubjectIndexSource) (*conflictSubjectIndexMaterialized, error) {
	entityType := source.SubjectType
	normalized := normalizeConflictIndexName(source.OriginalName, entityType)
	if normalized == "" {
		return nil, nil
	}
	aliases := normalizeConflictIndexAliases(source.Aliases, entityType, normalized)
	versionPayload := map[string]interface{}{
		"sourceType": source.SourceType, "sourceID": source.SourceID, "caseID": source.CaseID, "clientID": source.ClientID,
		"subjectRole": source.SubjectRole, "subjectType": entityType, "originalName": source.OriginalName,
		"normalizedName": normalized, "aliases": aliases, "relationType": source.RelationType,
		"caseNumber": source.CaseNumber, "caseTitle": source.CaseTitle, "caseType": source.CaseType,
		"caseDescription": source.CaseDescription, "clientName": source.ClientName, "lawyerName": source.LawyerName,
		"createdAt": source.CreatedAt,
	}
	if strings.TrimSpace(source.IdentifierDigest) != "" {
		// The protected digest is not exposed in the snapshot. It is included
		// only through a stable fingerprint so identity corrections create a
		// new immutable version instead of reusing the old one.
		versionPayload["identityVersion"] = stableConflictP0ID("IDENTITY", canonicalIndexIdentifier(source.IdentifierType)+":"+source.IdentifierDigest)
	}
	raw, err := json.Marshal(versionPayload)
	if err != nil {
		return nil, fmt.Errorf("序列化主体索引版本失败: %w", err)
	}
	versionHash := stableConflictP0ID("V", string(raw))
	subjectKey := source.SourceType + ":" + source.SourceID
	return &conflictSubjectIndexMaterialized{
		SourceType: source.SourceType, SourceID: source.SourceID,
		VersionID: stableConflictP0ID("IDX", subjectKey+":"+versionHash), VersionHash: versionHash, SubjectKey: subjectKey,
		Normalized: normalized, Aliases: aliases, Raw: string(raw),
	}, nil
}

func persistConflictSubjectIndexSource(tx *gorm.DB, source conflictSubjectIndexSource) error {
	materialized, err := materializeConflictSubjectIndexSource(source)
	if err != nil {
		return err
	}
	if materialized == nil {
		return nil
	}
	entityType := source.SubjectType
	normalized := materialized.Normalized
	aliases := materialized.Aliases
	raw := []byte(materialized.Raw)
	versionHash := materialized.VersionHash
	subjectKey := materialized.SubjectKey
	versionID := materialized.VersionID
	var existing int64
	if err := tx.Model(&models.ConflictSubjectVersion{}).Where("id = ?", versionID).Count(&existing).Error; err != nil {
		return fmt.Errorf("检查主体索引版本失败: %w", err)
	}
	if existing == 0 {
		var previousVersion int
		if err := tx.Model(&models.ConflictSubjectVersion{}).Where("subject_key = ?", subjectKey).Select("COALESCE(MAX(version_number), 0)").Scan(&previousVersion).Error; err != nil {
			return fmt.Errorf("读取主体索引版本号失败: %w", err)
		}
		version := &models.ConflictSubjectVersion{
			ID: versionID, SubjectKey: subjectKey, SourceType: source.SourceType, SourceID: source.SourceID,
			CaseID: source.CaseID, ClientID: source.ClientID, SubjectRole: source.SubjectRole, SubjectType: entityType,
			OriginalName: source.OriginalName, NormalizedName: normalized, AliasSnapshot: mustJSON(aliases),
			SourceVersion: versionHash, VersionNumber: previousVersion + 1, Verification: "SOURCE_REGISTERED",
			Snapshot: string(raw), CreatedAt: time.Now(),
		}
		result := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(version)
		if result.Error != nil {
			return fmt.Errorf("保存主体索引版本失败: %w", result.Error)
		}
	}
	if strings.TrimSpace(source.IdentifierDigest) != "" {
		identifier := &models.ConflictSubjectIdentifier{
			ID: stableConflictP0ID("IDXID", versionID+":"+source.IdentifierType+":"+source.IdentifierDigest), SubjectVersionID: versionID,
			IdentifierType: canonicalIndexIdentifier(source.IdentifierType), Digest: source.IdentifierDigest,
			Ciphertext: source.IdentifierCiphertext, Verification: "SOURCE_REGISTERED", SourceReference: source.SourceType + ":" + source.SourceID, CreatedAt: time.Now(),
		}
		if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(identifier).Error; err != nil {
			return fmt.Errorf("保存主体索引身份标识失败: %w", err)
		}
	}
	return nil
}

// SearchConflictSubjectIndex reads only historical source snapshots. The
// current request snapshots use source_type=CONFLICT_CHECK_SUBJECT and are
// intentionally excluded from the search source.
func (r *conflictRepository) SearchConflictSubjectIndex(ctx context.Context, subjects []models.ConflictNormalizedSubject, clientID string) ([]ConflictP0SubjectIndexHit, error) {
	var versions []models.ConflictSubjectVersion
	if err := r.db.WithContext(ctx).Where("source_type <> ?", "CONFLICT_CHECK_SUBJECT").Order("created_at DESC").Find(&versions).Error; err != nil {
		return nil, fmt.Errorf("读取全所冲突主体索引失败: %w", err)
	}
	var identifiers []models.ConflictSubjectIdentifier
	if err := r.db.WithContext(ctx).Find(&identifiers).Error; err != nil {
		return nil, fmt.Errorf("读取全所冲突主体身份索引失败: %w", err)
	}
	identifierByVersion := make(map[string][]models.ConflictSubjectIdentifier)
	for _, identifier := range identifiers {
		identifierByVersion[identifier.SubjectVersionID] = append(identifierByVersion[identifier.SubjectVersionID], identifier)
	}
	inputIDs := make([]map[string]string, len(subjects))
	for index, subject := range subjects {
		inputIDs[index] = make(map[string]string, len(subject.Identifiers))
		for kind, value := range subject.Identifiers {
			digest, err := security.IdentityDigest(value)
			if err != nil {
				return nil, fmt.Errorf("计算冲突主体身份摘要失败: %w", err)
			}
			inputIDs[index][canonicalIndexIdentifier(kind)] = digest
		}
	}

	hits := make([]ConflictP0SubjectIndexHit, 0)
	seen := make(map[string]struct{})
	for _, version := range versions {
		if strings.TrimSpace(clientID) != "" && strings.TrimSpace(version.ClientID) == strings.TrimSpace(clientID) {
			continue
		}
		aliases := []string{}
		_ = json.Unmarshal([]byte(version.AliasSnapshot), &aliases)
		for index, subject := range subjects {
			matchType, ruleCode, _ := matchConflictIndexSubject(version, aliases, subject, inputIDs[index], identifierByVersion[version.ID])
			if matchType == "" {
				continue
			}
			key := version.ID + ":" + subject.OriginalName + ":" + matchType
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			hits = append(hits, ConflictP0SubjectIndexHit{
				Version: version, Identifiers: identifierByVersion[version.ID], MatchType: matchType,
				RuleCode: ruleCode, RequestedParty: subject.OriginalName, MatchedName: version.OriginalName,
				MatchSource: version.SourceType, RelationType: relationTypeFromSnapshot(version.Snapshot),
			})
		}
	}
	return hits, nil
}

func matchConflictIndexSubject(version models.ConflictSubjectVersion, aliases []string, subject models.ConflictNormalizedSubject, inputIDs map[string]string, storedIDs []models.ConflictSubjectIdentifier) (string, string, string) {
	for _, stored := range storedIDs {
		if inputIDs[canonicalIndexIdentifier(stored.IdentifierType)] != "" && strings.EqualFold(inputIDs[canonicalIndexIdentifier(stored.IdentifierType)], stored.Digest) {
			return "EXACT", "STRUCTURED_IDENTITY_EXACT", subject.OriginalName
		}
	}
	requested := normalizeConflictIndexName(subject.OriginalName, subject.EntityType)
	if requested == "" {
		return "", "", ""
	}
	nameExact := strings.EqualFold(version.NormalizedName, requested)
	aliasExact := false
	for _, alias := range aliases {
		candidate := normalizeConflictIndexName(alias, subject.EntityType)
		if strings.EqualFold(candidate, requested) {
			aliasExact = true
			break
		}
	}
	if nameExact || aliasExact {
		if strings.EqualFold(indexSubjectType(subject.EntityType), "PERSON") && len(subject.Identifiers) == 0 {
			return "CANDIDATE", "PERSON_NAME_ONLY_INSUFFICIENT", subject.OriginalName
		}
		if version.SourceType == "RELATED_ENTITY" {
			return "RELATION", "RELATED_ENTITY_ADVERSE_REVIEW", subject.OriginalName
		}
		if aliasExact {
			return "CANDIDATE", "FORMER_NAME_CANDIDATE_REVIEW", subject.OriginalName
		}
		if version.SourceType == "CLIENT_ARCHIVE" || version.SourceType == "CLIENT_ARCHIVE_CASE" {
			return "CANDIDATE", "CLIENT_ARCHIVE_NAME_CANDIDATE", subject.OriginalName
		}
		return "CANDIDATE", "SUBJECT_CANDIDATE_REVIEW", subject.OriginalName
	}
	if len([]rune(requested)) >= 3 && len([]rune(version.NormalizedName)) >= 3 &&
		(strings.Contains(version.NormalizedName, requested) || strings.Contains(requested, version.NormalizedName)) {
		return "CANDIDATE", "SUBJECT_CANDIDATE_REVIEW", subject.OriginalName
	}
	return "", "", ""
}

func relationTypeFromSnapshot(snapshot string) string {
	var value struct {
		RelationType string `json:"relationType"`
	}
	_ = json.Unmarshal([]byte(snapshot), &value)
	return value.RelationType
}

func indexSubjectType(value string) string {
	switch strings.ToUpper(strings.TrimSpace(value)) {
	case "PERSON", "INDIVIDUAL", "自然人":
		return "PERSON"
	case "COMPANY", "LEGAL_PERSON", "ORGANIZATION", "法人", "企业":
		return "COMPANY"
	default:
		return "ANY"
	}
}

func indexPartyRole(party models.CaseParty) string {
	switch strings.ToUpper(strings.TrimSpace(string(party.PartyType))) {
	case "CLIENT":
		return "CLIENT"
	case "OPPOSING", "CO_DEFENDANT":
		return "OPPOSING_PARTY"
	default:
		return "RELATED_PARTY"
	}
}

func nonEmptyAliases(values ...string) []string {
	aliases := make([]string, 0, len(values))
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			aliases = append(aliases, value)
		}
	}
	return aliases
}

func splitAliases(value string) []string {
	parts := strings.FieldsFunc(value, func(r rune) bool { return r == ',' || r == '，' || r == ';' || r == '；' || r == '|' })
	return nonEmptyAliases(parts...)
}

func normalizeConflictIndexAliases(values []string, entityType, normalized string) []string {
	seen := map[string]struct{}{}
	result := make([]string, 0, len(values))
	for _, value := range values {
		candidate := normalizeConflictIndexName(value, entityType)
		if candidate == "" || candidate == normalized {
			continue
		}
		if _, ok := seen[candidate]; ok {
			continue
		}
		seen[candidate] = struct{}{}
		result = append(result, candidate)
	}
	return result
}

func normalizeConflictIndexName(value, entityType string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = strings.NewReplacer("（", "(", "）", ")", "【", "", "】", "", "[", "", "]", "", "·", "", "•", "", "-", "", "_", "", ".", "", "，", "", ",", "", " ", "", "　", "", "\t", "", "\n", "").Replace(value)
	if strings.EqualFold(indexSubjectType(entityType), "PERSON") {
		return value
	}
	for _, suffix := range []string{"股份有限公司", "有限责任公司", "集团有限公司", "有限公司", "律师事务所", "集团", "控股", "公司", "律所"} {
		value = strings.TrimSuffix(value, suffix)
	}
	return strings.TrimSpace(value)
}

func canonicalIndexIdentifier(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "id_card", "idcard", "身份证", "身份证号":
		return "ID_CARD"
	case "unified_social_credit_code", "social_credit_code", "统一社会信用代码":
		return "SOCIAL_CREDIT_CODE"
	case "business_license", "营业执照":
		return "BUSINESS_LICENSE"
	default:
		return strings.ToUpper(strings.TrimSpace(value))
	}
}

func mustJSON(value interface{}) string {
	raw, err := json.Marshal(value)
	if err != nil {
		return "[]"
	}
	return string(raw)
}
