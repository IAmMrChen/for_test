package srv

import (
	"fmt"
	"os"
	"testing"
	"time"

	"parent-child-api/model"
	"parent-child-api/resx"
)

func TestIntegrationVirtualChildBindInviteFlow(t *testing.T) {
	if os.Getenv("PARENT_CHILD_DB_INTEGRATION") != "1" {
		t.Skip("set PARENT_CHILD_DB_INTEGRATION=1 to run database integration flow")
	}

	resx.InitDb(resx.Conf.DB)
	ensureIntegrationSchema(t)

	seed := time.Now().UnixMilli()
	ownerUserId := seed + 600
	childUserId := seed + 601
	anotherUserId := seed + 602

	family := FamilyService.CreateFamily(ownerUserId, model.FamilyCreateRequest{
		Name:     fmt.Sprintf("virtual-bind-family-%d", seed),
		Nickname: "owner",
	})
	familyId := family.Family.Id
	t.Cleanup(func() {
		cleanupIntegrationFamily(t, familyId)
	})

	virtualChild := MemberService.CreateVirtualChild(ownerUserId, model.VirtualChildCreateRequest{
		FamilyId: familyId,
		Nickname: "virtual-child",
	})
	resx.Db.Main.MustExecute(`
		UPDATE family_members
		SET current_points=@p1, total_earned_points=@p1
		WHERE id=@p2
	`, 12, virtualChild.Id)

	bindInvite := MemberService.CreateVirtualChildBindInvite(ownerUserId, model.VirtualChildBindInviteCreateRequest{
		FamilyId: familyId,
		MemberId: virtualChild.Id,
	})
	if bindInvite.Token == "" || bindInvite.ChildMemberId != virtualChild.Id {
		t.Fatalf("bind invite = %+v", bindInvite)
	}

	bound := MemberService.AcceptVirtualChildBindInvite(childUserId, bindInvite.Token)
	if bound.Id != virtualChild.Id {
		t.Fatalf("bound member id = %d, want original %d", bound.Id, virtualChild.Id)
	}
	if bound.UserId == nil || *bound.UserId != childUserId {
		t.Fatalf("bound user id = %v, want %d", bound.UserId, childUserId)
	}
	if bound.IsVirtual {
		t.Fatalf("bound child should not be virtual: %+v", bound)
	}
	if bound.CurrentPoints != 12 || bound.TotalEarnedPoints != 12 {
		t.Fatalf("bound points = current %d total %d, want 12/12", bound.CurrentPoints, bound.TotalEarnedPoints)
	}

	secondInvite := InviteService.CreateInvite(ownerUserId, model.FamilyInviteCreateRequest{
		FamilyId:   familyId,
		TargetRole: model.FamilyRoleChild,
	})
	mustPanicWith(t, "user already has a family identity", func() {
		InviteService.AcceptInvite(childUserId, secondInvite.Token)
	})

	existingParentInvite := InviteService.CreateInvite(ownerUserId, model.FamilyInviteCreateRequest{
		FamilyId:   familyId,
		TargetRole: model.FamilyRoleParent,
	})
	InviteService.AcceptInvite(anotherUserId, existingParentInvite.Token)

	secondVirtualChild := MemberService.CreateVirtualChild(ownerUserId, model.VirtualChildCreateRequest{
		FamilyId: familyId,
		Nickname: "second-virtual-child",
	})
	secondBindInvite := MemberService.CreateVirtualChildBindInvite(ownerUserId, model.VirtualChildBindInviteCreateRequest{
		FamilyId: familyId,
		MemberId: secondVirtualChild.Id,
	})
	mustPanicWith(t, "user already has a family identity", func() {
		MemberService.AcceptVirtualChildBindInvite(anotherUserId, secondBindInvite.Token)
	})
}
