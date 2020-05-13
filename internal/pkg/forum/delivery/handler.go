package delivery

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"strconv"
	"techpark_db/internal/pkg/forum"
	"techpark_db/internal/pkg/forum/usecase"
	"techpark_db/internal/pkg/helpers"
	"techpark_db/internal/pkg/models"
)

type ForumHandler struct {
	ForumUC forum.UseCase
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

	w.Header().Set("Content-Type", "application/json")

	switch err {
	case models.SameForumExists:
		log.Print(err)
		w.WriteHeader(http.StatusConflict)
	case models.UserNotFound:
		log.Print(err)
		w.WriteHeader(http.StatusNotFound)
		helpers.EncodeAndSend(models.Message{Msg: err.Error()}, w)
		return
	case nil:
		w.WriteHeader(http.StatusCreated)
	default:
		log.Print(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

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

	w.Header().Set("Content-Type", "application/json")

	if err == models.ForumNotFound {
		log.Print(err)
		w.WriteHeader(http.StatusNotFound)
		helpers.EncodeAndSend(models.Message{Msg: err.Error()}, w)
		return
	}

	if err := json.NewEncoder(w).Encode(forum); err != nil {
		log.Print(err)
		http.Error(w, "failed to encode user to json", http.StatusInternalServerError)
		return
	}
}


func (uh ForumHandler) GetForumUsers(w http.ResponseWriter, r *http.Request) {
	slug, ok := mux.Vars(r)["slug"]
	if !ok {
		log.Print("no mux vars")
		http.Error(w, "no slug field found", http.StatusBadRequest)
		return
	}

	limit, err := strconv.Atoi(r.URL.Query().Get("limit"))
	if err != nil {
		limit = 0
	}

	since := r.URL.Query().Get("since")

	desc, err := strconv.ParseBool(r.URL.Query().Get("desc"))
	if err != nil {
		desc = false
	}

	dbUsers, err := uh.ForumUC.GetForumUsers(slug, since, limit, desc)

	w.Header().Set("Content-Type", "application/json")

	if err == models.ForumNotFound {
		log.Print(err)
		w.WriteHeader(http.StatusNotFound)
		helpers.EncodeAndSend(models.Message{Msg: err.Error()}, w)
		return
	}

	if err := json.NewEncoder(w).Encode(dbUsers); err != nil {
		log.Print(err)
		http.Error(w, "failed to encode user to json", http.StatusInternalServerError)
		return
	}
}
