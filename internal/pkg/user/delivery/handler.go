package delivery

import (
	"encoding/json"
	"github.com/gorilla/mux"
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

func (uh UserHandler) Create(w http.ResponseWriter, r *http.Request) {
	nick, ok := mux.Vars(r)["nickname"]
	if !ok {
		log.Print("no mux vars")
		http.Error(w, "no nickname field found", http.StatusBadRequest)
		return
	}

	userModel := models.User{Nickname: nick}
	if err := json.NewDecoder(r.Body).Decode(&userModel); err != nil {
		log.Print(err)
		http.Error(w, "can't decode user data from body", http.StatusBadRequest)
		return
	}

	dbUser, err := uh.UserUC.Create(userModel)
	if err != nil {
		log.Print(err)
		http.Error(w, "failed to create user", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if dbUser != nil {
		log.Print("user with same data is already exists")
		w.WriteHeader(http.StatusConflict)
		userModel = *dbUser
	} else {
		w.WriteHeader(http.StatusCreated)
	}

	if err := json.NewEncoder(w).Encode(userModel); err != nil {
		log.Print(err)
		http.Error(w, "failed to encode user to json", http.StatusInternalServerError)
		return
	}
}

func (uh UserHandler) Find(w http.ResponseWriter, r *http.Request) {
	nick, ok := mux.Vars(r)["nickname"]
	if !ok {
		log.Print("no mux vars")
		http.Error(w, "no nickname field found", http.StatusBadRequest)
		return
	}

	userModel := models.User{Nickname: nick}

	dbUser, err := uh.UserUC.Find(userModel)
	if err != nil {
		log.Print(err)
		http.Error(w, "failed to find user", http.StatusInternalServerError)
		return
	}

	if dbUser == nil {
		log.Print("user not found")
		http.Error(w, "can't find user with nickname: "+nick, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(dbUser); err != nil {
		log.Print(err)
		http.Error(w, "failed to encode user to json", http.StatusInternalServerError)
		return
	}
}

func (uh UserHandler) Update(w http.ResponseWriter, r *http.Request) {
	nick, ok := mux.Vars(r)["nickname"]
	if !ok {
		log.Print("no mux vars")
		http.Error(w, "no nickname field found", http.StatusBadRequest)
		return
	}

	userModel := models.User{Nickname: nick}
	if err := json.NewDecoder(r.Body).Decode(&userModel); err != nil {
		log.Print(err)
		http.Error(w, "can't decode user data from body", http.StatusBadRequest)
		return
	}

	dbUser, err := uh.UserUC.Update(userModel)

	switch err {
	case models.SameUserExists:
		log.Print(err)
		http.Error(w, err.Error(), http.StatusConflict)
		return
	case models.UserNotFound:
		log.Println(err)
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	case nil:
		break
	default:
		log.Print(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(dbUser); err != nil {
		log.Print(err)
		http.Error(w, "failed to encode user to json", http.StatusInternalServerError)
		return
	}
}
