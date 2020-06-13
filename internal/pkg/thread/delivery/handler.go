package delivery

import (
	"encoding/json"
	"github.com/valyala/fasthttp"
	"log"
	"net/http"
	"strconv"
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

func (uh ThreadHandler) Create(ctx *fasthttp.RequestCtx) {
	slug, ok := ctx.UserValue("slug").(string)
	if !ok {
		log.Print("no slug in vars")
		ctx.Error("no slug in vars", http.StatusBadRequest)
		return
	}

	threadModel := models.Thread{}
	if err := json.Unmarshal(ctx.PostBody(), &threadModel); err != nil {
		log.Print(err)
		ctx.Error("can't decode user data from body", http.StatusBadRequest)
		return
	}

	threadModel.Forum = slug

	err := uh.ThreadUC.Create(&threadModel)

	ctx.Response.Header.Set("Content-Type", "application/json")

	switch err {
	case models.SameThreadExists:
		log.Print(err)
		//threadModel.Votes = 0
		ctx.SetStatusCode(http.StatusConflict)
	case models.ForumNotFound:
		fallthrough
	case models.UserNotFound:
		log.Print(err)
		ctx.SetStatusCode(http.StatusNotFound)
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

	if err := json.NewEncoder(ctx.Response.BodyWriter()).Encode(threadModel); err != nil {
		log.Print(err)
		ctx.Error("failed to encode thread to json", http.StatusInternalServerError)
		return
	}
}

func (uh ThreadHandler) GetThreadsByForum(ctx *fasthttp.RequestCtx) {
	slug, ok := ctx.UserValue("slug").(string)
	if !ok {
		log.Print("no slug in vars")
		ctx.Error("no slug in vars", http.StatusBadRequest)
		return
	}
	limit, err := ctx.URI().QueryArgs().GetUint("limit")
	if err != nil {
		limit = 0
	}

	since, err := time.Parse("2006-01-02T15:04:05.000Z", string(ctx.URI().QueryArgs().Peek("since")))

	desc := ctx.URI().QueryArgs().GetBool("desc")

	threads, err := uh.ThreadUC.GetForumsThreads(slug, limit, since, desc)

	ctx.Response.Header.Set("Content-Type", "application/json")

	switch err {
	case models.ForumNotFound:
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

	if err := json.NewEncoder(ctx.Response.BodyWriter()).Encode(threads); err != nil {
		log.Print(err)
		ctx.Error("failed to encode threads to json", http.StatusInternalServerError)
		return
	}
}

func (uh ThreadHandler) Update(ctx *fasthttp.RequestCtx) {
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

	threadModel := models.Thread{}
	if err := json.Unmarshal(ctx.PostBody(), &threadModel); err != nil {
		log.Print(err)
		ctx.Error("can't decode user data from body", http.StatusBadRequest)
		return
	}

	threads, err := uh.ThreadUC.Update(int32(threadID), slugOrID, threadModel.Message, threadModel.Title)

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

	if err := json.NewEncoder(ctx.Response.BodyWriter()).Encode(threads); err != nil {
		log.Print(err)
		ctx.Error("failed to encode threads to json", http.StatusInternalServerError)
		return
	}
}

func (uh ThreadHandler) Vote(ctx *fasthttp.RequestCtx) {
	slugOrID, ok := ctx.UserValue("slug_or_id").(string)
	if !ok {
		log.Print("no slugOrID in vars")
		ctx.Error("no slugOrID in vars", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(slugOrID)
	if err != nil {
		id = -1
	}

	voteModel := models.Vote{}
	if err := json.Unmarshal(ctx.PostBody(), &voteModel); err != nil {
		log.Print(err)
		ctx.Error("can't decode user data from body", http.StatusBadRequest)
		return
	}

	dbThread, err := uh.ThreadUC.Vote(voteModel, slugOrID, int32(id))

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
	case nil:
		break
	default:
		log.Print(err)
		ctx.Error(err.Error(), http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(ctx.Response.BodyWriter()).Encode(dbThread); err != nil {
		log.Print(err)
		ctx.Error("failed to encode thread to json", http.StatusInternalServerError)
		return
	}
}

func (uh ThreadHandler) Get(ctx *fasthttp.RequestCtx) {
	slugOrID, ok := ctx.UserValue("slug_or_id").(string)
	if !ok {
		log.Print("no slugOrID in vars")
		ctx.Error("no slugOrID in vars", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(slugOrID)
	if err != nil {
		id = -1
	}

	dbThread, err := uh.ThreadUC.GetThread(slugOrID, int32(id))

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

	if err := json.NewEncoder(ctx.Response.BodyWriter()).Encode(dbThread); err != nil {
		log.Print(err)
		ctx.Error("failed to encode thread to json", http.StatusInternalServerError)
		return
	}
}
