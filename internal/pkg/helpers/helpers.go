package helpers

import (
	"encoding/json"
	"log"
	"net/http"
)

func EncodeAndSend(data interface{}, w http.ResponseWriter) {
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Print(err)
		http.Error(w, "failed to encode data to json", http.StatusInternalServerError)
	}
}
