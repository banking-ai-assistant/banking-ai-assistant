package tickets

import (
	"banking-ai-assistant/internal/ai"
	"banking-ai-assistant/internal/logger"
	"context"
	"errors"
	"testing"
	"time"
)

type dummyRepo struct {
	letters map[string]*Letter
	replies map[string][]GeneratedReply
	saveErr error
	listErr error
	getErr  error
	markErr error
}

func newDummyRepo() *dummyRepo {
	return &dummyRepo{
		letters: make(map[string]*Letter),
		replies: make(map[string][]GeneratedReply),
	}
}

func (r *dummyRepo) SaveLetter(ctx context.Context, l *Letter) error {
	if r.saveErr != nil {
		return r.saveErr
	}
	r.letters[l.ID] = l
	return nil
}

func (r *dummyRepo) GetLetterByID(ctx context.Context, id string) (*Letter, error) {
	if r.getErr != nil {
		return nil, r.getErr
	}
	l, ok := r.letters[id]
	if !ok {
		return nil, nil
	}
	return l, nil
}

func (r *dummyRepo) ListLetters(ctx context.Context, category, status string) ([]Letter, error) {
	if r.listErr != nil {
		return nil, r.listErr
	}
	out := []Letter{}
	for _, l := range r.letters {
		out = append(out, *l)
	}
	return out, nil
}

func (r *dummyRepo) SaveGeneratedReply(ctx context.Context, rr *GeneratedReply) error {
	if r.saveErr != nil {
		return r.saveErr
	}
	r.replies[rr.LetterID] = append(r.replies[rr.LetterID], *rr)
	return nil
}

func (r *dummyRepo) ListRepliesByLetter(ctx context.Context, letterID string) ([]GeneratedReply, error) {
	if r.listErr != nil {
		return nil, r.listErr
	}
	return r.replies[letterID], nil
}

func (r *dummyRepo) MarkLetterReplied(ctx context.Context, id string) error {
	if r.markErr != nil {
		return r.markErr
	}
	l, ok := r.letters[id]
	if ok {
		l.Status = "replied"
		now := time.Now().UTC()
		l.RepliedAt = &now
	}
	return nil
}

func (r *dummyRepo) UpdateLetterStatus(ctx context.Context, id string, status string) error {
	l, ok := r.letters[id]
	if !ok {
		return errors.New("not found")
	}
	l.Status = status
	return nil
}

type dummyAI struct{}

func (d *dummyAI) Analyze(ctx context.Context, text string) (*ai.AnalyzeResponse, error) {
	return &ai.AnalyzeResponse{
		Type:     "information_request",
		Urgency:  "medium",
		Tone:     "neutral",
		Summary:  "dummy summary",
		Keywords: []string{"dummy"},
	}, nil
}

func (d *dummyAI) Generate(ctx context.Context, content, action string) (*ai.GenerateResponse, error) {
	return &ai.GenerateResponse{
		Options: []ai.GenerateOption{
			{ID: "1", Content: "dummy reply 1", Style: "formal"},
			{ID: "2", Content: "dummy reply 2", Style: "informal"},
		},
	}, nil
}

type dummyNotifier struct{}

func (n *dummyNotifier) SendEmail(ctx context.Context, to, subject, body string) error {
	return nil
}

func TestCreateLetter(t *testing.T) {
	ctx := context.Background()
	repo := newDummyRepo()
	aiClient := &dummyAI{}
	notifier := &dummyNotifier{}
	log := &logger.StdLogger{}

	svc := NewService(repo, aiClient, notifier, log)

	req := &CreateLetterRequest{
		OriginalID:  "123",
		Subject:     "Hello",
		Content:     "Test content",
		SenderEmail: "a@b.com",
		SenderName:  "John Doe",
	}

	letter, err := svc.CreateLetter(ctx, req)
	if err != nil {
		t.Fatalf("CreateLetter failed: %v", err)
	}

	if letter.AIAnalysis == nil || letter.Category == "" {
		t.Errorf("AIAnalysis or Category not set")
	}

	if _, ok := repo.letters[letter.ID]; !ok {
		t.Errorf("Letter not saved in repo")
	}
}

func TestGenerateReplies(t *testing.T) {
	ctx := context.Background()
	repo := newDummyRepo()
	aiClient := &dummyAI{}
	notifier := &dummyNotifier{}
	log := &logger.StdLogger{}

	svc := NewService(repo, aiClient, notifier, log)

	// Создаём письмо
	letter := &Letter{
		ID:          "l1",
		OriginalID:  "o1",
		Subject:     "Subj",
		Content:     "Content",
		SenderEmail: "a@b.com",
		SenderName:  "John",
		ReceivedAt:  time.Now(),
		Status:      "new",
		Category:    "normal",
	}
	repo.SaveLetter(ctx, letter)

	replies, err := svc.GenerateReplies(ctx, letter.ID, "action")
	if err != nil {
		t.Fatalf("GenerateReplies failed: %v", err)
	}

	if len(replies) != 2 {
		t.Errorf("Expected 2 replies, got %d", len(replies))
	}
}

func TestSendReply(t *testing.T) {
	ctx := context.Background()
	repo := newDummyRepo()
	aiClient := &dummyAI{}
	notifier := &dummyNotifier{}
	log := &logger.StdLogger{}

	svc := NewService(repo, aiClient, notifier, log)

	letter := &Letter{
		ID:          "l1",
		OriginalID:  "o1",
		Subject:     "Subj",
		Content:     "Content",
		SenderEmail: "a@b.com",
		SenderName:  "John",
		ReceivedAt:  time.Now(),
		Status:      "new",
		Category:    "normal",
	}
	repo.SaveLetter(ctx, letter)

	req := &SendReplyRequest{
		CustomContent: "Custom reply",
	}

	status, err := svc.SendReply(ctx, letter.ID, req)
	if err != nil {
		t.Fatalf("SendReply failed: %v", err)
	}

	if status != "sent" {
		t.Errorf("Expected status 'sent', got %s", status)
	}

	if letter.Status != "replied" {
		t.Errorf("Letter status not updated, got %s", letter.Status)
	}
}
