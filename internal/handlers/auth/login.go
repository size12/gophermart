package auth

import (
	"crypto/sha256"
	"encoding/hex"
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

func LoginHandler(s storage.Storage) http.HandlerFunc {
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

		reqUser := models.User{}

		err = json.Unmarshal(resBody, &reqUser)
		if err != nil {
			http.Error(w, "wrong body: "+err.Error(), http.StatusBadRequest)
			return
		}

		user, err := s.GetUser(r.Context(), "login", reqUser.Login)

		if errors.Is(err, storage.ErrNotFound) {
			http.Error(w, "login doesn't exists", http.StatusUnauthorized)
			return
		}

		if err != nil {
			log.Println("Failed get user:", err)
			http.Error(w, "server error", http.StatusInternalServerError)
			return
		}

		h := sha256.New()
		h.Write([]byte(reqUser.Login + reqUser.Password))
		hash := hex.EncodeToString(h.Sum(nil))

		if hash != user.Password {
			http.Error(w, "wrong credentials", http.StatusUnauthorized)
			return
		}

		expiration := time.Now().Add(365 * 24 * time.Hour)
		cookie := http.Cookie{Name: "userCookie", Value: user.Cookie, Expires: expiration, Path: "/"}
		http.SetCookie(w, &cookie)
		w.WriteHeader(http.StatusOK)
	}
}
