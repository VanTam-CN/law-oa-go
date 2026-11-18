package services

import (
	"testing"

	"github.com/stretchr/testify/require"

	"law-oa-go/internal/models"
)

// N3/C11 acceptance (plan 6.5): late decisions carrying a stale round
// snapshot are 409 and never act on the new round; accepted decisions
// anchor their record to the snapshot they decided on; a replayed
// structured resubmit returns the original round, while different content
// cannot revive it.

func TestApprovalDecision_StaleSnapshotIsRejectedWithoutSideEffects(t *testing.T) {
	db := seedResubmitContractWorld(t)
	intakeID := "INT-C11-1"
	seedResubmitContractIntake(t, db, intakeID, "rev-1")
	approval, rec := seedPostCommitConflictWorld2(t, db, intakeID)
	initialSnap := seedActiveSubmissionRound(t, db, approval, intakeID)

	updated, err := NewApprovalService(db).StructuredResubmit("u-lawyer", approval.ID, pinnedResubmitRequest(rec, "rev-1", initialSnap))
	require.NoError(t, err)
	require.Equal(t, models.ApprovalStatusResubmitted, updated.Status)
	require.NotEmpty(t, updated.ActiveSubmissionSnapshotID)
	require.NotNil(t, updated.ActiveSubmissionSnapshotID)
	liveSnapshot := *updated.ActiveSubmissionSnapshotID

	// A stale page decided on an older round: typed 409, no record, the
	// approval stays resubmitted for the current round.
	req := &models.ApprovalDecisionRequest{
		Decision:                     models.ApprovalDecisionApprove,
		DecisionReason:               "迟到的旧页面通过",
		ExpectedSubmissionSnapshotID: "snap-from-old-round",
	}
	_, err = NewApprovalService(db).ProcessApprovalDecision(updated.CurrentApproverID, approval.ID, req)
	require.ErrorIs(t, err, ErrApprovalSnapshotStale)

	var recCount int64
	require.NoError(t, db.Table("approval_records").Where("approval_request_id = ?", approval.ID).Count(&recCount).Error)
	require.EqualValues(t, 1, recCount, "only the seeded return record may exist")

	// The anchored decision on the CURRENT round passes and anchors its
	// record to the snapshot it decided on.
	req.ExpectedSubmissionSnapshotID = liveSnapshot
	decided, err := NewApprovalService(db).ProcessApprovalDecision(updated.CurrentApproverID, approval.ID, req)
	require.NoError(t, err)
	require.Equal(t, models.ApprovalStatusApproved, decided.Status)

	var anchor string
	require.NoError(t, db.Raw("SELECT evidence_references FROM approval_records WHERE approval_request_id = ? AND decision = ?",
		approval.ID, models.ApprovalDecisionApprove).Scan(&anchor).Error)
	require.Contains(t, anchor, liveSnapshot)
}

func TestStructuredResubmit_ReplayReturnsOriginalRound(t *testing.T) {
	db := seedResubmitContractWorld(t)
	intakeID := "INT-C11-2"
	seedResubmitContractIntake(t, db, intakeID, "rev-1")
	approval, rec := seedPostCommitConflictWorld2(t, db, intakeID)
	initialSnap := seedActiveSubmissionRound(t, db, approval, intakeID)
	svc := NewApprovalService(db)

	first, err := svc.StructuredResubmit("u-lawyer", approval.ID, pinnedResubmitRequest(rec, "rev-1", initialSnap))
	require.NoError(t, err)
	originalSnapshot := first.ActiveSubmissionSnapshotID
	require.NotEmpty(t, originalSnapshot)

	// Byte-for-byte replay of the same request returns the original round
	// without a new snapshot or status churn.
	replay, err := svc.StructuredResubmit("u-lawyer", approval.ID, pinnedResubmitRequest(rec, "rev-1", initialSnap))
	require.NoError(t, err)
	require.Equal(t, originalSnapshot, replay.ActiveSubmissionSnapshotID)
	require.Equal(t, models.ApprovalStatusResubmitted, replay.Status)

	var snapCount int64
	require.NoError(t, db.Table("approval_snapshots").
		Where("approval_request_id = ? AND snapshot_type = ?", approval.ID, models.ApprovalSnapshotTypeIntakeSubmission).
		Count(&snapCount).Error)
	require.EqualValues(t, 2, snapCount, "replay must not create a new round")
}

func TestStructuredResubmit_DifferentContentCannotReviveRound(t *testing.T) {
	db := seedResubmitContractWorld(t)
	intakeID := "INT-C11-3"
	seedResubmitContractIntake(t, db, intakeID, "rev-1")
	approval, rec := seedPostCommitConflictWorld2(t, db, intakeID)
	initialSnap := seedActiveSubmissionRound(t, db, approval, intakeID)
	svc := NewApprovalService(db)

	_, err := svc.StructuredResubmit("u-lawyer", approval.ID, pinnedResubmitRequest(rec, "rev-1", initialSnap))
	require.NoError(t, err)

	// Same anchor, different reply content: a different request replayed on
	// the same round is a 409, never a second revival.
	altered := pinnedResubmitRequest(rec, "rev-1", initialSnap)
	altered.Replies[0].Response = "改成不同的回复内容"
	_, err = svc.StructuredResubmit("u-lawyer", approval.ID, altered)
	require.Error(t, err)
	var contractErr *ResubmitContractError
	require.ErrorAs(t, err, &contractErr)
	require.Equal(t, models.ErrCodeApprovalRoundStale, contractErr.Code)

	var snapCount int64
	require.NoError(t, db.Table("approval_snapshots").
		Where("approval_request_id = ? AND snapshot_type = ?", approval.ID, models.ApprovalSnapshotTypeIntakeSubmission).
		Count(&snapCount).Error)
	require.EqualValues(t, 2, snapCount)
}

func TestResubmitRequestDigestNormalizesReplies(t *testing.T) {
	base := &ResubmitApprovalRequest{
		ExpectedIntakeRevision: " rev-1 ",
		ReturnRecordID:         "REC-1",
		RevisionNote:           "说明",
		Replies: []SupplementReply{
			{ItemID: "item-2", Response: "回复二"},
			{ItemID: "item-1", Response: " 回复一 "},
		},
	}
	reordered := &ResubmitApprovalRequest{
		ExpectedIntakeRevision: "rev-1",
		ReturnRecordID:         "REC-1",
		RevisionNote:           "说明",
		Replies: []SupplementReply{
			{ItemID: "item-1", Response: "回复一"},
			{ItemID: "item-2", Response: "回复二"},
		},
	}
	d1, err := resubmitRequestDigest(base)
	require.NoError(t, err)
	d2, err := resubmitRequestDigest(reordered)
	require.NoError(t, err)
	require.Equal(t, d1, d2, "reply order and surrounding whitespace must not change the digest")

	different := &ResubmitApprovalRequest{
		ExpectedIntakeRevision: "rev-1",
		ReturnRecordID:         "REC-1",
		RevisionNote:           "说明",
		Replies: []SupplementReply{
			{ItemID: "item-1", Response: "回复一"},
			{ItemID: "item-2", Response: "回复二"},
			{ItemID: "item-3", Response: "新增回复"},
		},
	}
	d3, err := resubmitRequestDigest(different)
	require.NoError(t, err)
	require.NotEqual(t, d1, d3)
}
