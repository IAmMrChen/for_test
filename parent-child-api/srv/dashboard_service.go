package srv

import (
	"fmt"

	"parent-child-api/model"
	"parent-child-api/resx"
)

var DashboardService dashboardService

type dashboardService struct{}

func (x dashboardService) Summary(userId int64, familyId int64) model.DashboardSummary {
	if familyId <= 0 {
		panic(fmt.Errorf("family id is required"))
	}

	member := MemberService.LoadActiveMember(userId, familyId)
	if member == nil {
		panic(fmt.Errorf("permission denied"))
	}

	family := loadDashboardFamily(familyId)
	summary := model.DashboardSummary{
		FamilyId:          family.Id,
		FamilyName:        family.Name,
		MemberId:          member.Id,
		RoleType:          member.RoleType,
		Nickname:          member.Nickname,
		CurrentPoints:     member.CurrentPoints,
		TotalEarnedPoints: member.TotalEarnedPoints,
	}

	summary.ActiveTaskCount = mustScalarCount(
		resx.Db.Main,
		"SELECT COUNT(1) FROM tasks WHERE family_id=@p1 AND status=@p2",
		familyId,
		model.TaskStatusActive,
	)
	summary.MyClaimedTaskCount = mustScalarCount(
		resx.Db.Main,
		"SELECT COUNT(1) FROM task_records WHERE family_id=@p1 AND member_id=@p2 AND status=@p3",
		familyId,
		member.Id,
		model.TaskRecordStatusClaimed,
	)
	summary.MyPendingTaskCount = mustScalarCount(
		resx.Db.Main,
		"SELECT COUNT(1) FROM task_records WHERE family_id=@p1 AND member_id=@p2 AND status=@p3",
		familyId,
		member.Id,
		model.TaskRecordStatusPending,
	)
	summary.ActiveRewardCount = mustScalarCount(
		resx.Db.Main,
		"SELECT COUNT(1) FROM rewards WHERE family_id=@p1 AND status=@p2",
		familyId,
		model.RewardStatusActive,
	)
	summary.MyAppliedRewardCount = mustScalarCount(
		resx.Db.Main,
		"SELECT COUNT(1) FROM reward_records WHERE family_id=@p1 AND member_id=@p2 AND status=@p3",
		familyId,
		member.Id,
		model.RewardRecordStatusApplied,
	)

	if dashboardCanSeeFamilyPending(member.RoleType) {
		summary.FamilyPendingTaskCount = mustScalarCount(
			resx.Db.Main,
			"SELECT COUNT(1) FROM task_records WHERE family_id=@p1 AND status=@p2",
			familyId,
			model.TaskRecordStatusPending,
		)
		summary.FamilyAppliedRewardCount = mustScalarCount(
			resx.Db.Main,
			"SELECT COUNT(1) FROM reward_records WHERE family_id=@p1 AND status=@p2",
			familyId,
			model.RewardRecordStatusApplied,
		)
	}

	return summary
}

func loadDashboardFamily(familyId int64) model.Family {
	family := &model.Family{}
	ok := resx.Db.Main.MustGetStruct(family, "SELECT id, name, creator_id FROM families WHERE id=@p1", familyId)
	if !ok {
		panic(fmt.Errorf("family not found"))
	}
	return *family
}

func dashboardCanSeeFamilyPending(role model.FamilyRole) bool {
	return role.IsParentRole()
}

func mustScalarCount(db scalarIntGetter, query string, args ...any) int {
	value, ok := db.MustScalarInt(query, args...)
	if !ok || value == nil {
		return 0
	}
	return *value
}
