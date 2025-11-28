package tickets

import (
	"context"
	"errors"
	"sync"
	"time"

	"banking-ai-assistant/internal/ai"
	"banking-ai-assistant/internal/logger"
	"banking-ai-assistant/pkg/utils"
)

type Notifier interface {
	SendEmail(ctx context.Context, to, subject, body string) error
}

type stubNotifier struct {
	log *logger.StdLogger
}

func NewStubNotifier(log *logger.StdLogger) Notifier {
	return &stubNotifier{log: log}
}

func (n *stubNotifier) SendEmail(ctx context.Context, to, subject, body string) error {
	n.log.Info("stub send email", "to", to, "subject", subject)
	return nil
}

type serviceImpl struct {
	repo     Repository
	aiClient ai.Client
	notifier Notifier
	log      *logger.StdLogger

	workersWg sync.WaitGroup
	stopOnce  sync.Once
	stopCh    chan struct{}
}

func NewService(repo Repository, aiClient ai.Client, notifier Notifier, log *logger.StdLogger) Service {
	return &serviceImpl{
		repo:     repo,
		aiClient: aiClient,
		notifier: notifier,
		log:      log,
		stopCh:   make(chan struct{}),
	}
}

func (s *serviceImpl) ListLetters(ctx context.Context, category, status string) ([]Letter, error) {
	return s.repo.ListLetters(ctx, category, status)
}

func (s *serviceImpl) CreateLetter(ctx context.Context, req *CreateLetterRequest) (*Letter, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	letter := &Letter{
		ID:          utils.NewID(),
		OriginalID:  req.OriginalID,
		Subject:     req.Subject,
		Content:     req.Content,
		SenderEmail: req.SenderEmail,
		SenderName:  req.SenderName,
		ReceivedAt:  time.Now().UTC(),
		Status:      "new",
		Category:    "normal",
	}

	if err := s.repo.SaveLetter(ctx, letter); err != nil {
		return nil, err
	}

	analysis, err := s.aiClient.Analyze(ctx, req.Content)
	if err != nil {
		s.log.Error("ai analyze failed", "err", err)
		return letter, nil
	}
	letter.AIAnalysis = &AIAnalysis{
		Type:     analysis.Type,
		Urgency:  analysis.Urgency,
		Tone:     analysis.Tone,
		Summary:  analysis.Summary,
		Keywords: analysis.Keywords,
	}
	letter.Category = DetectCategory(*letter.AIAnalysis)

	if err := s.repo.SaveLetter(ctx, letter); err != nil {
		s.log.Error("failed to save letter with analysis", "err", err)
	}

	return letter, nil
}

func (s *serviceImpl) GetLetter(ctx context.Context, id string) (*LetterFull, error) {
	letter, err := s.repo.GetLetterByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if letter == nil {
		return nil, errors.New("not found")
	}
	replies, err := s.repo.ListRepliesByLetter(ctx, id)
	if err != nil {
		return nil, err
	}
	return &LetterFull{Letter: letter, GeneratedReplies: replies}, nil
}

func (s *serviceImpl) GenerateReplies(ctx context.Context, id string, action string) ([]GeneratedReply, error) {
	letter, err := s.repo.GetLetterByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if letter == nil {
		return nil, errors.New("letter not found")
	}
	aiResp, err := s.aiClient.Generate(ctx, letter.Content, action)
	if err != nil {
		return nil, err
	}
	var out []GeneratedReply
	for _, opt := range aiResp.Options {
		gr := GeneratedReply{
			ID:        opt.ID,
			LetterID:  id,
			Content:   opt.Content,
			Style:     opt.Style,
			Status:    "draft",
			CreatedAt: time.Now().UTC(),
		}
		if err := s.repo.SaveGeneratedReply(ctx, &gr); err != nil {
			s.log.Error("failed to save generated reply", "err", err)
			continue
		}
		out = append(out, gr)
	}
	return out, nil
}

func (s *serviceImpl) SendReply(ctx context.Context, id string, req *SendReplyRequest) (string, error) {
	if err := req.Validate(); err != nil {
		return "", err
	}
	var finalContent string
	if req.ReplyID != "" {
		replies, err := s.repo.ListRepliesByLetter(ctx, id)
		if err != nil {
			return "", err
		}
		var found *GeneratedReply
		for _, r := range replies {
			if r.ID == req.ReplyID {
				tmp := r
				found = &tmp
				break
			}
		}
		if found == nil {
			return "", errors.New("reply not found")
		}
		finalContent = found.Content
	} else {
		finalContent = req.CustomContent
	}

	letter, err := s.repo.GetLetterByID(ctx, id)
	if err != nil || letter == nil {
		return "", errors.New("letter not found")
	}

	if err := s.notifier.SendEmail(ctx, letter.SenderEmail, "Ответ по вашему запросу", finalContent); err != nil {
		return "", err
	}

	if err := s.repo.MarkLetterReplied(ctx, id); err != nil {
		s.log.Error("failed to mark replied", "err", err)
	}
	return "sent", nil
}

func (s *serviceImpl) StartWorkers(ctx context.Context) {
	s.workersWg.Add(1)
	go func() {
		defer s.workersWg.Done()
		s.log.Info("worker: started analysis worker (stub)")
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				s.log.Info("worker: stopping due to ctx done")
				return
			case <-s.stopCh:
				s.log.Info("worker: stop signal received")
				return
			case <-ticker.C:
				// s.log.Info("worker: tick")
			}
		}
	}()
}

func (s *serviceImpl) StopWorkers() {
	s.stopOnce.Do(func() {
		close(s.stopCh)
	})
	s.workersWg.Wait()
}
