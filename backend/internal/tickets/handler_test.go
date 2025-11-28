package tickets

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"banking-ai-assistant/internal/logger"

	"github.com/go-chi/chi/v5"
)

type dummyService struct {
	letters   map[string]*Letter
	replies   map[string][]GeneratedReply
	createErr error
	listErr   error
	getErr    error
	genErr    error
	sendErr   error
}

func newDummyService() *dummyService {
	return &dummyService{
		letters: make(map[string]*Letter),
		replies: make(map[string][]GeneratedReply),
	}
}

func (s *dummyService) ListLetters(ctx context.Context, category, status string) ([]Letter, error) {
	if s.listErr != nil {
		return nil, s.listErr
	}
	out := []Letter{}
	for _, l := range s.letters {
		out = append(out, *l)
	}
	return out, nil
}

func (s *dummyService) CreateLetter(ctx context.Context, req *CreateLetterRequest) (*Letter, error) {
	if s.createErr != nil {
		return nil, s.createErr
	}
	l := &Letter{
		ID:          "l1",
		OriginalID:  req.OriginalID,
		Subject:     req.Subject,
		Content:     req.Content,
		SenderEmail: req.SenderEmail,
		SenderName:  req.SenderName,
		ReceivedAt:  time.Now().UTC(),
		Status:      "new",
		Category:    "normal",
	}
	s.letters[l.ID] = l
	return l, nil
}

func (s *dummyService) GetLetter(ctx context.Context, id string) (*LetterFull, error) {
	if s.getErr != nil {
		return nil, s.getErr
	}
	l, ok := s.letters[id]
	if !ok {
		return nil, nil
	}
	return &LetterFull{Letter: l, GeneratedReplies: s.replies[id]}, nil
}

func (s *dummyService) GenerateReplies(ctx context.Context, id, action string) ([]GeneratedReply, error) {
	if s.genErr != nil {
		return nil, s.genErr
	}
	out := []GeneratedReply{
		{ID: "r1", LetterID: id, Content: "reply 1"},
		{ID: "r2", LetterID: id, Content: "reply 2"},
	}
	s.replies[id] = out
	return out, nil
}

func (s *dummyService) SendReply(ctx context.Context, id string, req *SendReplyRequest) (string, error) {
	if s.sendErr != nil {
		return "", s.sendErr
	}
	if l, ok := s.letters[id]; ok {
		l.Status = "replied"
	}
	return "sent", nil
}

func (s *dummyService) StartWorkers(ctx context.Context) {}
func (s *dummyService) StopWorkers()                     {}

func TestHandler_ListLetters(t *testing.T) {
	svc := newDummyService()
	log := &logger.StdLogger{}
	h := NewHandler(svc, log)

	svc.letters["l1"] = &Letter{ID: "l1", Subject: "Subj"}

	req := httptest.NewRequest(http.MethodGet, "/letters", nil)
	w := httptest.NewRecorder()
	h.ListLetters(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestHandler_CreateLetter(t *testing.T) {
	svc := newDummyService()
	log := &logger.StdLogger{}
	h := NewHandler(svc, log)

	body := &CreateLetterRequest{
		OriginalID:  "o1",
		Subject:     "Test",
		Content:     "Hello",
		SenderEmail: "a@b.com",
		SenderName:  "John",
	}
	buf, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/letters", bytes.NewBuffer(buf))
	w := httptest.NewRecorder()

	h.CreateLetter(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d", w.Code)
	}
}

func TestHandler_GetLetter(t *testing.T) {
	svc := newDummyService()
	log := &logger.StdLogger{}
	h := NewHandler(svc, log)

	svc.letters["l1"] = &Letter{ID: "l1", Subject: "Subj"}

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "l1")

	req := httptest.NewRequest(http.MethodGet, "/letters/l1", nil)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	h.GetLetter(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestHandler_GenerateReplies(t *testing.T) {
	svc := newDummyService()
	log := &logger.StdLogger{}
	h := NewHandler(svc, log)

	svc.letters["l1"] = &Letter{ID: "l1"}

	body := &GenerateRequest{Action: "act"}
	buf, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/letters/l1/generate", bytes.NewBuffer(buf))
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "l1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	h.GenerateReplies(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestHandler_SendReply(t *testing.T) {
	svc := newDummyService()
	log := &logger.StdLogger{}
	h := NewHandler(svc, log)

	svc.letters["l1"] = &Letter{ID: "l1"}

	body := &SendReplyRequest{CustomContent: "Hello"}
	buf, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/letters/l1/send", bytes.NewBuffer(buf))
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "l1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	h.SendReply(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}
