package middleware

import (
	"net/http"
)

func CheckAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		//проверка

		next.ServeHTTP(w, r)
	})
}
