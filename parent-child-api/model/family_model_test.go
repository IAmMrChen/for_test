package model

import (
	"encoding/json"
	"testing"
)

func TestFamilyModelsMarshalJSON(t *testing.T) {
	response := FamilyCreateResponse{
		Family: Family{
			Id:        1,
			Name:      "Home",
			CreatorId: 2,
		},
		Member: FamilyMember{
			Id:       3,
			FamilyId: 1,
			RoleType: FamilyRoleOwner,
			Nickname: "Dad",
		},
	}

	data, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	want := `{"family":{"id":1,"name":"Home","creatorId":2},"member":{"id":3,"familyId":1,"userId":null,"roleType":"OWNER","nickname":"Dad","currentPoints":0,"totalEarnedPoints":0,"isVirtual":false,"status":""}}`
	if string(data) != want {
		t.Fatalf("json.Marshal() = %s, want %s", data, want)
	}
}

func TestFamilyCreateRequestAndListItemMarshalJSON(t *testing.T) {
	createRequest := FamilyCreateRequest{
		Name:     "Home",
		Nickname: "Mom",
	}
	listItem := FamilyListItem{
		FamilyId:   1,
		FamilyName: "Home",
		MemberId:   2,
		RoleType:   FamilyRoleParent,
		Nickname:   "Mom",
	}

	requestData, err := json.Marshal(createRequest)
	if err != nil {
		t.Fatalf("json.Marshal(FamilyCreateRequest) error = %v", err)
	}
	if string(requestData) != `{"name":"Home","nickname":"Mom"}` {
		t.Fatalf("json.Marshal(FamilyCreateRequest) = %s", requestData)
	}

	listItemData, err := json.Marshal(listItem)
	if err != nil {
		t.Fatalf("json.Marshal(FamilyListItem) error = %v", err)
	}
	if string(listItemData) != `{"familyId":1,"familyName":"Home","memberId":2,"roleType":"PARENT","nickname":"Mom"}` {
		t.Fatalf("json.Marshal(FamilyListItem) = %s", listItemData)
	}
}
