package app

import (
	"context"
	"errors"
	"fmt"
	"github.com/bmbl-bumble2/recs-votes-storage/config"
	"github.com/bmbl-bumble2/recs-votes-storage/internal/app/api"
	"github.com/bmbl-bumble2/recs-votes-storage/internal/shared/platform"
	"github.com/danielgtaylor/huma/v2/humacli"
	"net/http"
	"time"
)

type ApiWebServer struct {
	handlerFactory api.HandlerFactory
	config         config.Config
	logger         platform.Logger
}

func NewApiWebServer(
	handlerFactory api.HandlerFactory,
	config config.Config,
	logger platform.Logger,
) *ApiWebServer {
	return &ApiWebServer{
		handlerFactory: handlerFactory,
		config:         config,
		logger:         logger,
	}
}

func (s *ApiWebServer) Serve() {
	cli := humacli.New(func(hooks humacli.Hooks, o *config.ServerOptions) {
		addr := fmt.Sprintf("%s:%d", o.Host, o.Port)

		server := &http.Server{
			Addr:    addr,
			Handler: s.handlerFactory.NewHumaApiServerHandler(),
		}

		hooks.OnStart(func() {
			s.logger.Info(fmt.Sprintf("Listening on http://%s", addr))
			if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				s.logger.Error(fmt.Sprintf("Server error: %s", err))
			}
		})

		hooks.OnStop(func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			_ = server.Shutdown(ctx)
		})
	})

	cli.Run()
}
