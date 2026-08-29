package httpapi

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type ctxKey string

const ctxUserKey ctxKey = "user"

// Claims is the JWT payload issued on login.
type Claims struct {
	UserID int64  `json:"uid"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// requireAPIKey guards the tracker ingest endpoint.
func (s *Server) requireAPIKey(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.Header.Get("X-API-Key")
		if key == "" {
			key = strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		}
		if subtle.ConstantTimeCompare([]byte(key), []byte(s.cfg.IngestAPIKey)) != 1 {
			writeError(w, http.StatusUnauthorized, "invalid api key")
			return
		}
		next.ServeHTTP(w, r)
	})
}

// optionalAuth attaches JWT claims when a valid token is present but does not
// reject anonymous callers — convenient for local development and demos.
// Swap for requireAuth to lock the read APIs down.
func (s *Server) optionalAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if c, ok := s.parseBearer(r); ok {
			r = r.WithContext(context.WithValue(r.Context(), ctxUserKey, c))
		}
		next.ServeHTTP(w, r)
	})
}

//nolint:unused // kept as the production-mode replacement for optionalAuth
func (s *Server) requireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, ok := s.parseBearer(r)
		if !ok {
			writeError(w, http.StatusUnauthorized, "authentication required")
			return
		}
		r = r.WithContext(context.WithValue(r.Context(), ctxUserKey, c))
		next.ServeHTTP(w, r)
	})
}

func (s *Server) parseBearer(r *http.Request) (*Claims, bool) {
	raw := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	if raw == "" || raw == r.Header.Get("Authorization") {
		return nil, false
	}
	claims := &Claims{}
	tok, err := jwt.ParseWithClaims(raw, claims, func(*jwt.Token) (any, error) {
		return []byte(s.cfg.JWTSecret), nil
	}, jwt.WithValidMethods([]string{"HS256"}))
	if err != nil || !tok.Valid {
		return nil, false
	}
	return claims, true
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type loginResponse struct {
	Token string `json:"token"`
	Role  string `json:"role"`
}

// handleLogin is a minimal password login that returns a signed JWT.
// NOTE: the demo accepts any known username with password "password"; wire it to
// users.password_hash (bcrypt) before real use.
func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}
	if req.Password != "password" || req.Username == "" {
		writeError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}
	claims := Claims{
		UserID: 1,
		Role:   "coach",
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   req.Username,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(12 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(s.cfg.JWTSecret))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "sign token")
		return
	}
	writeJSON(w, http.StatusOK, loginResponse{Token: signed, Role: claims.Role})
}
