package srv

import (
	"fmt"
	"strings"
	"time"

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

func (x memberService) ListMembers(userId, familyId int64) []model.FamilyMember {
	if familyId == 0 {
		panic(fmt.Errorf("family id is required"))
	}
	if x.LoadActiveMember(userId, familyId) == nil {
		panic(fmt.Errorf("permission denied"))
	}

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
		WHERE family_id=@p1 AND status=@p2
		ORDER BY CASE role_type
			WHEN 'OWNER' THEN 1
			WHEN 'ADMIN' THEN 2
			WHEN 'PARENT' THEN 3
			WHEN 'CHILD' THEN 4
			ELSE 5
		END, id ASC
	`

	return resx.Db.Main.MustListOf(model.FamilyMember{}, sql, familyId, model.FamilyMemberStatusActive).([]model.FamilyMember)
}

func memberCanSubmitForChild(operator model.FamilyMember, targetIsVirtual bool) bool {
	return operator.RoleType.IsParentRole()
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

func virtualChildBindInviteInsertArgs(familyId, childMemberId, inviterMemberId int64, token string, expiresAt time.Time) []any {
	return []any{
		familyId,
		childMemberId,
		inviterMemberId,
		token,
		model.FamilyInviteStatusActive,
		expiresAt,
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

func (x memberService) CreateVirtualChildBindInvite(operatorUserId int64, req model.VirtualChildBindInviteCreateRequest) model.VirtualChildBindInvite {
	if req.FamilyId == 0 {
		panic(fmt.Errorf("family id is required"))
	}
	if req.MemberId == 0 {
		panic(fmt.Errorf("member id is required"))
	}

	operator := x.RequireParentRole(operatorUserId, req.FamilyId)
	child := loadActiveVirtualChildForBind(resx.Db.Main, req.FamilyId, req.MemberId)
	token := newInviteToken()
	expiresAt := time.Now().Add(7 * 24 * time.Hour)

	tran := resx.Db.Main.MustCreateTransactionEx()
	defer tran.MustClose()

	const insertInviteSql = `
		INSERT INTO family_child_bind_invites(
			family_id
			, child_member_id
			, inviter_member_id
			, token
			, status
			, expires_at
		)
		VALUES (@p1, @p2, @p3, @p4, @p5, @p6)
	`
	tran.MustExecute(insertInviteSql, virtualChildBindInviteInsertArgs(req.FamilyId, child.Id, operator.Id, token, expiresAt)...)

	inviteIdValue, ok := tran.MustScalarInt("SELECT LAST_INSERT_ID()")
	if !ok || inviteIdValue == nil {
		panic(fmt.Errorf("failed to load created child bind invite id"))
	}
	inviteId := int64(*inviteIdValue)

	tran.MustCommit()

	return model.VirtualChildBindInvite{
		Id:              inviteId,
		FamilyId:        req.FamilyId,
		ChildMemberId:   child.Id,
		InviterMemberId: operator.Id,
		Token:           token,
		Status:          model.FamilyInviteStatusActive,
		ExpiresAt:       expiresAt,
	}
}

func (x memberService) AcceptVirtualChildBindInvite(userId int64, token string) model.FamilyMember {
	tran := resx.Db.Main.MustCreateTransactionEx()
	defer tran.MustClose()

	invite := loadActiveChildBindInviteFrom(tran, token)
	if invite.ExpiresAt.Before(time.Now()) {
		panic("bind invite expired")
	}
	if loadActiveFamilyMember(tran, userId, invite.FamilyId) != nil {
		panic("user already has a family identity")
	}

	child := loadActiveVirtualChildForBind(tran, invite.FamilyId, invite.ChildMemberId)

	affected := tran.MustExecute(
		claimChildBindInviteSql(),
		model.FamilyInviteStatusAccepted,
		userId,
		invite.Id,
		model.FamilyInviteStatusActive,
	)
	if affected != 1 {
		panic("bind invite has been changed")
	}

	affected = tran.MustExecute(`
		UPDATE family_members
		SET user_id=@p1, is_virtual=0
		WHERE id=@p2
			AND family_id=@p3
			AND role_type=@p4
			AND status=@p5
			AND is_virtual=1
			AND user_id IS NULL
	`, userId, child.Id, invite.FamilyId, model.FamilyRoleChild, model.FamilyMemberStatusActive)
	if affected != 1 {
		panic("virtual child not found")
	}

	tran.MustCommit()

	child.UserId = &userId
	child.IsVirtual = false
	return child
}

func loadActiveVirtualChildForBind(db structGetter, familyId, memberId int64) model.FamilyMember {
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
		WHERE id=@p1
			AND family_id=@p2
			AND role_type=@p3
			AND status=@p4
			AND is_virtual=1
			AND user_id IS NULL
	`

	member := &model.FamilyMember{}
	ok := db.MustGetStruct(member, sql, memberId, familyId, model.FamilyRoleChild, model.FamilyMemberStatusActive)
	if !ok {
		panic("virtual child not found")
	}
	return *member
}

func loadActiveChildBindInviteFrom(db structGetter, token string) model.VirtualChildBindInvite {
	const sql = `
		SELECT id
			, family_id
			, child_member_id
			, inviter_member_id
			, token
			, status
			, expires_at
			, accepted_by_user_id
			, accepted_at
		FROM family_child_bind_invites
		WHERE token=@p1 AND status=@p2
	`

	invite := &model.VirtualChildBindInvite{}
	ok := db.MustGetStruct(invite, sql, token, model.FamilyInviteStatusActive)
	if !ok {
		panic("bind invite not found")
	}
	return *invite
}

func claimChildBindInviteSql() string {
	return `
		UPDATE family_child_bind_invites
		SET status=@p1, accepted_by_user_id=@p2, accepted_at=NOW()
		WHERE id=@p3 AND status=@p4 AND expires_at>=NOW()
	`
}
