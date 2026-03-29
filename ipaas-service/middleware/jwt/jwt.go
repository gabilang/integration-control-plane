package jwt

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

// Claims holds the caller identity extracted from a JWT token.
type Claims struct {
	// ProductType is the product type identifier sourced from the client_id claim.
	ProductType string
	// OrgUUID is the organization UUID sourced from the ouId claim.
	OrgUUID string
	// OrgName is the human-readable organization name sourced from the ouName claim.
	OrgName string
	// OrgHandle is the organization handle (slug) sourced from the ouHandle claim.
	OrgHandle string
}

type claimsContextKey struct{}

// WithClaims returns a copy of ctx that carries the given Claims value.
func WithClaims(ctx context.Context, claims *Claims) context.Context {
	return context.WithValue(ctx, claimsContextKey{}, claims)
}

// ClaimsFromContext retrieves the Claims stored by WithClaims.
// Returns nil if no claims are present in ctx.
func ClaimsFromContext(ctx context.Context) *Claims {
	c, _ := ctx.Value(claimsContextKey{}).(*Claims)
	return c
}

// jwtClaims holds the raw JWT claim fields parsed from the token.
type jwtClaims struct {
	ClientID string `json:"client_id"`
	OuID     string `json:"ouId"`
	OuName   string `json:"ouName"`
	OuHandle string `json:"ouHandle"`
	jwt.RegisteredClaims
}

// ExtractJWTClaims extracts identity claims from the Bearer token in the
// Authorization header. Signature verification is intentionally skipped;
// the API gateway is responsible for token validation.
func ExtractJWTClaims(r *http.Request) (*Claims, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return nil, fmt.Errorf("authorization header not found")
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return nil, fmt.Errorf("invalid authorization header format")
	}

	raw := &jwtClaims{}
	parser := jwt.NewParser()
	if _, _, err := parser.ParseUnverified(parts[1], raw); err != nil {
		return nil, fmt.Errorf("failed to parse JWT: %w", err)
	}

	if raw.ClientID == "" {
		return nil, fmt.Errorf("'client_id' claim not found in JWT")
	}
	if raw.OuID == "" {
		return nil, fmt.Errorf("'ouId' claim not found in JWT")
	}
	if raw.OuName == "" {
		return nil, fmt.Errorf("'ouName' claim not found in JWT")
	}
	if raw.OuHandle == "" {
		return nil, fmt.Errorf("'ouHandle' claim not found in JWT")
	}

	return &Claims{
		ProductType: raw.ClientID,
		OrgUUID:     raw.OuID,
		OrgName:     raw.OuName,
		OrgHandle:   raw.OuHandle,
	}, nil
}
