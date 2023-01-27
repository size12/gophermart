package withdraw

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/size12/gophermart/internal/models"
	"github.com/size12/gophermart/internal/storage"
)

func WithdrawalHistoryHandler(s storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := r.Context().Value("user").(models.User)

		withdrawals, err := s.WithdrawalHistory(r.Context(), user)
		if err != nil {
			log.Println("Can't get withdrawal history:", err)
			http.Error(w, "server error", http.StatusInternalServerError)
			return
		}

		if len(withdrawals) == 0 {
			http.Error(w, "server error", http.StatusNoContent)
			return
		}

		b, err := json.Marshal(withdrawals)
		if err != nil {
			log.Println("Can't marshal withdrawal history:", err)
			http.Error(w, "server error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(b)
	}
}