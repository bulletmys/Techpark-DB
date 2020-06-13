package delivery

import (
	"encoding/json"
	"github.com/valyala/fasthttp"
	"log"
	"net/http"
	"techpark_db/internal/pkg/models"
	"techpark_db/internal/pkg/user"
)

type UserHandler struct {
	UserUC user.UseCase
}

func NewUserHandler(userUC user.UseCase) UserHandler {
	return UserHandler{UserUC: userUC}
}

func (uh UserHandler) Create(ctx *fasthttp.RequestCtx) {
	nick, ok := ctx.UserValue("nickname").(string)
	if !ok {
		log.Print("no mux vars")
		ctx.Error("no mux vars", http.StatusBadRequest)
		return
	}

	userModel := models.User{Nickname: nick}

	if err := json.Unmarshal(ctx.PostBody(), &userModel); err != nil {
		log.Print(err)
		ctx.Error("can't decode user data from body", http.StatusBadRequest)
		return
	}

	dbUser, err := uh.UserUC.Create(userModel)
	if err != nil {
		log.Print(err)
		ctx.Error("failed to create user", http.StatusInternalServerError)
		return
	}

	ctx.Response.Header.Set("Content-Type", "application/json")

	if dbUser != nil {
		log.Print("user with same data is already exists")
		ctx.SetStatusCode(http.StatusConflict)

		if err := json.NewEncoder(ctx.Response.BodyWriter()).Encode(dbUser); err != nil {
			log.Print(err)
			ctx.Error("failed to encode user to json", http.StatusInternalServerError)
			return
		}
		return
	}

	ctx.SetStatusCode(http.StatusCreated)

	if err := json.NewEncoder(ctx.Response.BodyWriter()).Encode(userModel); err != nil {
		log.Print(err)
		ctx.Error("failed to encode user to json", http.StatusInternalServerError)
		return
	}
}

func (uh UserHandler) Find(ctx *fasthttp.RequestCtx) {
	nick, ok := ctx.UserValue("nickname").(string)
	if !ok {
		log.Print("no mux vars")
		ctx.Error("no nickname field found", http.StatusBadRequest)
		return
	}

	userModel := models.User{Nickname: nick}

	dbUser, err := uh.UserUC.Find(userModel)

	ctx.Response.Header.Set("Content-Type", "application/json")

	if err == models.UserNotFound {
		log.Print("user not found")
		ctx.SetStatusCode(http.StatusNotFound)
		if err := json.NewEncoder(ctx.Response.BodyWriter()).Encode(models.Message{Msg: "can't find user with nickname: " + nick}); err != nil {
			log.Print(err)
			ctx.Error("failed to encode data to json", http.StatusInternalServerError)
		}
		return
	}
	if err != nil {
		log.Print(err)
		ctx.Error("failed to find user", http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(ctx.Response.BodyWriter()).Encode(dbUser); err != nil {
		log.Print(err)
		ctx.Error("failed to encode user to json", http.StatusInternalServerError)
		return
	}
}

func (uh UserHandler) Update(ctx *fasthttp.RequestCtx) {
	nick, ok := ctx.UserValue("nickname").(string)
	if !ok {
		log.Print("no mux vars")
		ctx.Error("no nickname field found", http.StatusBadRequest)
		return
	}

	userModel := models.User{Nickname: nick}
	if err := json.Unmarshal(ctx.PostBody(), &userModel); err != nil {
		log.Print(err)
		ctx.Error("can't decode user data from body", http.StatusBadRequest)
		return
	}

	dbUser, err := uh.UserUC.Update(userModel)
	log.Println("Error:", err)

	ctx.Response.Header.Set("Content-Type", "application/json")

	switch err {
	case models.SameUserExists:
		log.Print(err)
		ctx.SetStatusCode(http.StatusConflict)
		if err := json.NewEncoder(ctx.Response.BodyWriter()).Encode(models.Message{Msg: err.Error()}); err != nil {
			log.Print(err)
			ctx.Error("failed to encode data to json", http.StatusInternalServerError)
		}
		return
	case models.UserNotFound:
		log.Println(err)
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

	ctx.Response.Header.Set("Content-Type", "application/json")

	if err := json.NewEncoder(ctx.Response.BodyWriter()).Encode(dbUser); err != nil {
		log.Print(err)
		ctx.Error("failed to encode user to json", http.StatusInternalServerError)
		return
	}
}
