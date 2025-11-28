package tickets

import (
	"encoding/json"
	"net/http"
	"time"

	"banking-ai-assistant/internal/logger"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	svc Service
	log *logger.StdLogger
}

func NewHandler(s Service, log *logger.StdLogger) *Handler {
	return &Handler{svc: s, log: log}
}

func (h *Handler) ListLetters(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	category := q.Get("category")
	status := q.Get("status")

	res, err := h.svc.ListLetters(r.Context(), category, status)
	if err != nil {
		h.log.Error("ListLetters failed", "err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, res)
}

func (h *Handler) CreateLetter(w http.ResponseWriter, r *http.Request) {
	var req CreateLetterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if err := req.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	letter, err := h.svc.CreateLetter(r.Context(), &req)
	if err != nil {
		h.log.Error("CreateLetter failed", "err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusCreated, letter)
}

func (h *Handler) GetLetter(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	res, err := h.svc.GetLetter(r.Context(), id)
	if err != nil {
		h.log.Error("GetLetter failed", "err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, res)
}

func (h *Handler) GenerateReplies(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req GenerateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if err := req.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	out, err := h.svc.GenerateReplies(r.Context(), id, req.Action)
	if err != nil {
		h.log.Error("GenerateReplies failed", "err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"reply_options": out})
}

func (h *Handler) SendReply(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req SendReplyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if err := req.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	status, err := h.svc.SendReply(r.Context(), id, &req)
	if err != nil {
		h.log.Error("SendReply failed", "err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"status": status, "sent_at": timeNowISO()})
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func timeNowISO() string {
	return timeNow().Format(time.RFC3339)
}

func timeNow() time.Time {
	return time.Now().UTC()
}
