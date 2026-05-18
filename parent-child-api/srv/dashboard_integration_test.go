package srv

import (
	"fmt"
	"os"
	"testing"
	"time"

	"parent-child-api/model"
	"parent-child-api/resx"
)

func TestIntegrationDashboardSummaryFlow(t *testing.T) {
	if os.Getenv("PARENT_CHILD_DB_INTEGRATION") != "1" {
		t.Skip("set PARENT_CHILD_DB_INTEGRATION=1 to run database integration flow")
	}

	resx.InitDb(resx.Conf.DB)
	ensureIntegrationSchema(t)

	seed := time.Now().UnixMilli()
	ownerUserId := seed + 300
	childUserId := seed + 301

	family := FamilyService.CreateFamily(ownerUserId, model.FamilyCreateRequest{
		Name:     fmt.Sprintf("dashboard-family-%d", seed),
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

	approvedTask := TaskService.CreateTask(ownerUserId, model.TaskCreateRequest{
		FamilyId:  familyId,
		Title:     "Read book",
		Points:    5,
		CycleType: model.TaskCycleTypeDaily,
	})
	approvedClaim := TaskService.ClaimTask(childUserId, model.TaskClaimRequest{
		FamilyId: familyId,
		TaskId:   approvedTask.Id,
	})
	approvedRecord := TaskService.SubmitTask(childUserId, model.TaskSubmitRequest{
		FamilyId: familyId,
		RecordId: approvedClaim.Id,
	})
	TaskService.AuditTask(ownerUserId, model.TaskAuditRequest{
		RecordId: approvedRecord.Id,
		Approved: true,
	})

	pendingTask := TaskService.CreateTask(ownerUserId, model.TaskCreateRequest{
		FamilyId:  familyId,
		Title:     "Pack bag",
		Points:    2,
		CycleType: model.TaskCycleTypeDaily,
	})
	pendingClaim := TaskService.ClaimTask(childUserId, model.TaskClaimRequest{
		FamilyId: familyId,
		TaskId:   pendingTask.Id,
	})
	TaskService.SubmitTask(childUserId, model.TaskSubmitRequest{
		FamilyId: familyId,
		RecordId: pendingClaim.Id,
	})

	reward := RewardService.CreateReward(ownerUserId, model.RewardCreateRequest{
		FamilyId:   familyId,
		Name:       "Sticker",
		PointsCost: 2,
		Stock:      2,
	})
	RewardService.ApplyReward(childUserId, model.RewardApplyRequest{
		FamilyId: familyId,
		RewardId: reward.Id,
	})

	childSummary := DashboardService.Summary(childUserId, familyId)
	if childSummary.FamilyId != familyId || childSummary.FamilyName != family.Family.Name {
		t.Fatalf("child family = %d/%s, want %d/%s", childSummary.FamilyId, childSummary.FamilyName, familyId, family.Family.Name)
	}
	if childSummary.MemberId != child.Id || childSummary.RoleType != model.FamilyRoleChild {
		t.Fatalf("child member = %d/%s, want %d/%s", childSummary.MemberId, childSummary.RoleType, child.Id, model.FamilyRoleChild)
	}
	if childSummary.CurrentPoints != 3 || childSummary.TotalEarnedPoints != 5 {
		t.Fatalf("child points = current %d total %d, want 3/5", childSummary.CurrentPoints, childSummary.TotalEarnedPoints)
	}
	if childSummary.ActiveTaskCount != 2 || childSummary.MyPendingTaskCount != 1 {
		t.Fatalf("child task counters = active %d pending %d, want 2/1", childSummary.ActiveTaskCount, childSummary.MyPendingTaskCount)
	}
	if childSummary.ActiveRewardCount != 1 || childSummary.MyAppliedRewardCount != 1 {
		t.Fatalf("child reward counters = active %d applied %d, want 1/1", childSummary.ActiveRewardCount, childSummary.MyAppliedRewardCount)
	}
	if childSummary.FamilyPendingTaskCount != 0 || childSummary.FamilyAppliedRewardCount != 0 {
		t.Fatalf("child family counters = tasks %d rewards %d, want 0/0", childSummary.FamilyPendingTaskCount, childSummary.FamilyAppliedRewardCount)
	}

	parentSummary := DashboardService.Summary(ownerUserId, familyId)
	if parentSummary.FamilyPendingTaskCount != 1 {
		t.Fatalf("parent pending tasks = %d, want 1", parentSummary.FamilyPendingTaskCount)
	}
	if parentSummary.FamilyAppliedRewardCount != 1 {
		t.Fatalf("parent applied rewards = %d, want 1", parentSummary.FamilyAppliedRewardCount)
	}
	if parentSummary.MyPendingTaskCount != 0 || parentSummary.MyAppliedRewardCount != 0 {
		t.Fatalf("parent own counters = tasks %d rewards %d, want 0/0", parentSummary.MyPendingTaskCount, parentSummary.MyAppliedRewardCount)
	}
}
