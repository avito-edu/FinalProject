package app

import (
	"context"
	"net/http"
	"time"

	"github.com/ArtemMoroz51/FinalProject/internal/game"
	"github.com/ArtemMoroz51/FinalProject/internal/handler"
	"github.com/ArtemMoroz51/FinalProject/internal/logger"
	"github.com/ArtemMoroz51/FinalProject/internal/service"
	"github.com/ArtemMoroz51/FinalProject/internal/storage"
	"github.com/ArtemMoroz51/FinalProject/internal/ws"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type App struct {
	cfg Config
	log *zap.Logger
	db  *pgxpool.Pool
	srv *http.Server
}

func New(cfg Config) (*App, error) {
	l, err := logger.New(logger.Config{Level: cfg.LogLevel, File: cfg.LogFile})
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	db, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		_ = l.Sync()
		return nil, err
	}

	qs := storage.NewPostgresQuestionStore(db)
	rm := game.NewRoomManager()

	gameSvc := service.NewGameService(rm, qs, service.Config{
		AnsweringSeconds: cfg.AnsweringSeconds,
		ResultsPause:     cfg.ResultsPause,
		MaxRounds:        cfg.MaxRounds,
	})
	adminSvc := service.NewAdminService(qs)

	hub := ws.NewHub(gameSvc, l)

	mux := http.NewServeMux()
	handler.RegisterHandlers(mux, gameSvc, hub, l)
	handler.RegisterAdminHandlers(mux, adminSvc, cfg.AdminToken, l)

	srv := &http.Server{
		Addr:    cfg.HTTPAddr,
		Handler: mux,
	}

	return &App{cfg: cfg, log: l, db: db, srv: srv}, nil
}

func (a *App) Run() error {
	a.log.Info("server started",
		zap.String("addr", a.cfg.HTTPAddr),
		zap.String("log_level", a.cfg.LogLevel),
		zap.String("log_file", a.cfg.LogFile),
	)
	return a.srv.ListenAndServe()
}

func (a *App) Close() {
	if a.db != nil {
		a.db.Close()
	}
	if a.log != nil {
		_ = a.log.Sync()
	}
}
