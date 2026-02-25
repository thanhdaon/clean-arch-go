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
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func NewHTTPServer(createHandler func(router chi.Router) http.Handler) *http.Server {
	return newHTTPServerOnAddr(":"+os.Getenv("PORT"), createHandler)
}

func newHTTPServerOnAddr(addr string, createHandler func(router chi.Router) http.Handler) *http.Server {
	apiRouter := chi.NewRouter()
	SetMiddlewares(apiRouter)

	rootRouter := chi.NewRouter()
	rootRouter.Mount("/api", createHandler(apiRouter))
	SetSwaggerDoc(rootRouter)

	if addr == ":" {
		addr = ":8080"
	}

	return &http.Server{
		Addr:    addr,
		Handler: rootRouter,
	}
}

func RunHTTPServer(createHandler func(router chi.Router) http.Handler) {
	server := NewHTTPServer(createHandler)

	logrus.Info("Starting HTTP server on ", server.Addr)
	logrus.Info("API Documentation http://localhost", server.Addr, "/doc/index.html")

	if err := server.ListenAndServe(); err != nil {
		logrus.WithError(err).Panic("Unable to start HTTP server")
	}
}

func SetSwaggerDoc(router *chi.Mux) {
	router.Get("/doc/*", httpSwagger.Handler(httpSwagger.URL("/doc.yml")))
	router.Get("/doc.yml", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "ports/openapi.yml")
	})
}

func SetMiddlewares(router *chi.Mux) {
	router.Use(middleware.RealIP)
	router.Use(logs.NewStructuredLogger(logrus.StandardLogger()))
	router.Use(middleware.Recoverer)

	router.Use(otelhttp.NewMiddleware("tasks"))

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
