package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/size12/gophermart/internal/models"
	"github.com/size12/gophermart/internal/storage"
)

func GetBalanceHandler(s storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		user := r.Context().Value("user").(models.User)

		balance := models.Balance{
			Current:   user.Balance,
			Withdrawn: user.Withdrawn,
		}

		b, err := json.Marshal(balance)

		if err != nil {
			log.Println("Failed marshalling balance:", err)
			http.Error(w, "Server error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(b)
	}
}
