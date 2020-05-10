package server

import (
	"context"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v4"
	"log"
	"net/http"
	"os"
	forumDelivery "techpark_db/internal/pkg/forum/delivery"
	forumRepo "techpark_db/internal/pkg/forum/repository"
	forumUC "techpark_db/internal/pkg/forum/usecase"
	userDelivery "techpark_db/internal/pkg/user/delivery"
	userRepo "techpark_db/internal/pkg/user/repository"
	userUC "techpark_db/internal/pkg/user/usecase"
	"time"
)

func StartNew() {
	m := mux.NewRouter()

	server := http.Server{
		Addr:         ":5000",
		Handler:      m,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	conn, err := pgx.Connect(context.Background(), "host=localhost port=5432 user=postgres password=postgres dbname=db_forum")
	//conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close(context.Background())

	user := userDelivery.NewUserHandler(
		userUC.UserUC{
			UserRepo: userRepo.DBRepository{Conn: conn}})

	forum := forumDelivery.NewForumHandler(
		forumUC.ForumUC{
			ForumRepo: forumRepo.DBRepository{Conn: conn},
			UserRepo:  userRepo.DBRepository{Conn: conn},
		})

	m.HandleFunc("/user/{nickname}/create", user.Create).Methods("POST")
	m.HandleFunc("/user/{nickname}/profile", user.Find).Methods("GET")
	m.HandleFunc("/user/{nickname}/profile", user.Update).Methods("POST")

	m.HandleFunc("/forum/create", forum.Create).Methods("POST")
	m.HandleFunc("/forum/{slug}/details", forum.Find).Methods("GET")

	fmt.Println("starting server at :5000")

	if err := server.ListenAndServe(); err != nil {
		log.Fatal("failed to start server")
	}
}
