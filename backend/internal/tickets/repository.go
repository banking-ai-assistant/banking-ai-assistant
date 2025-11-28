package tickets

import "context"

type Repository interface {
	SaveLetter(ctx context.Context, l *Letter) error
	ListLetters(ctx context.Context, category, status string) ([]Letter, error)
	GetLetterByID(ctx context.Context, id string) (*Letter, error)
	SaveGeneratedReply(ctx context.Context, r *GeneratedReply) error
	ListRepliesByLetter(ctx context.Context, letterID string) ([]GeneratedReply, error)
	UpdateLetterStatus(ctx context.Context, id, status string) error
	MarkLetterReplied(ctx context.Context, id string) error
}
