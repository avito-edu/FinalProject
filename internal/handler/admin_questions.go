package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/ArtemMoroz51/FinalProject/internal/service"
	"github.com/ArtemMoroz51/FinalProject/internal/storage"
	"go.uber.org/zap"
)

type setActiveReq struct {
	IsActive bool `json:"isActive"`
}

func RegisterAdminHandlers(mux *http.ServeMux, admin service.AdminService, adminToken string, log *zap.Logger) {
	if log == nil {
		log = zap.NewNop()
	}

	mux.HandleFunc("/admin/questions", requireAdminToken(adminToken, func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			var in storage.CreateQuestionInput
			if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
				log.Warn("admin create question bad json", zap.Error(err))
				http.Error(w, "bad json", http.StatusBadRequest)
				return
			}

			ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
			defer cancel()

			row, err := admin.CreateQuestion(ctx, in)
			if err != nil {
				log.Warn("admin create question failed", zap.Error(err))
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			log.Info("question created", zap.Int64("id", row.ID), zap.Bool("active", row.IsActive))
			_ = json.NewEncoder(w).Encode(row)

		case http.MethodGet:
			includeInactive := r.URL.Query().Get("all") == "1"

			ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
			defer cancel()

			rows, err := admin.ListQuestions(ctx, includeInactive)
			if err != nil {
				log.Error("admin list questions failed", zap.Error(err))
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			log.Info("questions listed", zap.Bool("include_inactive", includeInactive), zap.Int("count", len(rows)))
			_ = json.NewEncoder(w).Encode(rows)

		default:
			log.Warn("method not allowed", zap.String("path", r.URL.Path), zap.String("method", r.Method))
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}))

	mux.HandleFunc("/admin/questions/", requireAdminToken(adminToken, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			log.Warn("method not allowed", zap.String("path", r.URL.Path), zap.String("method", r.Method))
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		idStr := strings.TrimPrefix(r.URL.Path, "/admin/questions/")
		idStr = strings.TrimSpace(idStr)
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil || id <= 0 {
			log.Warn("admin patch bad id", zap.String("id", idStr))
			http.Error(w, "bad id", http.StatusBadRequest)
			return
		}

		var req setActiveReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Warn("admin patch bad json", zap.Int64("id", id), zap.Error(err))
			http.Error(w, "bad json", http.StatusBadRequest)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		row, err := admin.SetQuestionActive(ctx, id, req.IsActive)
		if err != nil {
			log.Warn("admin set question active failed", zap.Int64("id", id), zap.Bool("active", req.IsActive), zap.Error(err))
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		log.Info("question active updated", zap.Int64("id", id), zap.Bool("active", row.IsActive))
		_ = json.NewEncoder(w).Encode(row)
	}))
}
