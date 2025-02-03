package ports

import (
	"clean-arch-go/common/logs"
	"context"
	"net/http"
	"os"
	"strconv"
	"strings"

	firebase "firebase.google.com/go/v4"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/sirupsen/logrus"
	httpSwagger "github.com/swaggo/http-swagger/v2"
	"google.golang.org/api/option"
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
	if mockAuth, _ := strconv.ParseBool(os.Getenv("MOCK_AUTH")); mockAuth {
		router.Use(HttpMockMiddleware)
		return
	}

	var opts []option.ClientOption
	if file := os.Getenv("SERVICE_ACCOUNT_FILE"); file != "" {
		opts = append(opts, option.WithCredentialsFile(file))
	}

	config := &firebase.Config{ProjectID: os.Getenv("GCP_PROJECT")}
	firebaseApp, err := firebase.NewApp(context.Background(), config, opts...)
	if err != nil {
		logrus.Fatalf("error initializing firebase app: %v\n", err)
	}

	authClient, err := firebaseApp.Auth(context.Background())
	if err != nil {
		logrus.WithError(err).Fatal("Unable to create firebase Auth client")
	}

	router.Use(FirebaseAuthHttpMiddleware{authClient}.Middleware)
}
