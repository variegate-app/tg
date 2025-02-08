package application

import (
	"context"
	"net/http"
	"telegram/cmd/api/internal/handler"
	"telegram/cmd/api/internal/middleware"
	"telegram/internal/logger"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/rs/cors"
)

type ApplicationConfig interface {
	GetAddress() string
}

type Server struct {
	address string
	log     *logger.Logger
	server  *http.Server
}

func NewServer(l *logger.Logger, address string) *Server {
	corsOptions := cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173"}, // Разрешенные источники запросов
		AllowCredentials: true,                              // Разрешить передачу кредитенциалов (например, куки)
		AllowedMethods:   []string{"HEAD", "GET", "POST", "PUT", "DELETE"},
		AllowedHeaders:   []string{"Origin", "Authorization", "Content-Type", "Accept-Encoding"},
		ExposedHeaders:   []string{"Content-Length", "Access-Control-Allow-Origin", "Access-Control-Allow-Credentials", "Access-Control-Allow-Headers", "Access-Control-Allow-Methods"},
	}
	productsHandler := cors.New(corsOptions).Handler(handler.NewGetProducts())
	router := chi.NewRouter()
	router.Method("GET", "/products", middleware.WithLogging(productsHandler, l))

	return &Server{
		address: address,
		log:     l,
		server: &http.Server{
			ErrorLog:          l.Std(),
			Handler:           router,
			Addr:              address,
			ReadHeaderTimeout: 1 * time.Second,
		},
	}
}

func (s *Server) Run(ctx context.Context) error {
	go func(c context.Context) {
		<-c.Done()
		err := s.server.Shutdown(c)
		if err != nil {
			s.log.ErrorCtx(c, "http server shutting down",
				logger.Field{Key: "error", Value: err.Error()},
			)
		} else {
			s.log.InfoCtx(c, "http server shutdown processed successfully")
		}
	}(ctx)

	s.log.InfoCtx(ctx, "http server running",
		logger.Field{Key: "address", Value: s.address},
	)
	return s.server.ListenAndServe()
}

func (s *Server) HealthCheck() error {
	return nil
}
