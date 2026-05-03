package ws

import (
	"context"
	"fmt"
	"net/http"

	"github.com/AnchoredLabs/rwa-backend/apps/ws-server/config"
	"github.com/AnchoredLabs/rwa-backend/libs/log"
	"github.com/olahol/melody"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type Server struct {
	conf *config.Config
	m    *melody.Melody
}

func NewServer(conf *config.Config, lc fx.Lifecycle) *Server {
	s := &Server{
		conf: conf,
		m:    melody.New(),
	}
	lc.Append(fx.Hook{
		OnStart: s.start,
		OnStop:  s.Stop,
	})
	return s
}

// GetMelody get melody instance
func (s *Server) GetMelody() *melody.Melody {
	return s.m
}

func (s *Server) start(ctx context.Context) error {
	basePath := s.conf.Server.BasePath
	port := s.conf.Server.Port
	ctx = context.WithValue(context.Background(), log.TraceID, "ws_server")
	go func() {
		s.bindHttp(ctx, basePath)
		log.InfoZ(ctx, "start ws server", zap.String("addr", fmt.Sprintf("ws://127.0.0.1:%d%s", port, basePath)))
		err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
		if err != nil {
			log.ErrorZ(ctx, "ws listen failed", zap.Error(err))
			panic("ws listen failed")
		}
		log.InfoZ(ctx, "ws listen success", zap.Int("port", port), zap.String("path", basePath))
	}()
	return nil
}

// bindHttp bind http handler
func (s *Server) bindHttp(ctx context.Context, basePath string) {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("ok"))
	})
	http.HandleFunc(basePath, func(w http.ResponseWriter, r *http.Request) {
		err := s.m.HandleRequest(w, r)
		if err != nil {
			log.ErrorZ(ctx, "ws handle request failed", zap.Any("error", err.Error()))
			return
		}
	})
}

func (s *Server) Stop(ctx context.Context) error {
	err := s.m.Close()
	if err != nil {
		log.ErrorZ(ctx, "ws server stop failed", zap.Error(err))
		return err
	}
	log.InfoZ(ctx, "ws server stop success")
	return nil
}
