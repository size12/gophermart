package middleware

import (
	"context"
	"errors"
	"log"
	"net/http"

	"github.com/size12/gophermart/internal/models"
	"github.com/size12/gophermart/internal/storage"
)

func RequireAuthentication(s storage.Storage) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			path := r.URL.Path
			if path == "/api/user/register" || path == "/api/user/login" {
				next.ServeHTTP(w, r)
				return
			}

			userCookie, err := r.Cookie("userCookie")

			if err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			user, err := s.GetUser(r.Context(), "cookie", userCookie.Value)

			if errors.Is(err, storage.ErrNotFound) {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			if err != nil {
				log.Println("Failed auth checking:", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			ctx := context.WithValue(r.Context(), models.CtxUserKey{}, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
