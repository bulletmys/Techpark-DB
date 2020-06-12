package delivery

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"strconv"
	"strings"
	"techpark_db/internal/pkg/helpers"
	"techpark_db/internal/pkg/models"
	"techpark_db/internal/pkg/post"
)

type PostHandler struct {
	PostUC post.UseCase
}

func NewPostHandler(uc post.UseCase) *PostHandler {
	return &PostHandler{PostUC: uc}
}

func (uh PostHandler) Create(w http.ResponseWriter, r *http.Request) {
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

	posts := make([]*models.Post, 0)
	if err := json.NewDecoder(r.Body).Decode(&posts); err != nil {
		log.Print(err)
		http.Error(w, "can't decode posts data from body", http.StatusBadRequest)
		return
	}

	err = uh.PostUC.Create(posts, slugOrID, int32(threadID))

	w.Header().Set("Content-Type", "application/json")

	switch err {
	case models.UserNotFound:
		fallthrough
	case models.ThreadNotFound:
		log.Print(err)
		w.WriteHeader(http.StatusNotFound)
		helpers.EncodeAndSend(models.Message{Msg: err.Error()}, w)
		return
	case models.PostNotFound:
		log.Print(err)
		w.WriteHeader(http.StatusConflict)
		helpers.EncodeAndSend(models.Message{Msg: err.Error()}, w)
		return
	case nil:
		w.WriteHeader(http.StatusCreated)
	default:
		log.Print(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(posts); err != nil {
		log.Print(err)
		http.Error(w, "failed to encode posts to json", http.StatusInternalServerError)
		return
	}
}

func (uh PostHandler) Find(w http.ResponseWriter, r *http.Request) {
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

	limit, err := strconv.Atoi(r.URL.Query().Get("limit"))
	if err != nil {
		limit = 0
	}

	since, err := strconv.ParseInt(r.URL.Query().Get("since"), 10, 64)
	if err != nil {
		since = 0
	}

	desc, err := strconv.ParseBool(r.URL.Query().Get("desc"))
	if err != nil {
		desc = false
	}

	sort := r.URL.Query().Get("sort")

	var sortType post.SortType

	switch sort {
	case string(post.FLAT):
		sortType = post.FLAT
	case string(post.TREE):
		sortType = post.TREE
	case string(post.PARENT_TREE):
		sortType = post.PARENT_TREE
	default:
		sortType = post.DEFAULT
	}

	posts, err := uh.PostUC.Find(slugOrID, int32(threadID), int32(limit), since, desc, sortType)

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

	if err := json.NewEncoder(w).Encode(posts); err != nil {
		log.Print(err)
		http.Error(w, "failed to encode posts to json", http.StatusInternalServerError)
		return
	}
}

func (uh PostHandler) GetDetails(w http.ResponseWriter, r *http.Request) {
	id, ok := mux.Vars(r)["id"]
	if !ok {
		log.Print("no id in vars")
		http.Error(w, "no id in vars", http.StatusBadRequest)
		return
	}

	postID, err := strconv.Atoi(id)
	if err != nil {
		log.Print("failed to parse id from vars")
		http.Error(w, "failed to parse id from vars", http.StatusBadRequest)
		return
	}

	userFlag := false
	threadFlag := false
	forumFlag := false

	related, ok := r.URL.Query()["related"]
	if ok {
		related = strings.Split(related[0], ",")
	}

	for _, elem := range related {
		switch elem {
		case "user":
			userFlag = true
		case "thread":
			threadFlag = true
		case "forum":
			forumFlag = true
		default:
			log.Print("falls to default")
			break
		}
	}
	details, err := uh.PostUC.FullPostInfo(int64(postID), userFlag, forumFlag, threadFlag)

	w.Header().Set("Content-Type", "application/json")

	switch err {
	case models.UserNotFound:
		fallthrough
	case models.ThreadNotFound:
		fallthrough
	case models.ForumNotFound:
		fallthrough
	case models.PostNotFound:
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

	if err := json.NewEncoder(w).Encode(details); err != nil {
		log.Print(err)
		http.Error(w, "failed to encode posts to json", http.StatusInternalServerError)
		return
	}
}

func (uh PostHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, ok := mux.Vars(r)["id"]
	if !ok {
		log.Print("no id in vars")
		http.Error(w, "no id in vars", http.StatusBadRequest)
		return
	}

	postID, err := strconv.Atoi(id)
	if err != nil {
		log.Print("failed to parse id from vars")
		http.Error(w, "failed to parse id from vars", http.StatusBadRequest)
		return
	}

	modelPost := models.Post{}
	if err := json.NewDecoder(r.Body).Decode(&modelPost); err != nil {
		log.Print(err)
		http.Error(w, "can't decode posts data from body", http.StatusBadRequest)
		return
	}

	dbPost, err := uh.PostUC.Update(int64(postID), modelPost.Message)

	w.Header().Set("Content-Type", "application/json")

	switch err {
	case models.PostNotFound:
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

	if err := json.NewEncoder(w).Encode(dbPost); err != nil {
		log.Print(err)
		http.Error(w, "failed to encode posts to json", http.StatusInternalServerError)
		return
	}
}
