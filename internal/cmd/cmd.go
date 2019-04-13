package main

import (
	"context"

	"github.com/rs/zerolog/log"
	"go.uber.org/fx"

	httpapi "github.com/hummerd/gophercon/internal/api/http"
	"github.com/hummerd/gophercon/internal/controller"
	"github.com/hummerd/gophercon/internal/dataprovider/pg"
	httpservice "github.com/hummerd/gophercon/internal/service/http"
)

func main() {
	app := fx.New(
		fx.NopLogger,
		fx.Provide(
			httpapi.NewServer,
			pg.NewNotificationStore,
			httpservice.NewSessionStore,
			controller.NewApp,
		),
	)

	ctx, cancel := context.WithTimeout(context.Background(), fx.DefaultTimeout)
	defer cancel()

	if err := app.Start(ctx); err != nil {
		log.Error().Err(err).Msg("Can not start service")
	}

	ctxStop, cancelStop := context.WithTimeout(context.Background(), fx.DefaultTimeout)
	defer cancelStop()
	if err := app.Stop(ctxStop); err != nil {
		return
	}
}
