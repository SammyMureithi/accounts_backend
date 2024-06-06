package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v4"
)

// UserClaims struct for JWT
type UserClaims struct {
    Email string   `json:"email"`
    Roles []string `json:"roles"`
    jwt.RegisteredClaims
}

// Checks if the user role is in the list of allowed roles
func isAllowedAccess(userRoles []string, allowedRoles []string) bool {
    roleMap := make(map[string]bool)
    for _, role := range userRoles {
        roleMap[role] = true
    }

    for _, allowed := range allowedRoles {
        if roleMap[allowed] {
            return true
        }
    }
    return false
}

// Middleware to enforce role-based access control
func RoleBasedJWTMiddleware(next http.Handler, allowedRoles []string) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        tokenString := r.Header.Get("Authorization")
        if tokenString == "" {
            http.Error(w, "Missing token", http.StatusUnauthorized)
            return
        }

        if strings.HasPrefix(tokenString, "Bearer ") {
            tokenString = strings.TrimPrefix(tokenString, "Bearer ")
        }

        claims := &UserClaims{}
        token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
            return jwtKey, nil
        })

        if err != nil || !token.Valid {
            http.Error(w, "Invalid token", http.StatusUnauthorized)
            return
        }

        if !isAllowedAccess(claims.Roles, allowedRoles) {
            http.Error(w, "Not Allowed", http.StatusForbidden)
            return
        }

        ctx := context.WithValue(r.Context(), "userEmail", claims.Email)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

var jwtKey = []byte("your_secret_key")
