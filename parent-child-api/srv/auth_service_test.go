package srv

import (
	"testing"

	"parent-child-api/model"
	"parent-child-api/ux"
)

func TestAuthServiceCreateDemoTokenUsesDefaultUser(t *testing.T) {
	res := AuthService.CreateDemoToken(model.DemoLoginRequest{}, "test-secret")
	if res.User.Id != 1 {
		t.Fatalf("default user id = %d, want 1", res.User.Id)
	}
	if res.Token == "" {
		t.Fatal("token should not be empty")
	}

	claims, err := ux.JwtUtil.ParseAuthToken(res.Token, "test-secret")
	if err != nil {
		t.Fatalf("parse token: %v", err)
	}
	if claims.UserId != res.User.Id {
		t.Fatalf("token user id = %d, want %d", claims.UserId, res.User.Id)
	}
}
