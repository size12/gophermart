package auth

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/size12/gophermart/internal/models"
	"github.com/size12/gophermart/internal/storage"
)

func RegisterHandler(s storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		contentType := r.Header.Get("Content-Type")

		if !strings.Contains(contentType, "application/json") {
			w.Header().Set("Accept", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		resBody, err := io.ReadAll(r.Body)
		defer r.Body.Close()

		if err != nil {
			http.Error(w, "wrong body: "+err.Error(), http.StatusBadRequest)
			return
		}

		user := models.User{}

		err = json.Unmarshal(resBody, &user)
		if err != nil {
			http.Error(w, "wrong body: "+err.Error(), http.StatusBadRequest)
			return
		}

		userCookie, err := s.AddUser(r.Context(), user)

		if errors.Is(err, storage.ErrLoginExists) {
			http.Error(w, "login already exists", http.StatusConflict)
			return
		}

		if err != nil {
			log.Println("Failed add user:", err)
			http.Error(w, "server error", http.StatusInternalServerError)
			return
		}

		expiration := time.Now().Add(365 * 24 * time.Hour)
		cookie := http.Cookie{Name: "userCookie", Value: userCookie, Expires: expiration, Path: "/"}
		http.SetCookie(w, &cookie)
		w.WriteHeader(http.StatusOK)
	}
}
