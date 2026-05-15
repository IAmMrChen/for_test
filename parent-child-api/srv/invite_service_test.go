package srv

import (
	"fmt"
	"strings"
	"testing"

	"parent-child-api/model"
	"parent-child-api/resx"
)

func TestNewInviteTokenGeneratesNonEmptyDistinctTokens(t *testing.T) {
	first := newInviteToken()
	second := newInviteToken()

	if first == "" {
		t.Fatal("first token is empty")
	}
	if second == "" {
		t.Fatal("second token is empty")
	}
	if first == second {
		t.Fatal("tokens should be distinct")
	}
}

func TestInviteServiceCreateInvitePanicsForInvalidTargetRoleBeforeDb(t *testing.T) {
	resx.Db = nil

	defer func() {
		v := recover()
		if v == nil {
			t.Fatal("CreateInvite should panic")
		}

		err, ok := v.(error)
		if !ok {
			t.Fatalf("panic = %T(%v), want error", v, v)
		}
		if fmt.Sprint(err) != "invalid target role: OWNER" {
			t.Fatalf("panic error = %v, want invalid target role: OWNER", err)
		}
	}()

	InviteService.CreateInvite(1, model.FamilyInviteCreateRequest{
		FamilyId:   1,
		TargetRole: model.FamilyRoleOwner,
	})
}

func TestInviteMemberInsertArgsUsesInviterUserIdAsCreatedBy(t *testing.T) {
	invite := model.FamilyInvite{
		FamilyId:   9,
		TargetRole: model.FamilyRoleParent,
	}

	args := inviteMemberInsertArgs(invite, 456, 123)

	if args[7] != int64(123) {
		t.Fatalf("created_by arg = %v, want inviter user id 123", args[7])
	}
}

func TestClaimInviteSqlRequiresActiveAndNotExpired(t *testing.T) {
	sql := claimInviteSql()
	normalized := strings.Join(strings.Fields(sql), " ")

	if !strings.Contains(normalized, "WHERE id=@p3 AND status=@p4 AND expires_at>=NOW()") {
		t.Fatalf("claim invite sql = %q, want active status and expiry guard", normalized)
	}
}
