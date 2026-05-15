package ux

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

var JwtUtil jwtUtil

type jwtUtil struct{}

type AuthTokenClaims struct {
	UserId    int64  `json:"uid"`
	OpenId    string `json:"openid"`
	Nickname  string `json:"nickname"`
	IssuedAt  int64  `json:"iat"`
	ExpiresAt int64  `json:"exp"`
}

func (x jwtUtil) SignAuthToken(claims AuthTokenClaims, secret string, ttl time.Duration) (string, error) {
	if secret == "" {
		return "", errors.New("jwt secret is empty")
	}

	now := time.Now()
	claims.IssuedAt = now.Unix()
	claims.ExpiresAt = now.Add(ttl).Unix()

	header := map[string]string{
		"alg": "HS256",
		"typ": "JWT",
	}

	headerPart, err := encodeJwtPart(header)
	if err != nil {
		return "", err
	}

	claimsPart, err := encodeJwtPart(claims)
	if err != nil {
		return "", err
	}

	signingInput := headerPart + "." + claimsPart
	signature := signJwtInput(signingInput, secret)
	return signingInput + "." + signature, nil
}

func (x jwtUtil) ParseAuthToken(token string, secret string) (*AuthTokenClaims, error) {
	if secret == "" {
		return nil, errors.New("jwt secret is empty")
	}

	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, errors.New("invalid jwt format")
	}

	signingInput := parts[0] + "." + parts[1]
	expectedSignature := signJwtInput(signingInput, secret)
	if !hmac.Equal([]byte(expectedSignature), []byte(parts[2])) {
		return nil, errors.New("invalid jwt signature")
	}

	var claims AuthTokenClaims
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("decode jwt payload: %w", err)
	}
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, fmt.Errorf("decode jwt claims: %w", err)
	}
	if claims.ExpiresAt <= time.Now().Unix() {
		return nil, errors.New("jwt expired")
	}

	return &claims, nil
}

func encodeJwtPart(v any) (string, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(data), nil
}

func signJwtInput(input, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(input))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}
