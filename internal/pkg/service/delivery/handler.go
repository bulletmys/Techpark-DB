package delivery

import (
	"encoding/json"
	"log"
	"net/http"
	"techpark_db/internal/pkg/service/repository"
)

type ServiceHandler struct {
	ServiceRepo repository.Repository
}

func (uh ServiceHandler) Status(w http.ResponseWriter, r *http.Request) {
	status, err := uh.ServiceRepo.Status()
	if err != nil {
		log.Print(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(status); err != nil {
		log.Print(err)
		http.Error(w, "failed to encode threads to json", http.StatusInternalServerError)
		return
	}
}

func (uh ServiceHandler) Clear(w http.ResponseWriter, r *http.Request) {
	err := uh.ServiceRepo.Clear()
	if err != nil {
		log.Print(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
