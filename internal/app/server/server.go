package server

import (
	"github.com/fasthttp/router"
	"github.com/jackc/pgx"
	"github.com/valyala/fasthttp"
	"log"
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
)

var er int64

func CheckReq(next fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		er++
		log.Println(er, " Req URI:", ctx.Request.URI().String())
		log.Println(er, " Req Body:", string(ctx.Request.Body()))

		next(ctx)
		log.Println(er, " Resp Body:", string(ctx.Response.Body()))
	}
}

func StartNew() {
	r := router.New()

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
			LogLevel:       1,
		},
		MaxConnections: 30,
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

	r.POST("/api/user/{nickname}/create", user.Create)
	//r.GET(, user.Create)
	r.GET("/api/user/{nickname}/profile", user.Find)
	r.POST("/api/user/{nickname}/profile", user.Update)

	r.POST("/api/forum/create", forum.Create)
	r.GET("/api/forum/{slug}/details", forum.Find)
	r.POST("/api/forum/{slug}/create", thread.Create)
	r.GET("/api/forum/{slug}/threads", thread.GetThreadsByForum)
	r.GET("/api/forum/{slug}/users", forum.GetForumUsers)

	r.POST("/api/thread/{slug_or_id}/create", post.Create)
	r.POST("/api/thread/{slug_or_id}/vote", thread.Vote)
	r.GET("/api/thread/{slug_or_id}/details", thread.Get)
	r.POST("/api/thread/{slug_or_id}/details", thread.Update)
	r.GET("/api/thread/{slug_or_id}/posts", post.Find)

	r.GET("/api/post/{id}/details", post.GetDetails)
	r.POST("/api/post/{id}/details", post.Update)

	r.GET("/api/service/status", service.Status)
	r.POST("/api/service/clear", service.Clear)

	log.Println("starting server at :5000")

	if err := fasthttp.ListenAndServe(":5000", CheckReq(r.Handler)); err != nil {
		log.Fatalf("Error in ListenAndServe: %s", err)
	}
}
