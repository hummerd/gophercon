package http

import (
	"context"
	"net/http"
	"net/http/pprof"
	"os"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"go.uber.org/fx"

	imiddleware "github.com/hummerd/gophercon/internal/api/http/middleware"
	"github.com/hummerd/gophercon/internal/controller"
)

var ()

type Server struct {
	*http.Server

	app controller.App
}

func NewServer(
	lc fx.Lifecycle,
	app controller.App,
) *Server {
	s := &Server{
		Server: &http.Server{
			// Always set time out
			Addr:         ":8091",
			ReadTimeout:  time.Second * 10,
			WriteTimeout: time.Second * 10,
		},
		app: app,
	}

	lc.Append(
		fx.Hook{
			OnStart: func(ctx context.Context) error {
				go func() {
					err := s.ListenAndServe()
					if err != nil && err != http.ErrServerClosed {
						log.Error().Err(err).Msgf("could not serve: %v", err)
					}
				}()
				return nil
			},
			OnStop: func(parentCtx context.Context) error {
				log.Info().Msg("gracefully shutdown http server")
				ctx, cancel := context.WithTimeout(parentCtx, time.Second*10)
				defer cancel()

				return s.Shutdown(ctx)

			},
		},
	)

	return s
}

const (
	MB = 1 << (10 * 2)
)

func Register(
	srv *Server,
) error {

	r := chi.NewRouter()

	logger := zerolog.New(os.Stderr)

	r.Use(imiddleware.Recover(&logger))
	r.Use(imiddleware.RequestID())
	r.Use(middleware.RealIP)
	r.Use(imiddleware.Log(&logger, "/api/v1/auth"))

	r.Route("/api/v1", func(r chi.Router) {
		r.Mount("/debug", middleware.Profiler())
		r.Handle("/debug/pprof/goroutine", pprof.Handler("goroutine"))

		r.Handle("/metrics", promhttp.Handler())

		r.Get("/health", srv.getHealth)
		r.Get("/version", srv.getVersion)

		r.Get("/log/level", srv.getLogLevel)
		r.Put("/log/level", srv.setLogLevel)

		r.Route("/notifications", func(r chi.Router) {
			r.Post("/", count("notifications", srv.createNotification))
		})
	})

	srv.Handler = r

	return nil
}

func count(method string, h http.HandlerFunc) http.HandlerFunc {
	opsProcessed := promauto.NewCounter(prometheus.CounterOpts{
		Name: method,
		Help: method + " counter",
	})
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h(w, r)
		opsProcessed.Inc()
	})
}
