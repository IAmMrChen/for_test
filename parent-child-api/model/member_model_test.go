package model

import "testing"

func TestFamilyRoleIsParentRole(t *testing.T) {
	tests := []struct {
		name string
		role FamilyRole
		want bool
	}{
		{name: "owner", role: FamilyRoleOwner, want: true},
		{name: "admin", role: FamilyRoleAdmin, want: true},
		{name: "parent", role: FamilyRoleParent, want: true},
		{name: "child", role: FamilyRoleChild, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.role.IsParentRole(); got != tt.want {
				t.Fatalf("IsParentRole() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFamilyRoleCanManageMembers(t *testing.T) {
	tests := []struct {
		name string
		role FamilyRole
		want bool
	}{
		{name: "owner", role: FamilyRoleOwner, want: true},
		{name: "admin", role: FamilyRoleAdmin, want: true},
		{name: "parent", role: FamilyRoleParent, want: false},
		{name: "child", role: FamilyRoleChild, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.role.CanManageMembers(); got != tt.want {
				t.Fatalf("CanManageMembers() = %v, want %v", got, tt.want)
			}
		})
	}
}
