package srv

import (
	"fmt"
	"testing"
	"time"

	"parent-child-api/model"
	"parent-child-api/resx"
)

func TestMemberServiceCreateVirtualChildPanicsWhenNicknameIsEmpty(t *testing.T) {
	resx.Db = nil

	defer func() {
		v := recover()
		if v == nil {
			t.Fatal("CreateVirtualChild should panic")
		}

		err, ok := v.(error)
		if !ok {
			t.Fatalf("panic = %T(%v), want error", v, v)
		}
		if fmt.Sprint(err) != "nickname is required" {
			t.Fatalf("panic error = %v, want nickname is required", err)
		}
	}()

	MemberService.CreateVirtualChild(1, model.VirtualChildCreateRequest{
		FamilyId: 1,
		Nickname: " \t\n ",
	})
}

func TestVirtualChildMemberInsertArgsUsesOperatorUserIdAsCreatedBy(t *testing.T) {
	req := model.VirtualChildCreateRequest{
		FamilyId: 9,
		Nickname: "Child",
	}

	args := virtualChildMemberInsertArgs(req, "Child", 123)

	if args[6] != int64(123) {
		t.Fatalf("created_by arg = %v, want operator user id 123", args[6])
	}
}

func TestMemberServiceCreateVirtualChildBindInviteRejectsMissingMember(t *testing.T) {
	resx.Db = nil

	mustPanicWith(t, "member id is required", func() {
		MemberService.CreateVirtualChildBindInvite(1, model.VirtualChildBindInviteCreateRequest{FamilyId: 1})
	})
}

func TestVirtualChildBindInviteInsertArgs(t *testing.T) {
	expiresAt := time.Date(2026, 5, 22, 10, 0, 0, 0, time.UTC)
	args := virtualChildBindInviteInsertArgs(9, 10, 11, "token", expiresAt)

	if args[0] != int64(9) || args[1] != int64(10) || args[2] != int64(11) {
		t.Fatalf("ids args = %+v", args[:3])
	}
	if args[3] != "token" {
		t.Fatalf("token arg = %v, want token", args[3])
	}
	if args[4] != model.FamilyInviteStatusActive {
		t.Fatalf("status arg = %v, want ACTIVE", args[4])
	}
	if args[5] != expiresAt {
		t.Fatalf("expires arg = %v, want %v", args[5], expiresAt)
	}
}

func TestMemberListRequiresFamilyId(t *testing.T) {
	mustPanicWith(t, "family id is required", func() {
		MemberService.ListMembers(1001, 0)
	})
}

func TestMemberRoleHelpers(t *testing.T) {
	parent := model.FamilyMember{RoleType: model.FamilyRoleParent}
	child := model.FamilyMember{RoleType: model.FamilyRoleChild}

	if !memberCanSubmitForChild(parent, true) {
		t.Fatal("parent should submit for virtual child")
	}
	if !memberCanSubmitForChild(parent, false) {
		t.Fatal("parent should submit for real child")
	}
	if memberCanSubmitForChild(child, true) {
		t.Fatal("child should not submit for virtual child through parent helper")
	}
	if memberCanSubmitForChild(child, false) {
		t.Fatal("child should not submit for real child through parent helper")
	}
}
