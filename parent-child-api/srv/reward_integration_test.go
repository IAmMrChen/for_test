package srv

import (
	"fmt"
	"os"
	"testing"
	"time"

	"parent-child-api/model"
	"parent-child-api/resx"
)

func TestIntegrationRewardExchangeFlow(t *testing.T) {
	if os.Getenv("PARENT_CHILD_DB_INTEGRATION") != "1" {
		t.Skip("set PARENT_CHILD_DB_INTEGRATION=1 to run database integration flow")
	}

	resx.InitDb(resx.Conf.DB)
	ensureIntegrationSchema(t)

	seed := time.Now().UnixMilli()
	ownerUserId := seed + 200
	childUserId := seed + 201

	family := FamilyService.CreateFamily(ownerUserId, model.FamilyCreateRequest{
		Name:     fmt.Sprintf("reward-family-%d", seed),
		Nickname: "owner",
	})
	familyId := family.Family.Id
	t.Cleanup(func() {
		cleanupIntegrationFamily(t, familyId)
	})

	invite := InviteService.CreateInvite(ownerUserId, model.FamilyInviteCreateRequest{
		FamilyId:   familyId,
		TargetRole: model.FamilyRoleChild,
	})
	child := InviteService.AcceptInvite(childUserId, invite.Token)

	resx.Db.Main.MustExecute(`
		UPDATE family_members
		SET current_points=@p1, total_earned_points=@p1
		WHERE id=@p2
	`, 10, child.Id)

	reward := RewardService.CreateReward(ownerUserId, model.RewardCreateRequest{
		FamilyId:   familyId,
		Name:       "Ice cream",
		PointsCost: 4,
		Stock:      2,
	})

	record := RewardService.ApplyReward(childUserId, model.RewardApplyRequest{
		FamilyId: familyId,
		RewardId: reward.Id,
	})
	if record.Status != model.RewardRecordStatusApplied {
		t.Fatalf("record status = %s, want APPLIED", record.Status)
	}

	reloaded := MemberService.LoadActiveMember(childUserId, familyId)
	if reloaded == nil || reloaded.CurrentPoints != 6 {
		t.Fatalf("current points = %+v, want 6", reloaded)
	}

	delivered := RewardService.DeliverReward(ownerUserId, model.RewardRecordOperateRequest{RecordId: record.Id})
	if delivered.Status != model.RewardRecordStatusDelivered {
		t.Fatalf("delivered status = %s, want DELIVERED", delivered.Status)
	}

	received := RewardService.ReceiveReward(childUserId, model.RewardRecordOperateRequest{RecordId: record.Id})
	if received.Status != model.RewardRecordStatusReceived {
		t.Fatalf("received status = %s, want RECEIVED", received.Status)
	}

	rejectReward := RewardService.CreateReward(ownerUserId, model.RewardCreateRequest{
		FamilyId:   familyId,
		Name:       "Movie",
		PointsCost: 3,
		Stock:      1,
	})
	rejectRecord := RewardService.ApplyReward(childUserId, model.RewardApplyRequest{
		FamilyId: familyId,
		RewardId: rejectReward.Id,
	})
	rejected := RewardService.RejectReward(ownerUserId, model.RewardRecordOperateRequest{RecordId: rejectRecord.Id})
	if rejected.Status != model.RewardRecordStatusRejected {
		t.Fatalf("rejected status = %s, want REJECTED", rejected.Status)
	}

	reloaded = MemberService.LoadActiveMember(childUserId, familyId)
	if reloaded == nil || reloaded.CurrentPoints != 6 {
		t.Fatalf("current points after reject = %+v, want 6", reloaded)
	}

	negativeCount, ok := resx.Db.Main.MustScalarInt(`
		SELECT COUNT(*)
		FROM point_logs
		WHERE family_id=@p1 AND member_id=@p2 AND points=-4 AND source_type=@p3 AND source_id=@p4
	`, familyId, child.Id, model.PointSourceTypeReward, record.Id)
	if !ok || negativeCount == nil || *negativeCount != 1 {
		t.Fatalf("negative point log count = %v, want 1", negativeCount)
	}

	refundCount, ok := resx.Db.Main.MustScalarInt(`
		SELECT COUNT(*)
		FROM point_logs
		WHERE family_id=@p1 AND member_id=@p2 AND points=3 AND source_type=@p3 AND source_id=@p4
	`, familyId, child.Id, model.PointSourceTypeReward, rejectRecord.Id)
	if !ok || refundCount == nil || *refundCount != 1 {
		t.Fatalf("refund point log count = %v, want 1", refundCount)
	}
}

func TestIntegrationRewardUpdateAndOffShelfFlow(t *testing.T) {
	if os.Getenv("PARENT_CHILD_DB_INTEGRATION") != "1" {
		t.Skip("set PARENT_CHILD_DB_INTEGRATION=1 to run database integration flow")
	}

	resx.InitDb(resx.Conf.DB)
	ensureIntegrationSchema(t)

	seed := time.Now().UnixMilli()
	ownerUserId := seed + 400
	childUserId := seed + 401

	family := FamilyService.CreateFamily(ownerUserId, model.FamilyCreateRequest{
		Name:     fmt.Sprintf("reward-management-family-%d", seed),
		Nickname: "owner",
	})
	familyId := family.Family.Id
	t.Cleanup(func() {
		cleanupIntegrationFamily(t, familyId)
	})

	invite := InviteService.CreateInvite(ownerUserId, model.FamilyInviteCreateRequest{
		FamilyId:   familyId,
		TargetRole: model.FamilyRoleChild,
	})
	InviteService.AcceptInvite(childUserId, invite.Token)

	reward := RewardService.CreateReward(ownerUserId, model.RewardCreateRequest{
		FamilyId:   familyId,
		Name:       "Old reward",
		PointsCost: 5,
		Stock:      2,
	})

	mustPanicWith(t, "permission denied", func() {
		RewardService.UpdateReward(childUserId, model.RewardUpdateRequest{
			FamilyId:   familyId,
			RewardId:   reward.Id,
			Name:       "Child update",
			PointsCost: 6,
			Stock:      3,
		})
	})

	updated := RewardService.UpdateReward(ownerUserId, model.RewardUpdateRequest{
		FamilyId:   familyId,
		RewardId:   reward.Id,
		Name:       "New reward",
		PointsCost: 9,
		Stock:      4,
	})
	if updated.Name != "New reward" || updated.PointsCost != 9 || updated.Stock != 4 {
		t.Fatalf("updated reward = %+v", updated)
	}

	mustPanicWith(t, "permission denied", func() {
		RewardService.OffShelfReward(childUserId, model.RewardOffShelfRequest{
			FamilyId: familyId,
			RewardId: reward.Id,
		})
	})

	offShelf := RewardService.OffShelfReward(ownerUserId, model.RewardOffShelfRequest{
		FamilyId: familyId,
		RewardId: reward.Id,
	})
	if offShelf.Status != model.RewardStatusOffShelf {
		t.Fatalf("off shelf status = %d, want off shelf", offShelf.Status)
	}

	rewards := RewardService.ListRewards(ownerUserId, familyId)
	for _, item := range rewards {
		if item.Id == reward.Id {
			t.Fatalf("off shelf reward should not be listed: %+v", item)
		}
	}

	mustPanicWith(t, "reward not found", func() {
		RewardService.ApplyReward(childUserId, model.RewardApplyRequest{
			FamilyId: familyId,
			RewardId: reward.Id,
		})
	})
}
