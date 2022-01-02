package auth

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/shafiquejamal/reactjs-golang-starter/errorhandler"
)

func RequireAuthentication(
	authorizationStrategy func(user User, r *http.Request) error,
	userIdentityFetcher func(bearerToken string, w *http.ResponseWriter) ([]byte, error),
	userDataFetcher func(userIdentity *UserIdentity, user *User, w *http.ResponseWriter) error) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := strings.Split(r.Header.Get("Authorization"), "Bearer")
			if len(authHeader) != 2 {
				errorhandler.ReturnError(&w, http.StatusBadRequest, "Malformed authorization header or token", errors.New("Malformed authorization header or token"))
				return
			} else {
				bearerToken := strings.TrimSpace(authHeader[1])
				body, err := userIdentityFetcher(bearerToken, &w)
				userIdentity := UserIdentity{}
				if err := json.Unmarshal(body, &userIdentity); err != nil {
					errorhandler.ReturnError(&w, http.StatusInternalServerError, "UserID error", err)
					return
				}
				user := User{}
				if err := userDataFetcher(&userIdentity, &user, &w); err != nil {
					errorhandler.ReturnError(&w, http.StatusInternalServerError, "UserID error", err)
					return
				}
				if !user.Identity.EmailVerified {
					errorhandler.ReturnError(&w, http.StatusUnauthorized, "Unauthorized - email not verified", err)
					return
				}
				if err := authorizationStrategy(user, r); err != nil {
					errorhandler.ReturnError(&w, http.StatusUnauthorized, "Unauthorized - denied by policy", err)
					return
				}
				ctx := context.WithValue(r.Context(), "User", user)
				next.ServeHTTP(w, r.WithContext(ctx))
			}
		})
	}
}
