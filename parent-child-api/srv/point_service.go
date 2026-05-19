package srv

import (
	"fmt"

	"parent-child-api/model"
	"parent-child-api/resx"
)

var PointService pointService

type pointService struct{}

func (x pointService) ListPointLogs(userId int64, req model.PointLogListRequest) []model.PointLogListItem {
	if req.FamilyId == 0 {
		panic(fmt.Errorf("family id is required"))
	}

	operator := MemberService.LoadActiveMember(userId, req.FamilyId)
	if operator == nil {
		panic(fmt.Errorf("permission denied"))
	}

	memberId := pointLogScopedMemberId(*operator, req.MemberId)
	if memberId != 0 {
		target := MemberService.LoadActiveMemberById(memberId)
		if target == nil || target.FamilyId != req.FamilyId {
			panic(fmt.Errorf("target member not found"))
		}
	}

	sql := `
		SELECT p.id
			, p.family_id
			, p.member_id
			, m.nickname
			, p.points
			, p.source_type
			, p.source_id
			, CASE p.source_type
				WHEN 'TASK' THEN IFNULL(t.title, '')
				WHEN 'REWARD' THEN IFNULL(w.name, '')
				ELSE ''
			END AS source_title
			, p.created_at
		FROM point_logs p
		INNER JOIN family_members m ON m.id = p.member_id
		LEFT JOIN task_records tr ON tr.id = p.source_id AND p.source_type = @p2
		LEFT JOIN tasks t ON t.id = tr.task_id
		LEFT JOIN reward_records rr ON rr.id = p.source_id AND p.source_type = @p3
		LEFT JOIN rewards w ON w.id = rr.reward_id
		WHERE p.family_id=@p1
	`
	args := []any{req.FamilyId, model.PointSourceTypeTask, model.PointSourceTypeReward}

	if memberId != 0 {
		sql += " AND p.member_id=@p4"
		args = append(args, memberId)
	}
	sql += " ORDER BY p.id DESC LIMIT 100"

	return resx.Db.Main.MustListOf(model.PointLogListItem{}, sql, args...).([]model.PointLogListItem)
}

func pointLogScopedMemberId(operator model.FamilyMember, requestedMemberId int64) int64 {
	if operator.RoleType.IsParentRole() {
		return requestedMemberId
	}
	if requestedMemberId == 0 || requestedMemberId == operator.Id {
		return operator.Id
	}
	panic(fmt.Errorf("permission denied"))
}
