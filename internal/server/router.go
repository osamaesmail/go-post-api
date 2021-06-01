package server

import (
	"net/http"

	"github.com/go-chi/chi"
	chimiddleware "github.com/go-chi/chi/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httprate"
	_ "github.com/osamaesmail/go-post-api/docs"
	"github.com/osamaesmail/go-post-api/internal/app/handler"
	"github.com/osamaesmail/go-post-api/internal/app/repository"
	"github.com/osamaesmail/go-post-api/internal/app/service"
	"github.com/osamaesmail/go-post-api/internal/config"
	"github.com/osamaesmail/go-post-api/internal/db/mysql"
	"github.com/osamaesmail/go-post-api/internal/db/redis"
	"github.com/osamaesmail/go-post-api/internal/security/middleware"
	httpSwagger "github.com/swaggo/http-swagger"
)

// @title Go Post API
// @version 1.0
// @description Implementing back-end services for blog application
// @BasePath /v1

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name X-API-Key
func NewRouter(mysqlClient mysql.Client, redisClient redis.Client) *chi.Mux {
	router := chi.NewRouter()

	router.Use(httprate.LimitByIP(
		config.Cfg().HttpRateLimitRequest,
		config.Cfg().HttpRateLimitTime,
	))
	router.Use(cors.AllowAll().Handler)
	router.Use(chimiddleware.Logger)
	router.Use(chimiddleware.Recoverer)

	accountRepository := repository.NewAccountRepository(mysqlClient, redisClient)
	postRepository := repository.NewPostRepository(mysqlClient, redisClient)
	commentRepository := repository.NewCommentRepository(mysqlClient, redisClient)

	authService := service.NewAuthService(accountRepository)
	accountService := service.NewAccountService(accountRepository)
	postService := service.NewPostService(postRepository)
	commentService := service.NewCommentService(commentRepository)

	authHandler := handler.NewAuthHandler(authService)
	accountHandler := handler.NewAccountHandler(accountService)
	postHandler := handler.NewPostHandler(postService)
	commentHandler := handler.NewCommentHandler(commentService)

	router.Options("/*", func(w http.ResponseWriter, r *http.Request) {})
	api := router.Route("/v1", func(router chi.Router) {})

	api.Route("/accounts", func(r chi.Router) {
		r.Post("/auth", authHandler.Login())

		r.Post("/", accountHandler.Create())
		r.Get("/", accountHandler.List())
		r.Get("/{account_id}", accountHandler.Get())
		r.With(middleware.JWTVerifier).Put("/{account_id}", accountHandler.Update())
		r.With(middleware.JWTVerifier).Put("/{account_id}/password", accountHandler.UpdatePassword())
		r.With(middleware.JWTVerifier).Delete("/{account_id}", accountHandler.Delete())
	})

	api.Route("/posts", func(r chi.Router) {
		r.With(middleware.JWTVerifier).Post("/", postHandler.Create())
		r.Get("/", postHandler.List())
		r.Get("/{post_id}", postHandler.Get())
		r.With(middleware.JWTVerifier).Put("/{post_id}", postHandler.Update())
		r.With(middleware.JWTVerifier).Delete("/{post_id}", postHandler.Delete())
	})

	api.Route("/comments", func(r chi.Router) {
		r.With(middleware.JWTVerifier).Post("/", commentHandler.Create())
		r.Get("/", commentHandler.List())
		r.Get("/{comment_id}", commentHandler.Get())
		r.With(middleware.JWTVerifier).Put("/{comment_id}", commentHandler.Update())
		r.With(middleware.JWTVerifier).Delete("/{comment_id}", commentHandler.Delete())
	})

	api.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("doc.json"),
	))

	return router
}
