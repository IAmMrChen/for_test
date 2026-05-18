package srv

import (
	"testing"

	"parent-child-api/model"
)

func TestDashboardSummaryRequiresFamilyId(t *testing.T) {
	mustPanicWith(t, "family id is required", func() {
		DashboardService.Summary(1001, 0)
	})
}

func TestDashboardCanSeeFamilyPendingOnlyParentRole(t *testing.T) {
	parentRoles := []model.FamilyRole{
		model.FamilyRoleOwner,
		model.FamilyRoleAdmin,
		model.FamilyRoleParent,
	}
	for _, role := range parentRoles {
		if !dashboardCanSeeFamilyPending(role) {
			t.Fatalf("role %s should see family pending counters", role)
		}
	}

	if dashboardCanSeeFamilyPending(model.FamilyRoleChild) {
		t.Fatalf("child should not see family pending counters")
	}
}
