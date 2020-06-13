package delivery

import (
	"bytes"
	"encoding/json"
	"github.com/valyala/fasthttp"
	"log"
	"net/http"
	"strconv"
	"techpark_db/internal/pkg/models"
	"techpark_db/internal/pkg/post"
	"time"
)

type PostHandler struct {
	PostUC post.UseCase
}

func NewPostHandler(uc post.UseCase) *PostHandler {
	return &PostHandler{PostUC: uc}
}

func (uh PostHandler) Create(ctx *fasthttp.RequestCtx) {
	slugOrID, ok := ctx.UserValue("slug_or_id").(string)
	if !ok {
		log.Print("no slug in vars")
		ctx.Error("no slug in vars", http.StatusBadRequest)
		return
	}

	threadID, err := strconv.Atoi(slugOrID)
	if err != nil {
		threadID = -1
	}

	posts := make([]*models.Post, 0)
	if err := json.Unmarshal(ctx.PostBody(), &posts); err != nil {
		log.Print(err)
		ctx.Error("can't decode posts data from body", http.StatusBadRequest)
		return
	}
	t1 := time.Now().Nanosecond()
	err = uh.PostUC.Create(posts, slugOrID, int32(threadID))
	log.Println("CREATE_TIME_POSTS", time.Now().Nanosecond()-t1)

	ctx.Response.Header.Set("Content-Type", "application/json")

	switch err {
	case models.UserNotFound:
		fallthrough
	case models.ThreadNotFound:
		log.Print(err)
		ctx.SetStatusCode(http.StatusNotFound)
		if err := json.NewEncoder(ctx.Response.BodyWriter()).Encode(models.Message{Msg: err.Error()}); err != nil {
			log.Print(err)
			ctx.Error("failed to encode data to json", http.StatusInternalServerError)
		}
		return
	case models.PostNotFound:
		log.Print(err)
		ctx.SetStatusCode(http.StatusConflict)
		if err := json.NewEncoder(ctx.Response.BodyWriter()).Encode(models.Message{Msg: err.Error()}); err != nil {
			log.Print(err)
			ctx.Error("failed to encode data to json", http.StatusInternalServerError)
		}
		return
	case nil:
		ctx.SetStatusCode(http.StatusCreated)
	default:
		log.Print(err)
		ctx.Error(err.Error(), http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(ctx.Response.BodyWriter()).Encode(posts); err != nil {
		log.Print(err)
		ctx.Error("failed to encode posts to json", http.StatusInternalServerError)
		return
	}
}

func (uh PostHandler) Find(ctx *fasthttp.RequestCtx) {
	slugOrID, ok := ctx.UserValue("slug_or_id").(string)
	if !ok {
		log.Print("no slug in vars")
		ctx.Error("no slug in vars", http.StatusBadRequest)
		return
	}

	threadID, err := strconv.Atoi(slugOrID)
	if err != nil {
		threadID = -1
	}

	limit, err := ctx.URI().QueryArgs().GetUint("limit")
	if err != nil {
		limit = 0
	}

	since, err := strconv.ParseInt(string(ctx.URI().QueryArgs().Peek("since")), 10, 64)
	if err != nil {
		since = 0
	}

	desc := ctx.URI().QueryArgs().GetBool("desc")

	sort := string(ctx.URI().QueryArgs().Peek("sort"))

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

	ctx.Response.Header.Set("Content-Type", "application/json")

	switch err {
	case models.ThreadNotFound:
		log.Print(err)
		ctx.SetStatusCode(http.StatusNotFound)
		if err := json.NewEncoder(ctx.Response.BodyWriter()).Encode(models.Message{Msg: err.Error()}); err != nil {
			log.Print(err)
			ctx.Error("failed to encode data to json", http.StatusInternalServerError)
		}
		return
	case nil:
		break
	default:
		log.Print(err)
		ctx.Error(err.Error(), http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(ctx.Response.BodyWriter()).Encode(posts); err != nil {
		log.Print(err)
		ctx.Error("failed to encode posts to json", http.StatusInternalServerError)
		return
	}
}

func (uh PostHandler) GetDetails(ctx *fasthttp.RequestCtx) {
	id, ok := ctx.UserValue("id").(string)
	if !ok {
		log.Print("no id in vars")
		ctx.Error("no id in vars", http.StatusBadRequest)
		return
	}

	postID, err := strconv.Atoi(id)
	if err != nil {
		log.Print("failed to parse id from vars")
		ctx.Error("failed to parse id from vars", http.StatusBadRequest)
		return
	}

	userFlag := false
	threadFlag := false
	forumFlag := false

	related := ctx.URI().QueryArgs().PeekMulti("related")
	if related != nil {
		related = bytes.Split(related[0], []byte(","))
	}

	for _, elem := range related {
		switch string(elem) {
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

	ctx.Response.Header.Set("Content-Type", "application/json")

	switch err {
	case models.UserNotFound:
		fallthrough
	case models.ThreadNotFound:
		fallthrough
	case models.ForumNotFound:
		fallthrough
	case models.PostNotFound:
		log.Print(err)
		ctx.SetStatusCode(http.StatusNotFound)
		if err := json.NewEncoder(ctx.Response.BodyWriter()).Encode(models.Message{Msg: err.Error()}); err != nil {
			log.Print(err)
			ctx.Error("failed to encode data to json", http.StatusInternalServerError)
		}
		return
	case nil:
		break
	default:
		log.Print(err)
		ctx.Error(err.Error(), http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(ctx.Response.BodyWriter()).Encode(details); err != nil {
		log.Print(err)
		ctx.Error("failed to encode posts to json", http.StatusInternalServerError)
		return
	}
}

func (uh PostHandler) Update(ctx *fasthttp.RequestCtx) {
	id, ok := ctx.UserValue("id").(string)
	if !ok {
		log.Print("no id in vars")
		ctx.Error("no id in vars", http.StatusBadRequest)
		return
	}

	postID, err := strconv.Atoi(id)
	if err != nil {
		log.Print("failed to parse id from vars")
		ctx.Error("failed to parse id from vars", http.StatusBadRequest)
		return
	}

	modelPost := models.Post{}
	if err := json.Unmarshal(ctx.PostBody(), &modelPost); err != nil {
		log.Print(err)
		ctx.Error("can't decode posts data from body", http.StatusBadRequest)
		return
	}

	dbPost, err := uh.PostUC.Update(int64(postID), modelPost.Message)

	ctx.Response.Header.Set("Content-Type", "application/json")

	switch err {
	case models.PostNotFound:
		log.Print(err)
		ctx.SetStatusCode(http.StatusNotFound)
		if err := json.NewEncoder(ctx.Response.BodyWriter()).Encode(models.Message{Msg: err.Error()}); err != nil {
			log.Print(err)
			ctx.Error("failed to encode data to json", http.StatusInternalServerError)
		}
		return
	case nil:
		break
	default:
		log.Print(err)
		ctx.Error(err.Error(), http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(ctx.Response.BodyWriter()).Encode(dbPost); err != nil {
		log.Print(err)
		ctx.Error("failed to encode posts to json", http.StatusInternalServerError)
		return
	}
}
