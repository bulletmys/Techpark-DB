package delivery

import (
	"encoding/json"
	"github.com/valyala/fasthttp"
	"log"
	"net/http"
	"techpark_db/internal/pkg/forum"
	"techpark_db/internal/pkg/forum/usecase"
	"techpark_db/internal/pkg/models"
)

type ForumHandler struct {
	ForumUC forum.UseCase
}

func NewForumHandler(uc usecase.ForumUC) *ForumHandler {
	return &ForumHandler{ForumUC: uc}
}

func (uh ForumHandler) Create(ctx *fasthttp.RequestCtx) {
	forumModel := models.Forum{}
	if err := json.Unmarshal(ctx.PostBody(), &forumModel); err != nil {
		log.Print(err)
		ctx.Error("can't decode user data from body", http.StatusBadRequest)
		return
	}

	forum, err := uh.ForumUC.Create(forumModel)

	ctx.Response.Header.Set("Content-Type", "application/json")

	switch err {
	case models.SameForumExists:
		log.Print(err)
		ctx.SetStatusCode(http.StatusConflict)
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

	if err := json.NewEncoder(ctx.Response.BodyWriter()).Encode(forum); err != nil {
		log.Print(err)
		ctx.Error("failed to encode user to json", http.StatusInternalServerError)
		return
	}
}

func (uh ForumHandler) Find(ctx *fasthttp.RequestCtx) {
	slug, ok := ctx.UserValue("slug").(string)
	if !ok {
		log.Print("no mux vars")
		ctx.Error("no nickname field found", http.StatusBadRequest)
		return
	}

	forum, err := uh.ForumUC.Find(slug)

	ctx.Response.Header.Set("Content-Type", "application/json")

	if err == models.ForumNotFound {
		log.Print(err)
		ctx.SetStatusCode(http.StatusNotFound)
		if err := json.NewEncoder(ctx.Response.BodyWriter()).Encode(models.Message{Msg: err.Error()}); err != nil {
			log.Print(err)
			ctx.Error("failed to encode data to json", http.StatusInternalServerError)
		}
		return
	}

	if err := json.NewEncoder(ctx.Response.BodyWriter()).Encode(forum); err != nil {
		log.Print(err)
		ctx.Error("failed to encode user to json", http.StatusInternalServerError)
		return
	}
}

func (uh ForumHandler) GetForumUsers(ctx *fasthttp.RequestCtx) {
	slug, ok := ctx.UserValue("slug").(string)
	if !ok {
		log.Print("no mux vars")
		ctx.Error("no slug field found", http.StatusBadRequest)
		return
	}

	limit, err := ctx.URI().QueryArgs().GetUint("limit")
	if err != nil {
		limit = 0
	}

	since := string(ctx.URI().QueryArgs().Peek("since"))

	desc := ctx.URI().QueryArgs().GetBool("desc")

	dbUsers, err := uh.ForumUC.GetForumUsers(slug, since, limit, desc)

	ctx.Response.Header.Set("Content-Type", "application/json")

	if err == models.ForumNotFound {
		log.Print(err)
		ctx.SetStatusCode(http.StatusNotFound)
		if err := json.NewEncoder(ctx.Response.BodyWriter()).Encode(models.Message{Msg: err.Error()}); err != nil {
			log.Print(err)
			ctx.Error("failed to encode data to json", http.StatusInternalServerError)
		}
		return
	}

	if err := json.NewEncoder(ctx.Response.BodyWriter()).Encode(dbUsers); err != nil {
		log.Print(err)
		ctx.Error("failed to encode user to json", http.StatusInternalServerError)
		return
	}
}
