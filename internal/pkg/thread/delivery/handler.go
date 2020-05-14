package delivery

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"strconv"
	"techpark_db/internal/pkg/helpers"
	"techpark_db/internal/pkg/models"
	"techpark_db/internal/pkg/thread"
	"time"
)

type ThreadHandler struct {
	ThreadUC thread.UseCase
}

func NewForumHandler(uc thread.UseCase) *ThreadHandler {
	return &ThreadHandler{ThreadUC: uc}
}

func (uh ThreadHandler) Create(w http.ResponseWriter, r *http.Request) {
	slug, ok := mux.Vars(r)["slug"]
	if !ok {
		log.Print("no slug in vars")
		http.Error(w, "no slug in vars", http.StatusBadRequest)
		return
	}

	threadModel := models.Thread{}
	if err := json.NewDecoder(r.Body).Decode(&threadModel); err != nil {
		log.Print(err)
		http.Error(w, "can't decode user data from body", http.StatusBadRequest)
		return
	}

	threadModel.Forum = slug

	err := uh.ThreadUC.Create(&threadModel)

	w.Header().Set("Content-Type", "application/json")

	switch err {
	case models.SameThreadExists:
		log.Print(err)
		//threadModel.Votes = 0
		w.WriteHeader(http.StatusConflict)
	case models.ForumNotFound:
		fallthrough
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

	if err := json.NewEncoder(w).Encode(threadModel); err != nil {
		log.Print(err)
		http.Error(w, "failed to encode thread to json", http.StatusInternalServerError)
		return
	}
}

func (uh ThreadHandler) GetThreadsByForum(w http.ResponseWriter, r *http.Request) {
	slug, ok := mux.Vars(r)["slug"]
	if !ok {
		log.Print("no slug in vars")
		http.Error(w, "no slug in vars", http.StatusBadRequest)
		return
	}
	limit, err := strconv.Atoi(r.URL.Query().Get("limit"))
	if err != nil {
		limit = 0
	}

	since, err := time.Parse("2006-01-02T15:04:05.000Z", r.URL.Query().Get("since"))

	desc, err := strconv.ParseBool(r.URL.Query().Get("desc"))
	if err != nil {
		desc = false
	}

	threads, err := uh.ThreadUC.GetForumsThreads(slug, limit, since, desc)

	w.Header().Set("Content-Type", "application/json")

	switch err {
	case models.ForumNotFound:
		log.Print(err)
		w.WriteHeader(http.StatusNotFound)
		helpers.EncodeAndSend(models.Message{Msg: err.Error()}, w)
		return
	case nil:
		break
	default:
		log.Print(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(threads); err != nil {
		log.Print(err)
		http.Error(w, "failed to encode threads to json", http.StatusInternalServerError)
		return
	}
}

func (uh ThreadHandler) Update(w http.ResponseWriter, r *http.Request) {
	slugOrID, ok := mux.Vars(r)["slug_or_id"]
	if !ok {
		log.Print("no slug in vars")
		http.Error(w, "no slug in vars", http.StatusBadRequest)
		return
	}

	threadID, err := strconv.Atoi(slugOrID)
	if err != nil {
		threadID = -1
	}

	threadModel := models.Thread{}
	if err := json.NewDecoder(r.Body).Decode(&threadModel); err != nil {
		log.Print(err)
		http.Error(w, "can't decode user data from body", http.StatusBadRequest)
		return
	}

	threads, err := uh.ThreadUC.Update(int32(threadID), slugOrID, threadModel.Message, threadModel.Title)

	w.Header().Set("Content-Type", "application/json")

	switch err {
	case models.ThreadNotFound:
		log.Print(err)
		w.WriteHeader(http.StatusNotFound)
		helpers.EncodeAndSend(models.Message{Msg: err.Error()}, w)
		return
	case nil:
		break
	default:
		log.Print(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(threads); err != nil {
		log.Print(err)
		http.Error(w, "failed to encode threads to json", http.StatusInternalServerError)
		return
	}
}

func (uh ThreadHandler) Vote(w http.ResponseWriter, r *http.Request) {
	slugOrID, ok := mux.Vars(r)["slug_or_id"]
	if !ok {
		log.Print("no slugOrID in vars")
		http.Error(w, "no slugOrID in vars", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(slugOrID)
	if err != nil {
		id = -1
	}

	voteModel := models.Vote{}
	if err := json.NewDecoder(r.Body).Decode(&voteModel); err != nil {
		log.Print(err)
		http.Error(w, "can't decode user data from body", http.StatusBadRequest)
		return
	}

	dbThread, err := uh.ThreadUC.Vote(voteModel, slugOrID, int32(id))

	w.Header().Set("Content-Type", "application/json")

	switch err {
	case models.UserNotFound:
		fallthrough
	case models.ThreadNotFound:
		log.Print(err)
		w.WriteHeader(http.StatusNotFound)
		helpers.EncodeAndSend(models.Message{Msg: err.Error()}, w)
		return
	case nil:
		break
	default:
		log.Print(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(dbThread); err != nil {
		log.Print(err)
		http.Error(w, "failed to encode thread to json", http.StatusInternalServerError)
		return
	}
}

func (uh ThreadHandler) Get(w http.ResponseWriter, r *http.Request) {
	slugOrID, ok := mux.Vars(r)["slug_or_id"]
	if !ok {
		log.Print("no slugOrID in vars")
		http.Error(w, "no slugOrID in vars", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(slugOrID)
	if err != nil {
		id = -1
	}

	dbThread, err := uh.ThreadUC.GetThread(slugOrID, int32(id))

	w.Header().Set("Content-Type", "application/json")

	switch err {
	case models.ThreadNotFound:
		log.Print(err)
		w.WriteHeader(http.StatusNotFound)
		helpers.EncodeAndSend(models.Message{Msg: err.Error()}, w)
		return
	case nil:
		break
	default:
		log.Print(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(dbThread); err != nil {
		log.Print(err)
		http.Error(w, "failed to encode thread to json", http.StatusInternalServerError)
		return
	}
}
