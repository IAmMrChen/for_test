package ux

import (
	"testing"
	"time"
)

func TestJwtUtilSignAndParseAuthToken(t *testing.T) {
	token, err := JwtUtil.SignAuthToken(AuthTokenClaims{
		UserId:   12,
		OpenId:   "openid-demo",
		Nickname: "大宝爸爸",
	}, "demo-secret", time.Hour)
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}

	claims, err := JwtUtil.ParseAuthToken(token, "demo-secret")
	if err != nil {
		t.Fatalf("parse token: %v", err)
	}

	if claims.UserId != 12 {
		t.Fatalf("UserId = %d, want 12", claims.UserId)
	}
	if claims.OpenId != "openid-demo" {
		t.Fatalf("OpenId = %q, want openid-demo", claims.OpenId)
	}
	if claims.Nickname != "大宝爸爸" {
		t.Fatalf("Nickname = %q, want 大宝爸爸", claims.Nickname)
	}
}

func TestJwtUtilRejectsTamperedToken(t *testing.T) {
	token, err := JwtUtil.SignAuthToken(AuthTokenClaims{
		UserId: 7,
	}, "demo-secret", time.Hour)
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}

	tampered := token[:len(token)-1] + "x"
	if _, err := JwtUtil.ParseAuthToken(tampered, "demo-secret"); err == nil {
		t.Fatal("tampered token should be rejected")
	}
}
