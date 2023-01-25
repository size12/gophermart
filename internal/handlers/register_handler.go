package handlers

import "net/http"

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	//чтение тела
	//запрос к базе

	w.WriteHeader(http.StatusOK)
}