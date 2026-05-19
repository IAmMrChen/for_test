package srv

import (
	"testing"

	"parent-child-api/model"
)

func TestPointLogListRequiresFamilyId(t *testing.T) {
	mustPanicWith(t, "family id is required", func() {
		PointService.ListPointLogs(1001, model.PointLogListRequest{})
	})
}

func TestPointLogMemberScopeForChild(t *testing.T) {
	member := model.FamilyMember{Id: 10, RoleType: model.FamilyRoleChild}

	if got := pointLogScopedMemberId(member, 0); got != 10 {
		t.Fatalf("scoped member id = %d, want 10", got)
	}
	if got := pointLogScopedMemberId(member, 10); got != 10 {
		t.Fatalf("scoped member id = %d, want 10", got)
	}
	mustPanicWith(t, "permission denied", func() {
		pointLogScopedMemberId(member, 11)
	})
}

func TestPointLogMemberScopeForParent(t *testing.T) {
	member := model.FamilyMember{Id: 10, RoleType: model.FamilyRoleParent}

	if got := pointLogScopedMemberId(member, 0); got != 0 {
		t.Fatalf("scoped member id = %d, want 0", got)
	}
	if got := pointLogScopedMemberId(member, 11); got != 11 {
		t.Fatalf("scoped member id = %d, want 11", got)
	}
}
