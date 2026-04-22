package server

import (
	"context"
	"net/http"
	"strconv"

	"github.com/AnchoredLabs/rwa-backend/apps/api/config"
	"github.com/AnchoredLabs/rwa-backend/libs/errors"
	"github.com/AnchoredLabs/rwa-backend/libs/log"
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type Server struct {
	s *http.Server
}

func NewServer(g *gin.Engine, conf *config.Config, lc fx.Lifecycle) *Server {
	server := &http.Server{
		Addr:    ":" + strconv.Itoa(conf.Server.Port),
		Handler: g,
	}
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				log.InfoZ(ctx, "http server started", zap.Int("port", conf.Server.Port))
				if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
					log.ErrorZ(ctx, "start server failed", zap.Error(err))
					return
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			log.InfoZ(ctx, "shutting down server...")
			if err := server.Shutdown(ctx); err != nil {
				log.ErrorZ(ctx, "server shutdown failed", zap.Error(err))
				return err
			}
			log.InfoZ(ctx, "server exited properly")
			return nil
		},
	})
	return &Server{s: server}
}
