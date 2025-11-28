package server

import (
	"github.com/go-chi/chi/v5"

	"banking-ai-assistant/internal/logger"
	"banking-ai-assistant/internal/tickets"
)

func NewRouter(svc tickets.Service, log *logger.StdLogger) *chi.Mux {
	r := chi.NewRouter()

	r.Use(LoggingMiddleware(log))
	r.Use(Recoverer)

	h := tickets.NewHandler(svc, log)

	r.Get("/api/letters", h.ListLetters)
	r.Post("/api/letters", h.CreateLetter)
	r.Get("/api/letters/{id}", h.GetLetter)
	r.Post("/api/letters/{id}/generate", h.GenerateReplies)
	r.Post("/api/letters/{id}/send-reply", h.SendReply)

	return r
}
