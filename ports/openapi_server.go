package ports

import (
	"clean-arch-go/core/auth"
	"clean-arch-go/core/logs"
	"net/http"
	"os"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/sirupsen/logrus"
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

func RunHTTPServer(createHandler func(router chi.Router) http.Handler) {
	runHTTPServerOnAddr(":"+os.Getenv("PORT"), createHandler)
}

func runHTTPServerOnAddr(addr string, createHandler func(router chi.Router) http.Handler) {
	apiRouter := chi.NewRouter()
	setMiddlewares(apiRouter)

	rootRouter := chi.NewRouter()
	rootRouter.Mount("/api", createHandler(apiRouter))
	setSwaggerDoc(rootRouter)

	logrus.Info("Starting HTTP server on ", addr)
	logrus.Info("API Documentation ", "http://localhost:8000/doc/index.html")

	if err := http.ListenAndServe(addr, rootRouter); err != nil {
		logrus.WithError(err).Panic("Unable to start HTTP server")
	}
}

func setSwaggerDoc(router *chi.Mux) {
	router.Get("/doc/*", httpSwagger.Handler(httpSwagger.URL("/doc.yml")))
	router.Get("/doc.yml", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "ports/openapi.yml")
	})
}

func setMiddlewares(router *chi.Mux) {
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(logs.NewStructuredLogger(logrus.StandardLogger()))
	router.Use(middleware.Recoverer)

	addCorsMiddleware(router)
	addAuthMiddleware(router)

	router.Use(middleware.SetHeader("X-Content-Type-Options", "nosniff"))
	router.Use(middleware.SetHeader("X-Frame-Options", "deny"))

	router.Use(middleware.NoCache)
}

func addCorsMiddleware(router *chi.Mux) {
	allowedOrigins := strings.Split(os.Getenv("CORS_ALLOWED_ORIGINS"), ";")
	if len(allowedOrigins) == 0 {
		return
	}

	corsMiddleware := cors.New(cors.Options{
		AllowedOrigins:   allowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	})
	router.Use(corsMiddleware.Handler)
}

func addAuthMiddleware(router *chi.Mux) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		logrus.Fatalln("missing JWT SECRET env")
	}
	auth := auth.NewAuth(secret)
	router.Use(newOpenapiAuthMiddleware(auth).Middleware)
}
