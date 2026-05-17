package srv

import (
	"fmt"
	"strings"

	"parent-child-api/model"
	"parent-child-api/resx"
)

var MemberService memberService

type memberService struct{}

func (x memberService) LoadActiveMember(userId, familyId int64) *model.FamilyMember {
	const sql = `
		SELECT id
			, family_id
			, user_id
			, role_type
			, nickname
			, current_points
			, total_earned_points
			, is_virtual
			, status
		FROM family_members
		WHERE family_id=@p1 AND user_id=@p2 AND status=@p3
	`

	member := &model.FamilyMember{}
	ok := resx.Db.Main.MustGetStruct(member, sql, familyId, userId, model.FamilyMemberStatusActive)
	if !ok {
		return nil
	}
	return member
}

func (x memberService) RequireParentRole(userId, familyId int64) model.FamilyMember {
	member := x.LoadActiveMember(userId, familyId)
	if member == nil || !member.RoleType.IsParentRole() {
		panic(fmt.Errorf("permission denied"))
	}
	return *member
}

func (x memberService) LoadActiveMemberById(memberId int64) *model.FamilyMember {
	const sql = `
		SELECT id
			, family_id
			, user_id
			, role_type
			, nickname
			, current_points
			, total_earned_points
			, is_virtual
			, status
		FROM family_members
		WHERE id=@p1 AND status=@p2
	`

	member := &model.FamilyMember{}
	ok := resx.Db.Main.MustGetStruct(member, sql, memberId, model.FamilyMemberStatusActive)
	if !ok {
		return nil
	}
	return member
}

func memberCanSubmitForChild(operator model.FamilyMember, targetIsVirtual bool) bool {
	return targetIsVirtual && operator.RoleType.IsParentRole()
}

func virtualChildMemberInsertArgs(req model.VirtualChildCreateRequest, nickname string, operatorUserId int64) []any {
	return []any{
		req.FamilyId,
		model.FamilyRoleChild,
		nickname,
		0,
		0,
		1,
		operatorUserId,
		model.FamilyMemberStatusActive,
	}
}

func (x memberService) CreateVirtualChild(operatorUserId int64, req model.VirtualChildCreateRequest) model.FamilyMember {
	nickname := strings.TrimSpace(req.Nickname)
	if nickname == "" {
		panic(fmt.Errorf("nickname is required"))
	}

	x.RequireParentRole(operatorUserId, req.FamilyId)

	tran := resx.Db.Main.MustCreateTransactionEx()
	defer tran.MustClose()

	const insertMemberSql = `
		INSERT INTO family_members(
			family_id
			, user_id
			, role_type
			, nickname
			, current_points
			, total_earned_points
			, is_virtual
			, created_by
			, status
		)
		VALUES (@p1, NULL, @p2, @p3, @p4, @p5, @p6, @p7, @p8)
	`
	tran.MustExecute(insertMemberSql, virtualChildMemberInsertArgs(req, nickname, operatorUserId)...)

	memberIdValue, ok := tran.MustScalarInt("SELECT LAST_INSERT_ID()")
	if !ok || memberIdValue == nil {
		panic(fmt.Errorf("failed to load created family member id"))
	}
	memberId := int64(*memberIdValue)

	tran.MustCommit()

	return model.FamilyMember{
		Id:                memberId,
		FamilyId:          req.FamilyId,
		UserId:            nil,
		RoleType:          model.FamilyRoleChild,
		Nickname:          nickname,
		CurrentPoints:     0,
		TotalEarnedPoints: 0,
		IsVirtual:         true,
		Status:            model.FamilyMemberStatusActive,
	}
}
