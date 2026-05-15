package srv

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"parent-child-api/model"
	"parent-child-api/resx"
)

var InviteService inviteService

type inviteService struct{}

func (x inviteService) CreateInvite(operatorUserId int64, req model.FamilyInviteCreateRequest) model.FamilyInvite {
	if !isValidInviteTargetRole(req.TargetRole) {
		panic(fmt.Errorf("invalid target role: %s", req.TargetRole))
	}

	operator := MemberService.RequireParentRole(operatorUserId, req.FamilyId)
	if req.TargetRole == model.FamilyRoleAdmin && operator.RoleType != model.FamilyRoleOwner {
		panic(fmt.Errorf("only owner can invite admin"))
	}

	token := newInviteToken()
	expiresAt := time.Now().Add(7 * 24 * time.Hour)

	tran := resx.Db.Main.MustCreateTransactionEx()
	defer tran.MustClose()

	const insertInviteSql = `
		INSERT INTO family_invites(
			family_id
			, inviter_member_id
			, target_role
			, token
			, status
			, expires_at
		)
		VALUES (@p1, @p2, @p3, @p4, @p5, @p6)
	`
	tran.MustExecute(
		insertInviteSql,
		req.FamilyId,
		operator.Id,
		req.TargetRole,
		token,
		model.FamilyInviteStatusActive,
		expiresAt,
	)

	inviteIdValue, ok := tran.MustScalarInt("SELECT LAST_INSERT_ID()")
	if !ok || inviteIdValue == nil {
		panic(fmt.Errorf("failed to load created family invite id"))
	}
	inviteId := int64(*inviteIdValue)

	tran.MustCommit()

	return model.FamilyInvite{
		Id:              inviteId,
		FamilyId:        req.FamilyId,
		InviterMemberId: operator.Id,
		TargetRole:      req.TargetRole,
		Token:           token,
		Status:          model.FamilyInviteStatusActive,
		ExpiresAt:       expiresAt,
	}
}

func (x inviteService) AcceptInvite(userId int64, token string) model.FamilyMember {
	tran := resx.Db.Main.MustCreateTransactionEx()
	defer tran.MustClose()

	invite := x.loadActiveInviteFrom(tran, token)
	if invite.ExpiresAt.Before(time.Now()) {
		panic("invite expired")
	}

	if loadActiveFamilyMember(tran, userId, invite.FamilyId) != nil {
		panic("user already has a family identity")
	}

	inviterUserId := loadInviteInviterUserId(tran, invite)

	affected := tran.MustExecute(
		claimInviteSql(),
		model.FamilyInviteStatusAccepted,
		userId,
		invite.Id,
		model.FamilyInviteStatusActive,
	)
	if affected != 1 {
		if isInviteExpired(tran, invite.Id) {
			panic("invite expired")
		}
		panic("invite has been changed")
	}

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
		VALUES (@p1, @p2, @p3, @p4, @p5, @p6, @p7, @p8, @p9)
	`
	tran.MustExecute(insertMemberSql, inviteMemberInsertArgs(invite, userId, inviterUserId)...)

	memberIdValue, ok := tran.MustScalarInt("SELECT LAST_INSERT_ID()")
	if !ok || memberIdValue == nil {
		panic(fmt.Errorf("failed to load accepted family member id"))
	}
	memberId := int64(*memberIdValue)

	tran.MustCommit()

	return model.FamilyMember{
		Id:                memberId,
		FamilyId:          invite.FamilyId,
		UserId:            &userId,
		RoleType:          invite.TargetRole,
		Nickname:          "",
		CurrentPoints:     0,
		TotalEarnedPoints: 0,
		IsVirtual:         false,
		Status:            model.FamilyMemberStatusActive,
	}
}

func (x inviteService) loadActiveInvite(token string) model.FamilyInvite {
	return x.loadActiveInviteFrom(resx.Db.Main, token)
}

func (x inviteService) loadActiveInviteFrom(db structGetter, token string) model.FamilyInvite {
	const sql = `
		SELECT id
			, family_id
			, inviter_member_id
			, target_role
			, token
			, status
			, expires_at
			, accepted_by_user_id
			, accepted_at
		FROM family_invites
		WHERE token=@p1 AND status=@p2
	`

	invite := &model.FamilyInvite{}
	ok := db.MustGetStruct(invite, sql, token, model.FamilyInviteStatusActive)
	if !ok {
		panic("invite not found")
	}
	return *invite
}

func claimInviteSql() string {
	return `
		UPDATE family_invites
		SET status=@p1, accepted_by_user_id=@p2, accepted_at=NOW()
		WHERE id=@p3 AND status=@p4 AND expires_at>=NOW()
	`
}

func isValidInviteTargetRole(role model.FamilyRole) bool {
	return role == model.FamilyRoleAdmin || role == model.FamilyRoleParent || role == model.FamilyRoleChild
}

func inviteMemberInsertArgs(invite model.FamilyInvite, userId, inviterUserId int64) []any {
	return []any{
		invite.FamilyId,
		userId,
		invite.TargetRole,
		"",
		0,
		0,
		0,
		inviterUserId,
		model.FamilyMemberStatusActive,
	}
}

type scalarIntGetter interface {
	MustScalarInt(query string, args ...any) (value *int, ok bool)
}

type structGetter interface {
	MustGetStruct(ptr any, query string, args ...any) (ok bool)
}

func loadActiveFamilyMember(db structGetter, userId, familyId int64) *model.FamilyMember {
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
	ok := db.MustGetStruct(member, sql, familyId, userId, model.FamilyMemberStatusActive)
	if !ok {
		return nil
	}
	return member
}

func loadInviteInviterUserId(db scalarIntGetter, invite model.FamilyInvite) int64 {
	const sql = `
		SELECT user_id
		FROM family_members
		WHERE id=@p1 AND family_id=@p2 AND status=@p3
	`

	inviterUserIdValue, ok := db.MustScalarInt(sql, invite.InviterMemberId, invite.FamilyId, model.FamilyMemberStatusActive)
	if !ok || inviterUserIdValue == nil {
		panic(fmt.Errorf("invite inviter not found"))
	}
	return int64(*inviterUserIdValue)
}

func isInviteExpired(db scalarIntGetter, inviteId int64) bool {
	const sql = `
		SELECT IF(expires_at<NOW(), 1, 0)
		FROM family_invites
		WHERE id=@p1
	`

	expiredValue, ok := db.MustScalarInt(sql, inviteId)
	return ok && expiredValue != nil && *expiredValue == 1
}

func newInviteToken() string {
	var data [32]byte
	if _, err := rand.Read(data[:]); err != nil {
		panic(err)
	}
	return base64.RawURLEncoding.EncodeToString(data[:])
}
