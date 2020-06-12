package server

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx"
	"log"
	"net/http"
	forumDelivery "techpark_db/internal/pkg/forum/delivery"
	forumRepo "techpark_db/internal/pkg/forum/repository"
	forumUC "techpark_db/internal/pkg/forum/usecase"
	postDelivery "techpark_db/internal/pkg/post/delivery"
	postRepo "techpark_db/internal/pkg/post/repository"
	postUC "techpark_db/internal/pkg/post/usecase"
	"techpark_db/internal/pkg/service/delivery"
	"techpark_db/internal/pkg/service/repository"
	threadDelivery "techpark_db/internal/pkg/thread/delivery"
	threadRepo "techpark_db/internal/pkg/thread/repository"
	threadUC "techpark_db/internal/pkg/thread/usecase"
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

	//conn, err := pgxpool.Poolect(context.Background(), os.Getenv("DATABASE_URL"))
	conn, err := pgx.NewConnPool(pgx.ConnPoolConfig{
		ConnConfig: pgx.ConnConfig{
			Host:           "localhost",
			User:           "docker",
			Password:       "docker",
			Port:           5432,
			TLSConfig:      nil,
			UseFallbackTLS: false,
			Database:       "docker",
		},
		MaxConnections: 20,

		AfterConnect:   nil,
		AcquireTimeout: 0,
	})
	if err != nil {
		log.Fatalf("failed to connecto to db: %v", err)
	}
	defer conn.Close()

	user := userDelivery.NewUserHandler(
		userUC.UserUC{
			UserRepo: userRepo.DBRepository{Conn: conn}})

	forum := forumDelivery.NewForumHandler(
		forumUC.ForumUC{
			ForumRepo: forumRepo.DBRepository{Conn: conn},
			UserRepo:  userRepo.DBRepository{Conn: conn},
		})

	thread := threadDelivery.NewForumHandler(
		threadUC.ThreadUC{
			ThreadRepo: threadRepo.DBRepository{Conn: conn},
			ForumRepo:  forumRepo.DBRepository{Conn: conn},
			UserRepo:   userRepo.DBRepository{Conn: conn},
		})

	post := postDelivery.NewPostHandler(
		postUC.PostUC{
			UserRepo:   userRepo.DBRepository{Conn: conn},
			PostRepo:   postRepo.DBRepository{Conn: conn},
			ThreadRepo: threadRepo.DBRepository{Conn: conn},
			ForumRepo:  forumRepo.DBRepository{Conn: conn},
		})

	service := delivery.ServiceHandler{ServiceRepo: repository.NewDBServiceRepository(conn)}

	m.HandleFunc("/api/user/{nickname}/create", user.Create).Methods("POST")
	m.HandleFunc("/api/user/{nickname}/profile", user.Find).Methods("GET")
	m.HandleFunc("/api/user/{nickname}/profile", user.Update).Methods("POST")

	m.HandleFunc("/api/forum/create", forum.Create).Methods("POST")
	m.HandleFunc("/api/forum/{slug}/details", forum.Find).Methods("GET")
	m.HandleFunc("/api/forum/{slug}/create", thread.Create).Methods("POST")
	m.HandleFunc("/api/forum/{slug}/threads", thread.GetThreadsByForum).Methods("GET")
	m.HandleFunc("/api/forum/{slug}/users", forum.GetForumUsers).Methods("GET")

	m.HandleFunc("/api/thread/{slug_or_id}/create", post.Create).Methods("POST")
	m.HandleFunc("/api/thread/{slug_or_id}/vote", thread.Vote).Methods("POST")
	m.HandleFunc("/api/thread/{slug_or_id}/details", thread.Get).Methods("GET")
	m.HandleFunc("/api/thread/{slug_or_id}/details", thread.Update).Methods("POST")
	m.HandleFunc("/api/thread/{slug_or_id}/posts", post.Find).Methods("GET")

	m.HandleFunc("/api/post/{id}/details", post.GetDetails).Methods("GET")
	m.HandleFunc("/api/post/{id}/details", post.Update).Methods("POST")

	m.HandleFunc("/api/service/status", service.Status).Methods("GET")
	m.HandleFunc("/api/service/clear", service.Clear).Methods("POST")

	fmt.Println("starting server at :5000")

	if err := server.ListenAndServe(); err != nil {
		log.Fatal("failed to start server")
	}
}
