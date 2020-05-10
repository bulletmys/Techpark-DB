package delivery

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"techpark_db/internal/pkg/forum/usecase"
	"techpark_db/internal/pkg/models"
)

type ForumHandler struct {
	ForumUC usecase.ForumUC
}

func NewForumHandler(uc usecase.ForumUC) *ForumHandler {
	return &ForumHandler{ForumUC: uc}
}

func (uh ForumHandler) Create(w http.ResponseWriter, r *http.Request) {
	forumModel := models.Forum{}
	if err := json.NewDecoder(r.Body).Decode(&forumModel); err != nil {
		log.Print(err)
		http.Error(w, "can't decode user data from body", http.StatusBadRequest)
		return
	}

	forum, err := uh.ForumUC.Create(forumModel)
	switch err {
	case models.SameForumExists:
		log.Print(err)
		w.WriteHeader(http.StatusConflict)
	case models.UserNotFound:
		log.Print(err)
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	case nil:
		w.WriteHeader(http.StatusCreated)
	default:
		log.Print(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(forum); err != nil {
		log.Print(err)
		http.Error(w, "failed to encode user to json", http.StatusInternalServerError)
		return
	}
}


func (uh ForumHandler) Find(w http.ResponseWriter, r *http.Request) {
	slug, ok := mux.Vars(r)["slug"]
	if !ok {
		log.Print("no mux vars")
		http.Error(w, "no nickname field found", http.StatusBadRequest)
		return
	}

	forum, err := uh.ForumUC.Find(slug)
	if err == models.ForumNotFound {
		log.Print(err)
		http.Error(w, "форум не найден", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(forum); err != nil {
		log.Print(err)
		http.Error(w, "failed to encode user to json", http.StatusInternalServerError)
		return
	}
}
