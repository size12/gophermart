package orders

import (
	"errors"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/size12/gophermart/internal/models"
	"github.com/size12/gophermart/internal/storage"
)

func OrderHandler(s storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		contentType := r.Header.Get("Content-Type")

		if !strings.Contains(contentType, "text/plain") {
			w.Header().Set("Accept", "text/plain")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		resBody, err := io.ReadAll(r.Body)
		defer r.Body.Close()

		if err != nil {
			http.Error(w, "wrong body: "+err.Error(), http.StatusBadRequest)
			return
		}

		orderNumber, err := strconv.Atoi(string(resBody))
		if err != nil {
			http.Error(w, "wrong order number", http.StatusBadRequest)
			return
		}

		user := r.Context().Value("user").(models.User)

		order := models.Order{
			UserID:    user.ID,
			Number:    orderNumber,
			Status:    "NEW",
			EventTime: time.Now(),
		}

		err = s.AddOrder(r.Context(), order)

		if errors.Is(err, storage.ErrBadOrderNum) {
			http.Error(w, "wrong order number", http.StatusUnprocessableEntity)
			return
		}

		if errors.Is(err, storage.ErrAlreadyLoaded) {
			w.WriteHeader(http.StatusOK)
			return
		}

		if errors.Is(err, storage.ErrLoadedByOtherUser) {
			http.Error(w, "already loaded by another user", http.StatusConflict)
			return
		}

		if err != nil {
			log.Println("Can't add new order:", err)
			http.Error(w, "server error", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusAccepted)
	}
}
