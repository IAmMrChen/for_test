package srv

import (
	"fmt"
	"os"
	"testing"
	"time"

	"parent-child-api/model"
	"parent-child-api/resx"
)

func TestIntegrationMemberListAndPointLogFlow(t *testing.T) {
	if os.Getenv("PARENT_CHILD_DB_INTEGRATION") != "1" {
		t.Skip("set PARENT_CHILD_DB_INTEGRATION=1 to run database integration flow")
	}

	resx.InitDb(resx.Conf.DB)
	ensureIntegrationSchema(t)

	seed := time.Now().UnixMilli()
	ownerUserId := seed + 400
	childUserId := seed + 401

	family := FamilyService.CreateFamily(ownerUserId, model.FamilyCreateRequest{
		Name:     fmt.Sprintf("point-log-family-%d", seed),
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

	task := TaskService.CreateTask(ownerUserId, model.TaskCreateRequest{
		FamilyId:  familyId,
		Title:     "Read",
		Points:    5,
		CycleType: model.TaskCycleTypeDaily,
	})
	claimed := TaskService.ClaimTask(childUserId, model.TaskClaimRequest{
		FamilyId: familyId,
		TaskId:   task.Id,
	})
	submitted := TaskService.SubmitTask(childUserId, model.TaskSubmitRequest{
		FamilyId: familyId,
		RecordId: claimed.Id,
	})
	TaskService.AuditTask(ownerUserId, model.TaskAuditRequest{
		RecordId: submitted.Id,
		Approved: true,
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

	members := MemberService.ListMembers(ownerUserId, familyId)
	if !memberListContains(members, family.Member.Id, model.FamilyRoleOwner) {
		t.Fatalf("member list does not contain owner: %+v", members)
	}
	if !memberListContains(members, child.Id, model.FamilyRoleChild) {
		t.Fatalf("member list does not contain child: %+v", members)
	}

	parentLogs := PointService.ListPointLogs(ownerUserId, model.PointLogListRequest{FamilyId: familyId})
	if len(parentLogs) != 2 {
		t.Fatalf("parent logs length = %d, want 2: %+v", len(parentLogs), parentLogs)
	}
	if parentLogs[0].SourceType != model.PointSourceTypeReward || parentLogs[0].SourceTitle != reward.Name || parentLogs[0].Points != -2 {
		t.Fatalf("reward log = %+v, want reward title and -2 points", parentLogs[0])
	}
	if parentLogs[1].SourceType != model.PointSourceTypeTask || parentLogs[1].SourceTitle != task.Title || parentLogs[1].Points != 5 {
		t.Fatalf("task log = %+v, want task title and 5 points", parentLogs[1])
	}

	childLogs := PointService.ListPointLogs(childUserId, model.PointLogListRequest{FamilyId: familyId})
	if len(childLogs) != 2 {
		t.Fatalf("child logs length = %d, want 2: %+v", len(childLogs), childLogs)
	}

	mustPanicWith(t, "permission denied", func() {
		PointService.ListPointLogs(childUserId, model.PointLogListRequest{
			FamilyId: familyId,
			MemberId: family.Member.Id,
		})
	})
}

func memberListContains(list []model.FamilyMember, memberId int64, role model.FamilyRole) bool {
	for _, item := range list {
		if item.Id == memberId && item.RoleType == role {
			return true
		}
	}
	return false
}
