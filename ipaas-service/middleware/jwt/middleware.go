package jwt

import (
	"log/slog"
	"net/http"

	"github.com/wso2/integration-control-plane/ipaas-service/utils"
)

// Middleware returns an HTTP middleware that extracts JWT identity claims from
// the Authorization header and stores them in the request context.
//
// If the Authorization header is absent or malformed, or if any required claim
// is missing, the middleware responds with HTTP 401 and halts the request.
// Downstream handlers retrieve claims with ClaimsFromContext.
func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, err := ExtractJWTClaims(r)
		if err != nil {
			slog.WarnContext(r.Context(), "JWT claim extraction failed",
				"remote_addr", r.RemoteAddr,
				"error", err,
			)
			utils.WriteErrorResponse(w, http.StatusUnauthorized, "authentication failed: "+err.Error())
			return
		}

		next.ServeHTTP(w, r.WithContext(WithClaims(r.Context(), claims)))
	})
}
