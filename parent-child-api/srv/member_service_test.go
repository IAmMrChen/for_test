package srv

import (
	"fmt"
	"testing"

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
